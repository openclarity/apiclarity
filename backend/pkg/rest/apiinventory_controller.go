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
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-openapi/runtime/middleware"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api/server/restapi/operations"
	_database "github.com/openclarity/apiclarity/backend/pkg/database"
)

func (s *Server) PostAPIInventory(params operations.PostAPIInventoryParams) middleware.Responder {
	if err := params.Body.APIType.Validate(nil); err != nil {
		return operations.NewPostAPIInventoryDefault(http.StatusBadRequest).WithPayload(&models.APIResponse{
			Message: fmt.Sprintf("apiType invalid: %q", err),
		})
	}
	if params.Body.Name == "" {
		return operations.NewPostAPIInventoryDefault(http.StatusBadRequest).WithPayload(&models.APIResponse{
			Message: "please provide name",
		})
	}
	if params.Body.APIType == models.APITypeINTERNAL && params.Body.DestinationNamespace == "" {
		return operations.NewPostAPIInventoryDefault(http.StatusBadRequest).WithPayload(&models.APIResponse{
			Message: "please provide destinationNamespace for internal apis",
		})
	}
	if params.Body.Port < 1 {
		return operations.NewPostAPIInventoryDefault(http.StatusBadRequest).WithPayload(&models.APIResponse{
			Message: "please provide a valid port",
		})
	}

	uid, _ := uuid.Parse(params.Body.TraceSourceID.String())
	traceSource, err := s.dbHandler.TraceSourcesTable().GetTraceSource(uid)
	if err != nil {
		return operations.NewPostAPIInventoryDefault(http.StatusBadRequest).WithPayload(&models.APIResponse{
			Message: fmt.Sprintf("invalid trace source '%s'", params.Body.TraceSourceID),
		})
	}

	apiInfo := &_database.APIInfo{
		Type:                 params.Body.APIType,
		Name:                 params.Body.Name,
		Port:                 params.Body.Port,
		DestinationNamespace: params.Body.DestinationNamespace,
		TraceSourceID:        traceSource.ID,
	}
	if _, err := s.dbHandler.APIInventoryTable().FirstOrCreate(apiInfo); err != nil {
		log.Error(err)
		return operations.NewPostAPIInventoryDefault(http.StatusInternalServerError).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	_ = s.speculators.Get(apiInfo.TraceSourceID).InitSpec(params.Body.Name, strconv.Itoa(int(params.Body.Port)))

	return operations.NewPostAPIInventoryOK().WithPayload(_database.APIInfoFromDB(apiInfo))
}

func (s *Server) GetAPIInventory(params operations.GetAPIInventoryParams) middleware.Responder {
	var apiInventory []*models.APIInfo

	apiInventoryFromDB, total, err := s.dbHandler.APIInventoryTable().GetAPIInventoryAndTotal(params)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetAPIInventoryDefault(http.StatusInternalServerError).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	log.Debugf("GetAPIInventory controller was invoked. params=%+v, apiInventoryFromDB=%+v, total=%+v", params, apiInventoryFromDB, total)

	for i := range apiInventoryFromDB {
		apiInventory = append(apiInventory, _database.APIInfoFromDB(&apiInventoryFromDB[i]))
	}

	return operations.NewGetAPIInventoryOK().WithPayload(
		&operations.GetAPIInventoryOKBody{
			Items: apiInventory,
			Total: &total,
		})
}

func (s *Server) GetAPIInventoryAPIIDFromHostAndPort(params operations.GetAPIInventoryAPIIDFromHostAndPortParams) middleware.Responder {
	apiID, err := s.dbHandler.APIInventoryTable().GetAPIID(params.Host, params.Port, nil)
	if err != nil {
		log.Errorf("Failed to get API ID: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return operations.NewGetAPIInventoryAPIIDFromHostAndPortNotFound().WithPayload(&models.APIResponse{Message: err.Error()})
		}

		return operations.NewGetAPIInventoryAPIIDFromHostAndPortDefault(http.StatusInternalServerError)
	}

	return operations.NewGetAPIInventoryAPIIDFromHostAndPortOK().WithPayload(uint32(apiID))
}

func (s *Server) GetAPIInventoryAPIIDFromHostAndPortAndTraceSourceID(params operations.GetAPIInventoryAPIIDFromHostAndPortAndTraceSourceIDParams) middleware.Responder {
	uid, _ := uuid.Parse(params.TraceSourceID.String())
	apiID, err := s.dbHandler.APIInventoryTable().GetAPIID(params.Host, params.Port, &uid)
	if err != nil {
		log.Errorf("Failed to get API ID: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return operations.NewGetAPIInventoryAPIIDFromHostAndPortAndTraceSourceIDNotFound().WithPayload(&models.APIResponse{Message: err.Error()})
		}

		return operations.NewGetAPIInventoryAPIIDFromHostAndPortAndTraceSourceIDDefault(http.StatusInternalServerError)
	}

	return operations.NewGetAPIInventoryAPIIDFromHostAndPortAndTraceSourceIDOK().WithPayload(uint32(apiID))
}

func (s *Server) GetAPIInventoryAPIIDAPIInfo(params operations.GetAPIInventoryAPIIDAPIInfoParams) middleware.Responder {
	apiInfo := &_database.APIInfo{}
	if err := s.dbHandler.APIInventoryTable().First(apiInfo, params.APIID); err != nil {
		log.Errorf("Failed to retrieve API info for apiID=%v: %v", params.APIID, err)
		return operations.NewGetAPIInventoryAPIIDAPIInfoDefault(http.StatusInternalServerError)
	}

	return operations.NewGetAPIInventoryAPIIDAPIInfoOK().WithPayload(&models.APIInfoWithType{
		APIInfo: *_database.APIInfoFromDB(apiInfo),
		APIType: apiInfo.Type,
	})
}
