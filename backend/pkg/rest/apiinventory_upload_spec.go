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

	"github.com/ghodss/yaml"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	"github.com/apiclarity/apiclarity/backend/pkg/database"
	"github.com/apiclarity/speculator/pkg/speculator"
)

func (s *Server) PutAPIInventoryAPIIDSpecsProvidedSpec(params operations.PutAPIInventoryAPIIDSpecsProvidedSpecParams) middleware.Responder {
	apiInfo := &database.APIInfo{}

	log.Debugf("Got PutAPIInventoryAPIIDSpecsProvidedSpecParams: %+v", params)

	// Convert YAML to JSON. Since JSON is a subset of YAML, passing JSON through
	// this method should be a no-op.
	jsonSpecBytes, err := yaml.YAMLToJSON([]byte(params.Body.RawSpec))
	if err != nil {
		log.Errorf("Failed to convert yaml spec to json: %s. %v", params.Body.RawSpec, err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}

	// Creates a new analyzed spec document for the provided spec
	analyzed, err := loads.Analyzed(jsonSpecBytes, "")
	if err != nil {
		log.Errorf("failed to analyze spec. Spec: %s. %v", jsonSpecBytes, err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecBadRequest().WithPayload("Spec validation failed")
	}

	// Validates an OpenAPI 2.0 specification document.
	err = validate.Spec(analyzed, strfmt.Default)
	if err != nil {
		log.Errorf("spec validation failed. %v", err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecBadRequest().WithPayload("Spec validation failed")
	}

	// Save the provided spec in the DB without expanding the ref fields
	if err = database.PutProvidedAPISpec(params); err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Errorf("Failed to put provided API spec. %v", err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}

	if err := database.GetAPIInventoryTable().First(&apiInfo, params.APIID).Error; err != nil {
		log.Errorf("Failed to get APIInventory table with api id: %v. %v", params.APIID, err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}

	// Expands the ref fields in the analyzed spec document
	jsonSpecBytes, err = getExpandedSpec(analyzed)
	if err != nil {
		log.Errorf("Failed to get expanded spec: %v", err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}

	if err := s.speculator.LoadProvidedSpec(speculator.GetSpecKey(apiInfo.Name, strconv.Itoa(int(apiInfo.Port))), jsonSpecBytes); err != nil {
		log.Errorf("Failed to load provided spec. %v", err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}

	return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecCreated().WithPayload(
		&models.RawSpec{
			RawSpec: params.Body.RawSpec,
		})
}

// getExpandedSpec expands the ref fields in the analyzed spec document
func getExpandedSpec(analyzed *loads.Document) ([]byte, error) {
	expandedSpec, err := analyzed.Expanded()
	if err != nil {
		return nil, fmt.Errorf("failed to expanded spec. %v", err)
	}

	expandedSpecB, err := json.Marshal(expandedSpec.Spec())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal expanded spec. %v", err)
	}

	return expandedSpecB, nil
}
