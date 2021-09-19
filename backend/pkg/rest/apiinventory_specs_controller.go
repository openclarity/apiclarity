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
	"fmt"
	"net/http"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	"github.com/apiclarity/apiclarity/backend/pkg/database"
)

const defaultTagName = "default-tag"

func (s *Server) GetAPIInventoryAPIIDSpecs(params operations.GetAPIInventoryAPIIDSpecsParams) middleware.Responder {
	apiSpecFromDB, err := database.GetAPISpecs(params.APIID)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Errorf("Failed to get api specs from DB. %v", err)
		return operations.NewGetAPIInventoryAPIIDSpecsDefault(http.StatusInternalServerError)
	}

	log.Debugf("Got GetAPIInventoryAPIIDSpecsParams=%+v, Got apiSpecFromDB=%+v", params, apiSpecFromDB)

	providedSpec, err := createSpecInfo(apiSpecFromDB.ProvidedSpec)
	if err != nil {
		log.Errorf("Failed to create spec info from provided spec. %v", err)
		return operations.NewGetAPIInventoryAPIIDSpecsDefault(http.StatusInternalServerError)
	}
	reconstructedSpec, err := createSpecInfo(apiSpecFromDB.ReconstructedSpec)
	if err != nil {
		log.Errorf("Failed to create spec info from reconstructed spec. %v", err)
		return operations.NewGetAPIInventoryAPIIDSpecsDefault(http.StatusInternalServerError)
	}

	return operations.NewGetAPIInventoryAPIIDSpecsOK().WithPayload(
		&models.OpenAPISpecs{
			ProvidedSpec:      providedSpec,
			ReconstructedSpec: reconstructedSpec,
		})
}

func createSpecInfo(rawSpec string) (*models.SpecInfo, error) {
	if rawSpec == "" {
		return nil, nil
	}
	tags, err := createTagsListFromRawSpec(rawSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to create tags list from raw spec: %v. %v", rawSpec, err)
	}
	return &models.SpecInfo{
		Tags: tags,
	}, nil
}

func createTagsListFromRawSpec(rawSpec string) ([]*models.SpecTag, error) {
	var tagList []*models.SpecTag

	tagListMap := map[string][]*models.MethodAndPath{}

	analyzed, err := loads.Analyzed([]byte(rawSpec), "")
	if err != nil {
		return nil, fmt.Errorf("failed to analyze spec: %v. %v", rawSpec, err)
	}
	analyzedSpec := analyzed.Spec()

	for path, pathItem := range analyzedSpec.Paths.Paths {
		addOperationToTagList(pathItem.Get, models.HTTPMethodGET, path, tagListMap)
		addOperationToTagList(pathItem.Put, models.HTTPMethodPUT, path, tagListMap)
		addOperationToTagList(pathItem.Post, models.HTTPMethodPOST, path, tagListMap)
		addOperationToTagList(pathItem.Patch, models.HTTPMethodPATCH, path, tagListMap)
		addOperationToTagList(pathItem.Options, models.HTTPMethodOPTIONS, path, tagListMap)
		addOperationToTagList(pathItem.Delete, models.HTTPMethodDELETE, path, tagListMap)
		addOperationToTagList(pathItem.Head, models.HTTPMethodHEAD, path, tagListMap)
	}

	for tag, methodAndPaths := range tagListMap {
		tagList = append(tagList, &models.SpecTag{
			Description:       "", // TODO from review?
			MethodAndPathList: methodAndPaths,
			Name:              tag,
		})
	}
	return tagList, nil
}

func addOperationToTagList(operation *spec.Operation, method models.HTTPMethod, path string, tagList map[string][]*models.MethodAndPath) {
	if operation == nil {
		return
	}
	if len(operation.Tags) == 0 {
		tagList[defaultTagName] = append(tagList[defaultTagName], &models.MethodAndPath{
			Method: method,
			Path:   path,
		})
		return
	}
	for _, tag := range operation.Tags {
		tagList[tag] = append(tagList[tag], &models.MethodAndPath{
			Method: method,
			Path:   path,
		})
	}
}
