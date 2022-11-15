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
	"strconv"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/runtime/middleware"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api/server/restapi/operations"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	speculatorspec "github.com/openclarity/speculator/pkg/spec"
	"github.com/openclarity/speculator/pkg/speculator"
)

func (s *Server) PutAPIInventoryAPIIDSpecsProvidedSpec(params operations.PutAPIInventoryAPIIDSpecsProvidedSpecParams) middleware.Responder {
	log.Debugf("Got PutAPIInventoryAPIIDSpecsProvidedSpecParams: %+v", params)

	// Convert YAML to JSON. Since JSON is a subset of YAML, passing JSON through
	// this method should be a no-op.
	jsonSpecBytes, err := yaml.YAMLToJSON([]byte(params.Body.RawSpec))
	if err != nil {
		log.Errorf("Failed to convert yaml spec to json: %s. %v", params.Body.RawSpec, err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}

	doc, _, err := speculatorspec.LoadAndValidateRawJSONSpec(jsonSpecBytes)
	if err != nil {
		log.Errorf("failed to analyze spec. Spec: %s. %v", jsonSpecBytes, err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecBadRequest().WithPayload("Spec validation failed")
	}

	// Create a Path to PathID map for each path in the provided spec
	pathToPathID := make(map[string]string)
	for path := range doc.Paths {
		pathToPathID[path] = uuid.NewV4().String()
	}

	// Load provided spec to Speculator
	if err := s.loadProvidedSpec(params.APIID, jsonSpecBytes, pathToPathID); err != nil {
		log.Errorf("Failed to load provided API spec: %v", err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}

	// Since we don't have a mapping between events paths to the parametrized path,
	// We will not set the old events with provided path IDs

	specInfo, err := createSpecInfo(params.Body.RawSpec, pathToPathID)
	if err != nil {
		log.Errorf("Failed to create spec info. %v", err)
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}

	// Save the provided spec in the DB without expanding the ref fields
	if err = s.dbHandler.APIInventoryTable().PutAPISpec(uint(params.APIID), params.Body.RawSpec, specInfo, database.ProvidedSpecType, params.Body.CreatedAt); err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Errorf("Failed to put provided API spec. %v", err)
		if s.unsetProvidedSpec(params.APIID) != nil {
			// We cannot do much more here while trying to gracefully recovery from a store to DB error.
			log.Errorf("Failed to remove provided spec from the system: %v", err)
		}
		return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}

	return operations.NewPutAPIInventoryAPIIDSpecsProvidedSpecCreated().
		WithPayload(&models.RawSpec{RawSpec: params.Body.RawSpec})
}

func (s *Server) loadProvidedSpec(apiID uint32, jsonSpec []byte, pathToPathID map[string]string) error {
	speculator, specKey, err := s.getSpeculatorAndKey(apiID)
	if err != nil {
		return fmt.Errorf("failed to get spec key: %v", err)
	}

	if err := speculator.LoadProvidedSpec(specKey, jsonSpec, pathToPathID); err != nil {
		return fmt.Errorf("failed to load provided spec: %v", err)
	}

	return nil
}

func (s *Server) unsetProvidedSpec(apiID uint32) error {
	speculator, specKey, err := s.getSpeculatorAndKey(apiID)
	if err != nil {
		return fmt.Errorf("failed to get spec key: %v", err)
	}

	if err := speculator.UnsetProvidedSpec(specKey); err != nil {
		return fmt.Errorf("failed to unset provided spec: %w", err)
	}

	return nil
}

func (s *Server) getSpeculatorAndKey(apiID uint32) (*speculator.Speculator, speculator.SpecKey, error) {
	apiInfo := &database.APIInfo{}

	if err := s.dbHandler.APIInventoryTable().First(apiInfo, apiID); err != nil {
		return nil, "", fmt.Errorf("failed to get API Info from DB. id=%v: %v", apiID, err)
	}

	return s.speculators.Get(apiInfo.TraceSourceID), speculator.GetSpecKey(apiInfo.Name, strconv.Itoa(int(apiInfo.Port))), nil
}
