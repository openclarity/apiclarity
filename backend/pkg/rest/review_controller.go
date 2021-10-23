// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	"github.com/apiclarity/apiclarity/backend/pkg/database"
	speculatorspec "github.com/apiclarity/speculator/pkg/spec"
	"github.com/apiclarity/speculator/pkg/speculator"
)

func (s *Server) PostAPIInventoryReviewIDApprovedReview(params operations.PostAPIInventoryReviewIDApprovedReviewParams) middleware.Responder {
	review := database.Review{}
	pathToPathItem := map[string]*spec.PathItem{}

	// find the relevant review
	if err := s.dbHandler.ReviewTable().First(&review, params.ReviewID); err != nil {
		log.Errorf("Failed to find review with id %v in db. %v", params.ReviewID, err)
		return operations.NewPostAPIInventoryReviewIDApprovedReviewDefault(http.StatusInternalServerError)
	}

	// deserialized pathToPathItem that was saved during the suggested review phase
	if err := json.Unmarshal([]byte(review.PathToPathItemStr), &pathToPathItem); err != nil {
		log.Errorf("Failed to unmarshal pathToPathItem: %v. %v", review.PathToPathItemStr, err)
		return operations.NewPostAPIInventoryReviewIDApprovedReviewDefault(http.StatusInternalServerError)
	}

	approvedReview := createApprovedReviewForSpeculator(params.Body, pathToPathItem)
	// apply approved review to the speculator
	if err := s.speculator.ApplyApprovedReview(speculator.SpecKey(review.SpecKey), approvedReview); err != nil {
		errMsg := fmt.Sprintf("Failed to apply the approved review. %v", err)
		log.Error(errMsg)
		return operations.NewPostAPIInventoryReviewIDApprovedReviewDefault(http.StatusInternalServerError).WithPayload(&models.APIResponse{
			Message: errMsg,
		})
	}

	// mark review as approved for later deletion
	if err := s.dbHandler.ReviewTable().UpdateApprovedReview(true, params.ReviewID); err != nil {
		log.Errorf("Failed to update approve in review table. %v", err)
	}

	// generate reconstructed spec and save it to db
	reviewSpec, ok := s.speculator.Specs[speculator.SpecKey(review.SpecKey)]
	if !ok {
		log.Errorf("Failed to find spec with specKey: %v", review.SpecKey)
		return operations.NewPostAPIInventoryReviewIDApprovedReviewDefault(http.StatusInternalServerError)
	}
	oapSpec, err := reviewSpec.GenerateOASJson()
	if err != nil {
		log.Errorf("Failed to generate Open API Spec. %v", err)
		return operations.NewPostAPIInventoryReviewIDApprovedReviewDefault(http.StatusInternalServerError)
	}

	host, port, err := speculator.GetHostAndPortFromSpecKey(speculator.SpecKey(review.SpecKey))
	if err != nil {
		log.Errorf("Failed to parse spec key %v. %v", review.SpecKey, err)
		return operations.NewPostAPIInventoryReviewIDApprovedReviewDefault(http.StatusInternalServerError)
	}

	// TODO: Update PostAPIInventoryReviewIDApprovedReview params to include api ID AND review ID
	apiID, err := s.dbHandler.APIInventoryTable().GetAPIID(host, port)
	if err != nil {
		log.Errorf("Failed to get API ID: %v", err)
		return operations.NewPostAPIInventoryReviewIDApprovedReviewDefault(http.StatusInternalServerError)
	}

	specInfo, err := createSpecInfo(string(oapSpec), getPathToPathIDMap(approvedReview))
	if err != nil {
		log.Errorf("Failed to create spec info: %v", err)
		return operations.NewPostAPIInventoryReviewIDApprovedReviewDefault(http.StatusInternalServerError)
	}

	if err := s.dbHandler.APIInventoryTable().PutAPISpec(apiID, string(oapSpec), specInfo, database.ReconstructedSpecType); err != nil {
		log.Errorf("Failed to save reconstructed API spec to db: %v", err)
		return operations.NewPostAPIInventoryReviewIDApprovedReviewDefault(http.StatusInternalServerError)
	}

	// update all the API events corresponding to the APIEventsPaths in the approved review
	go func() {
		if err := s.dbHandler.APIEventsTable().SetAPIEventsReconstructedPathID(approvedReview.PathItemsReview, host, port); err != nil {
			log.Errorf("Failed to set path ID on API events: %v", err)
		}
	}()

	return operations.NewPostAPIInventoryReviewIDApprovedReviewOK().WithPayload(&models.SuccessResponse{
		Message: "Success",
	})
}

func getPathToPathIDMap(review *speculatorspec.ApprovedSpecReview) map[string]string {
	pathToPathID := make(map[string]string)

	for _, item := range review.PathItemsReview {
		pathToPathID[item.ParameterizedPath] = item.PathUUID
	}

	return pathToPathID
}

