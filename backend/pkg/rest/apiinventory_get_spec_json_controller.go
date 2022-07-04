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

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api/server/restapi/operations"
	speculatorspec "github.com/openclarity/speculator/pkg/spec"
)

type swaggerType string

const (
	swaggerTypeProvided      swaggerType = "Provided"
	swaggerTypeReconstructed swaggerType = "Reconstructed"
)

func (s *Server) GetAPIReconstructedSwaggerJSON(params operations.GetAPIInventoryAPIIDReconstructedSwaggerJSONParams) middleware.Responder {
	swaggerJSON, err := s.getAPISwaggerJSON(params.APIID, swaggerTypeReconstructed)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetAPIInventoryAPIIDReconstructedSwaggerJSONDefault(http.StatusInternalServerError).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	return operations.NewGetAPIInventoryAPIIDReconstructedSwaggerJSONOK().WithPayload(swaggerJSON)
}

func (s *Server) GetAPIProvidedSwaggerJSON(params operations.GetAPIInventoryAPIIDProvidedSwaggerJSONParams) middleware.Responder {
	swaggerJSON, err := s.getAPISwaggerJSON(params.APIID, swaggerTypeProvided)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetAPIInventoryAPIIDProvidedSwaggerJSONDefault(http.StatusInternalServerError).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	return operations.NewGetAPIInventoryAPIIDProvidedSwaggerJSONOK().WithPayload(swaggerJSON)
}

func (s *Server) getAPISwaggerJSON(apiID uint32, typ swaggerType) (interface{}, error) {
	apiSpecFromDB, err := s.dbHandler.APIInventoryTable().GetAPISpecs(apiID)
	if err != nil {
		return nil, fmt.Errorf("failed to get api specs from DB: %v", err)
	}

	var specToReturn []byte
	switch typ {
	case swaggerTypeProvided:
		specToReturn = []byte(apiSpecFromDB.ProvidedSpec)
	case swaggerTypeReconstructed:
		specToReturn = []byte(apiSpecFromDB.ReconstructedSpec)
	}

	if len(specToReturn) == 0 {
		return nil, fmt.Errorf("%v spec not found", typ)
	}

	specToReturn, err = yaml.YAMLToJSON(specToReturn)
	if err != nil {
		return nil, fmt.Errorf("failed to convert spec into json (%s): %v", specToReturn, err)
	}

	oasVersion, err := speculatorspec.GetJSONSpecVersion(specToReturn)
	if err != nil {
		return nil, fmt.Errorf("failed to get spec version: %v", err)
	}

	// nolint:exhaustive
	switch oasVersion {
	case speculatorspec.OASv2:
		var doc openapi2.T
		if err = json.Unmarshal(specToReturn, &doc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal to v2 spec: %v", err)
		}

		return doc, nil
	default:
		var doc *openapi3.T
		doc, _, err = speculatorspec.LoadAndValidateRawJSONSpec(specToReturn)
		if err != nil {
			return nil, fmt.Errorf("failed to load spec and validate spec: %v", err)
		}
		return doc, nil
	}
}
