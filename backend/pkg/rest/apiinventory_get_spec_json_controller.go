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
	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	"github.com/apiclarity/apiclarity/backend/pkg/database"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	log "github.com/sirupsen/logrus"
)

type swaggerType string

const (
	swaggerTypeProvided      swaggerType = "Provided"
	swaggerTypeReconstructed swaggerType = "Reconstructed"
)

func (s *RESTServer) GetAPIReconstructedSwaggerJSON(params operations.GetAPIInventoryAPIIDReconstructedSwaggerJSONParams) middleware.Responder {
	swaggerJSON, err := getAPISwaggerJSON(params.APIID, swaggerTypeReconstructed)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetAPIInventoryAPIIDReconstructedSwaggerJSONDefault(500).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	return operations.NewGetAPIInventoryAPIIDReconstructedSwaggerJSONOK().WithPayload(swaggerJSON)
}

func (s *RESTServer) GetAPIProvidedSwaggerJSON(params operations.GetAPIInventoryAPIIDProvidedSwaggerJSONParams) middleware.Responder {
	swaggerJSON, err := getAPISwaggerJSON(params.APIID, swaggerTypeProvided)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetAPIInventoryAPIIDProvidedSwaggerJSONDefault(500).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	return operations.NewGetAPIInventoryAPIIDProvidedSwaggerJSONOK().WithPayload(swaggerJSON)
}

func getAPISwaggerJSON(apiID uint32, typ swaggerType) (*spec.Swagger, error) {
	apiSpecFromDb, err := database.GetAPISpecs(apiID)
	if err != nil {
		return nil, fmt.Errorf("failed to get api specs from DB: %v", err)
	}

	var specToReturn string
	switch typ {
	case swaggerTypeProvided:
		specToReturn = apiSpecFromDb.ProvidedSpec
	case swaggerTypeReconstructed:
		specToReturn = apiSpecFromDb.ReconstructedSpec
	}

	if specToReturn == "" {
		return nil, fmt.Errorf("%v spec not found", typ)
	}

	analyzed, err := loads.Analyzed([]byte(specToReturn), "")
	if err != nil {
		return nil, fmt.Errorf("failed to analyzed spec: %v", err)
	}

	return analyzed.Spec(), nil
}