func createApprovedReviewForSpeculator(review *models.ApprovedReview, pathToPathItem map[string]*spec.PathItem) *speculatorspec.ApprovedSpecReview {
	ret := &speculatorspec.ApprovedSpecReview{}
	for _, reviewPathItem := range review.ReviewPathItems {
		approvedSpecReviewPathItem := &speculatorspec.ApprovedSpecReviewPathItem{
			ReviewPathItem: speculatorspec.ReviewPathItem{
				ParameterizedPath: reviewPathItem.SuggestedPath,
				Paths:             createPathMap(reviewPathItem.APIEventsPaths),
			},
			PathUUID: uuid.NewV4().String(),
		}
		log.Infof("Approving path item: %+v", approvedSpecReviewPathItem)
		ret.PathItemsReview = append(ret.PathItemsReview, approvedSpecReviewPathItem)
	}
	ret.PathToPathItem = pathToPathItem
	return ret
}

func createPathMap(apiEventPathAndMethods []*models.APIEventPathAndMethods) map[string]bool {
	ret := make(map[string]bool)
	for _, item := range apiEventPathAndMethods {
		ret[item.Path] = true
	}

	return ret
}

func (s *Server) GetAPIInventoryAPIIDSuggestedReview(params operations.GetAPIInventoryAPIIDSuggestedReviewParams) middleware.Responder {
	// get api data from db
	apiInfo := database.APIInfo{}
	if err := s.dbHandler.APIInventoryTable().First(&apiInfo, params.APIID); err != nil {
		log.Errorf("Failed to find api with  id %v in db. %v", params.APIID, err)
		return operations.NewGetAPIInventoryAPIIDSuggestedReviewDefault(http.StatusInternalServerError)
	}

	// get suggested review from the engine using the spec key (host + port)
	specKey := speculator.GetSpecKey(apiInfo.Name, strconv.Itoa(int(apiInfo.Port)))
	suggestedSpecReview, err := s.speculator.SuggestedReview(specKey)
	if err != nil {
		log.Errorf("Failed to create suggested review with spec key: %v. %v", specKey, err)
		return operations.NewGetAPIInventoryAPIIDSuggestedReviewDefault(http.StatusInternalServerError)
	}

	// save pathToPathItem in the database for use when calling approve for that review id
	pathToPathItemB, err := json.Marshal(suggestedSpecReview.PathToPathItem)
	if err != nil {
		log.Errorf("Failed to marshal pathToPathItem map. %v", err)
		return operations.NewGetAPIInventoryAPIIDSuggestedReviewDefault(http.StatusInternalServerError)
	}
	review := &database.Review{
		SpecKey:           string(specKey),
		PathToPathItemStr: string(pathToPathItemB),
		Approved:          false,
	}
	if err := s.dbHandler.ReviewTable().Create(review); err != nil {
		log.Errorf("Failed to create review in database: %v. %v", review, err)
		return operations.NewGetAPIInventoryAPIIDSuggestedReviewDefault(http.StatusInternalServerError)
	}

	// convert suggested review to models review
	reviewPathItems := []*models.ReviewPathItem{}
	for _, reviewPathItem := range suggestedSpecReview.PathItemsReview {
		reviewPathItems = append(reviewPathItems, createModelsReviewPathItem(&reviewPathItem.ReviewPathItem, suggestedSpecReview.PathToPathItem))
	}
	suggestedReview := &models.SuggestedReview{
		ID:              uint32(review.ID),
		ReviewPathItems: reviewPathItems,
	}

	return operations.NewGetAPIInventoryAPIIDSuggestedReviewOK().WithPayload(suggestedReview)
}

func createModelsReviewPathItem(speculatorReviewPathItem *speculatorspec.ReviewPathItem, pathToPathItem map[string]*spec.PathItem) *models.ReviewPathItem {
	var reviewPathItem models.ReviewPathItem
	var apiEventsPaths []*models.APIEventPathAndMethods

	for path := range speculatorReviewPathItem.Paths {
		var methods []models.HTTPMethod
		apiEventPathAndMethods := &models.APIEventPathAndMethods{}
		pathItem, ok := pathToPathItem[path]
		if !ok {
			log.Errorf("Failed to find path: %v in pathToPathItem map", path)
			continue
		}

		if pathItem.Get != nil {
			methods = append(methods, models.HTTPMethodGET)
		}
		if pathItem.Put != nil {
			methods = append(methods, models.HTTPMethodPUT)
		}
		if pathItem.Post != nil {
			methods = append(methods, models.HTTPMethodPOST)
		}
		if pathItem.Patch != nil {
			methods = append(methods, models.HTTPMethodPATCH)
		}
		if pathItem.Head != nil {
			methods = append(methods, models.HTTPMethodHEAD)
		}
		if pathItem.Delete != nil {
			methods = append(methods, models.HTTPMethodDELETE)
		}
		if pathItem.Options != nil {
			methods = append(methods, models.HTTPMethodOPTIONS)
		}

		apiEventPathAndMethods.Path = path
		apiEventPathAndMethods.Methods = methods
		apiEventsPaths = append(apiEventsPaths, apiEventPathAndMethods)
	}

	reviewPathItem.APIEventsPaths = apiEventsPaths
	reviewPathItem.SuggestedPath = speculatorReviewPathItem.ParameterizedPath

	return &reviewPathItem
}
