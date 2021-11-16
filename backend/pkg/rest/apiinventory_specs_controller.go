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
	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
)

const defaultTagName = "default-tag"

func (s *Server) GetAPIInventoryAPIIDSpecs(params operations.GetAPIInventoryAPIIDSpecsParams) middleware.Responder {
	specsInfo, err := s.dbHandler.APIInventoryTable().GetAPISpecsInfo(params.APIID)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Errorf("Failed to get api specs from DB. %v", err)
		return operations.NewGetAPIInventoryAPIIDSpecsDefault(http.StatusInternalServerError)
	}

	log.Debugf("Got GetAPIInventoryAPIIDSpecsParams=%+v, Got specsInfo=%+v", params, specsInfo)

	return operations.NewGetAPIInventoryAPIIDSpecsOK().WithPayload(specsInfo)
}

func createSpecInfo(rawSpec string, pathToPathID map[string]string) (*models.SpecInfo, error) {
	if rawSpec == "" {
		return nil, nil
	}
	tags, err := createTagsListFromRawSpec(rawSpec, pathToPathID)
	if err != nil {
		return nil, fmt.Errorf("failed to create tags list from raw spec: %v. %v", rawSpec, err)
	}
	return &models.SpecInfo{
		Tags: tags,
	}, nil
}

func createTagsListFromRawSpec(rawSpec string, pathToPathID map[string]string) ([]*models.SpecTag, error) {
	var tagList []*models.SpecTag

	tagListMap := map[string][]*models.MethodAndPath{}

	analyzed, err := loads.Analyzed([]byte(rawSpec), "")
	if err != nil {
		return nil, fmt.Errorf("failed to analyze spec: %v. %v", rawSpec, err)
	}
	analyzedSpec := analyzed.Spec()

	for path, pathItem := range analyzedSpec.Paths.Paths {
		pathID := pathToPathID[path]
		addOperationToTagList(pathItem.Get, models.HTTPMethodGET, path, pathID, tagListMap)
		addOperationToTagList(pathItem.Put, models.HTTPMethodPUT, path, pathID, tagListMap)
		addOperationToTagList(pathItem.Post, models.HTTPMethodPOST, path, pathID, tagListMap)
		addOperationToTagList(pathItem.Patch, models.HTTPMethodPATCH, path, pathID, tagListMap)
		addOperationToTagList(pathItem.Options, models.HTTPMethodOPTIONS, path, pathID, tagListMap)
		addOperationToTagList(pathItem.Delete, models.HTTPMethodDELETE, path, pathID, tagListMap)
		addOperationToTagList(pathItem.Head, models.HTTPMethodHEAD, path, pathID, tagListMap)
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

func addOperationToTagList(operation *spec.Operation, method models.HTTPMethod, path, pathID string, tagList map[string][]*models.MethodAndPath) {
	if operation == nil {
		return
	}

	methodAndPath := &models.MethodAndPath{
		Method: method,
		Path:   path,
		PathID: strfmt.UUID(pathID),
	}

	if len(operation.Tags) == 0 {
		tagList[defaultTagName] = append(tagList[defaultTagName], methodAndPath)
	} else {
		for _, tag := range operation.Tags {
			tagList[tag] = append(tagList[tag], methodAndPath)
		}
	}
}
