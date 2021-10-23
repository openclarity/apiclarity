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
	"net/http"
	"strconv"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	"github.com/apiclarity/apiclarity/backend/pkg/database"
	"github.com/apiclarity/speculator/pkg/speculator"
)

func (s *Server) DeleteAPIInventoryAPIIDSpecsProvidedSpec(params operations.DeleteAPIInventoryAPIIDSpecsProvidedSpecParams) middleware.Responder {
	apiInfo := &database.APIInfo{}

	if err := s.dbHandler.APIInventoryTable().First(apiInfo, params.APIID); err != nil {
		log.Errorf("Failed to get API info. id=%v: %v", params.APIID, err)
		return operations.NewDeleteAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}

	if err := s.speculator.UnsetProvidedSpec(speculator.GetSpecKey(apiInfo.Name, strconv.Itoa(int(apiInfo.Port)))); err != nil {
		log.Errorf("Failed to unset provided spec. %v", err)
		return operations.NewDeleteAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}
	if err := s.dbHandler.APIInventoryTable().DeleteProvidedAPISpec(params.APIID); err != nil {
		log.Errorf("Failed to delete provided spec with api id: %v from DB. %v", params.APIID, err)
		return operations.NewDeleteAPIInventoryAPIIDSpecsProvidedSpecDefault(http.StatusInternalServerError)
	}

	return operations.NewDeleteAPIInventoryAPIIDSpecsProvidedSpecOK().WithPayload(&models.SuccessResponse{
		Message: "Success",
	})
}

func (s *Server) DeleteAPIInventoryAPIIDSpecsReconstructedSpec(params operations.DeleteAPIInventoryAPIIDSpecsReconstructedSpecParams) middleware.Responder {
	apiInfo := &database.APIInfo{}

	if err := s.dbHandler.APIInventoryTable().First(apiInfo, params.APIID); err != nil {
		log.Errorf("Failed to get API info. id=%v: %v", params.APIID, err)
		return operations.NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault(http.StatusInternalServerError)
	}

	if err := s.speculator.UnsetApprovedSpec(speculator.GetSpecKey(apiInfo.Name, strconv.Itoa(int(apiInfo.Port)))); err != nil {
		log.Errorf("Failed to unset reconstructed spec. %v", err)
		return operations.NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault(http.StatusInternalServerError)
	}

	if err := s.dbHandler.APIInventoryTable().DeleteApprovedAPISpec(params.APIID); err != nil {
		log.Errorf("Failed to delete reconstructed spec with api id: %v from DB. %v", params.APIID, err)
		return operations.NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault(http.StatusInternalServerError)
	}

	return operations.NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecOK().WithPayload(&models.SuccessResponse{
		Message: "Success",
	})
}
