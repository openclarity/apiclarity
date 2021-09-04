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
	"strconv"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	"github.com/apiclarity/apiclarity/backend/pkg/database"
	"github.com/apiclarity/speculator/pkg/speculator"
)

func (s *RESTServer) PutAPIInventoryAPIIDSpecsProvidedSpec(params operations.PutAPIInventoryAPIIDSpecsProvidedSpecParams) middleware.Responder {
	var apiInfo = &database.APIInfo{}
	var jsonSpec []byte
	var err error

	log.Debugf("Got PutAPIInventoryAPIIDSpecsProvidedSpecParams: %+v", params)

	jsonSpec = []byte(params.Body.RawSpec)

	// if spec is a yaml spec, convert it to json format for saving in db
	if isYamlSpec([]byte(params.Body.RawSpec)) {
		jsonSpec, err = yaml.YAMLToJSON([]byte(params.Body.RawSpec))
		if err != nil {
			// The spec was already validated as a valid yaml, so error here is an internal error, not a validation error
			log.Errorf("Failed to convert yaml spec to json: %s. %v", params.Body.RawSpec, err)
			return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(500)
		}
	}
	if err := validateJsonSpec(jsonSpec); err != nil {
		log.Errorf("Spec validation failed. Spec: %s. %v", jsonSpec, err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecBadRequest().WithPayload("Spec validation failed")
	}

	if err = database.PutProvidedAPISpec(params); err != nil{
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Errorf("Failed to put provided API spec. %v", err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(500)
	}


	if err := database.GetAPIInventoryTable().First(&apiInfo, params.APIID).Error; err != nil {
		log.Errorf("Failed to get APIInventory table with api id: %v. %v", params.APIID, err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(500)
	}
	if err := s.speculator.LoadProvidedSpec(speculator.GetSpecKey(apiInfo.Name, strconv.Itoa(int(apiInfo.Port))), jsonSpec); err != nil {
		log.Errorf("Failed to load provided spec. %v", err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(500)
	}

	return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecCreated().WithPayload(
		&models.RawSpec{
			RawSpec: params.Body.RawSpec,
		})
}

func isJsonSpec(rawSpec []byte) bool {
	swagger := spec.Swagger{}

	if err := json.Unmarshal(rawSpec, &swagger); err != nil {
		return false
	}
	return true
}

func isYamlSpec(rawSpec []byte) bool {
	swagger := spec.Swagger{}

	if err := yaml.Unmarshal(rawSpec, &swagger); err != nil {
		return false
	}
	return true
}

func validateRawJsonSpec(rawSpec []byte) error {
	doc, err := loads.Analyzed(rawSpec, "")
	if err != nil {
		return fmt.Errorf("failed to analyze spec: %s. %v", rawSpec, err)
	}
	err = validate.Spec(doc, strfmt.Default)
	if err != nil {
		return fmt.Errorf("spec validation failed. %v", err)
	}
	return nil
}

func validateJsonSpec(rawSpec []byte) error {
	if !isJsonSpec(rawSpec) {
		return fmt.Errorf("not a vaild json spec schema: %s", rawSpec)
	}

	if err := validateRawJsonSpec(rawSpec); err != nil {
		return fmt.Errorf("failed to validate json spec. %v", err)
	}
	return nil
}
