// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api/server/restapi/operations"
	_database "github.com/openclarity/apiclarity/backend/pkg/database"
)

func (s *Server) GetControlAPIGateways(params operations.GetControlGatewaysParams) middleware.Responder {
	log.Debugf("GetControlAPIGateways controller was invoked")

	gateways, err := s.dbHandler.TraceSourcesTable().GetTraceSources()
	if err != nil {
		log.Errorf("Failed to Get APIGateways: %v", err)
		return operations.NewGetControlGatewaysDefault(http.StatusInternalServerError)
	}

	payload := operations.GetControlGatewaysOKBody{
		Gateways: []*models.APIGateway{},
	}

	for _, dbGw := range gateways {
		payload.Gateways = append(payload.Gateways, &models.APIGateway{
			ID:          int64(dbGw.ID),
			Name:        &dbGw.Name,
			Type:        (*models.APIGatewayType)(&dbGw.Type),
			Description: dbGw.Description,
		})
	}

	return operations.NewGetControlGatewaysOK().WithPayload(&payload)
}

func (s *Server) PostControlAPIGateways(params operations.PostControlGatewaysParams) middleware.Responder {
	log.Debugf("PostControlAPIGateways controller was invoked")

	dbGw := _database.TraceSource{
		Name:        *params.Body.Name,
		Type:        string(*params.Body.Type),
		Description: params.Body.Description,
	}
	if err := s.dbHandler.TraceSourcesTable().CreateTraceSource(&dbGw); err != nil {
		log.Errorf("Failed to create new APIGateway: %v", err)
		return operations.NewPostControlGatewaysDefault(http.StatusInternalServerError)
	}

	gw := models.APIGateway{
		ID:          int64(dbGw.ID),
		Name:        &dbGw.Name,
		Type:        (*models.APIGatewayType)(&dbGw.Type),
		Description: dbGw.Description,
		Token:       dbGw.Token,
	}
	return operations.NewPostControlGatewaysCreated().WithPayload(&gw)
}

func (s *Server) GetControlAPIGatewaysGatewayID(params operations.GetControlGatewaysGatewayIDParams) middleware.Responder {
	log.Debugf("GetControlAPIGatewaysGatewayID controller was invoked")

	dbGw, err := s.dbHandler.TraceSourcesTable().GetTraceSource(uint(params.GatewayID))
	if err != nil {
		log.Errorf("Failed to get Gateway: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return operations.NewGetControlGatewaysGatewayIDNotFound().WithPayload(&models.APIResponse{Message: err.Error()})
		}

		return operations.NewGetAPIInventoryAPIIDFromHostAndPortDefault(http.StatusInternalServerError)
	}

	gw := models.APIGateway{
		ID:          int64(dbGw.ID),
		Name:        &dbGw.Name,
		Type:        (*models.APIGatewayType)(&dbGw.Type),
		Description: dbGw.Description,
	}
	return operations.NewGetControlGatewaysGatewayIDOK().WithPayload(&gw)
}

func (s *Server) DeleteControlAPIGatewaysGatewayID(params operations.DeleteControlGatewaysGatewayIDParams) middleware.Responder {
	log.Debugf("DeleteControlAPIGatewaysGatewayID controller was invoked")

	if err := s.dbHandler.TraceSourcesTable().DeleteTraceSource(uint(params.GatewayID)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return operations.NewDeleteControlGatewaysGatewayIDNotFound()
		}
		log.Errorf("Failed to delete Gateway '%d': %v", params.GatewayID, err)
		return operations.NewDeleteControlGatewaysGatewayIDDefault(http.StatusInternalServerError)
	}
	return operations.NewDeleteControlGatewaysGatewayIDNoContent()
}
