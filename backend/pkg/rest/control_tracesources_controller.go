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
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api/server/restapi/operations"
	_database "github.com/openclarity/apiclarity/backend/pkg/database"
)

func (s *Server) GetControlTraceSources(params operations.GetControlTraceSourcesParams) middleware.Responder {
	log.Debugf("GetControlTraceSources controller was invoked")

	sources, err := s.dbHandler.TraceSourcesTable().GetTraceSources()
	if err != nil {
		log.Errorf("Failed to Get TraceSources: %v", err)
		return operations.NewGetControlTraceSourcesDefault(http.StatusInternalServerError)
	}

	payload := operations.GetControlTraceSourcesOKBody{
		TraceSources: []*models.TraceSource{},
	}

	for _, dbGw := range sources {
		payload.TraceSources = append(payload.TraceSources, &models.TraceSource{
			ID:          int64(dbGw.ID),
			Name:        &dbGw.Name,
			Type:        (*models.TraceSourceType)(&dbGw.Type),
			Description: dbGw.Description,
		})
	}

	return operations.NewGetControlTraceSourcesOK().WithPayload(&payload)
}

func (s *Server) PostControlTraceSources(params operations.PostControlTraceSourcesParams) middleware.Responder {
	log.Debugf("PostControlTraceSources controller was invoked")

	dbSource := _database.TraceSource{
		Name:        *params.Body.Name,
		Type:        string(*params.Body.Type),
		Description: params.Body.Description,
		ExternalID:  params.Body.ExternalID.String(),
	}
	if params.Body.Token != nil {
		dbSource.Token = params.Body.Token
	}
	if dbSource.ExternalID == "" {
		dbSource.ExternalID = uuid.New().String()
	}
	if err := s.dbHandler.TraceSourcesTable().CreateTraceSource(&dbSource); err != nil {
		log.Errorf("Failed to create new TraceSource: %v", err)
		return operations.NewPostControlTraceSourcesDefault(http.StatusInternalServerError)
	}

	var extID strfmt.UUID
	if err := extID.UnmarshalText([]byte(dbSource.ExternalID)); err != nil {
		log.Errorf("Can't read external ID as UUID: %v", err)
		return operations.NewPostControlTraceSourcesDefault(http.StatusInternalServerError)
	}
	gw := models.TraceSource{
		ID:          int64(dbSource.ID),
		Name:        &dbSource.Name,
		Type:        (*models.TraceSourceType)(&dbSource.Type),
		Description: dbSource.Description,
		Token:       dbSource.Token,
		ExternalID:  extID,
	}
	return operations.NewPostControlTraceSourcesCreated().WithPayload(&gw)
}

func (s *Server) GetControlTraceSourcesTraceSourceID(params operations.GetControlTraceSourcesTraceSourceIDParams) middleware.Responder {
	log.Debugf("GetControlTraceSourcesTraceSourceID controller was invoked")

	dbSource, err := s.dbHandler.TraceSourcesTable().GetTraceSourceFromExternalID(params.TraceSourceID.String())
	if err != nil {
		log.Errorf("Failed to get Trace Source: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return operations.NewGetControlTraceSourcesTraceSourceIDNotFound().WithPayload(&models.APIResponse{Message: err.Error()})
		}

		return operations.NewGetControlTraceSourcesTraceSourceIDDefault(http.StatusInternalServerError)
	}

	gw := models.TraceSource{
		ID:          int64(dbSource.ID),
		Name:        &dbSource.Name,
		Type:        (*models.TraceSourceType)(&dbSource.Type),
		Description: dbSource.Description,
	}
	return operations.NewGetControlTraceSourcesTraceSourceIDOK().WithPayload(&gw)
}

func (s *Server) DeleteControlTraceSourcesTraceSourceID(params operations.DeleteControlTraceSourcesTraceSourceIDParams) middleware.Responder {
	log.Debugf("DeleteControlTraceSourcesTraceSourceID controller was invoked")

	if err := s.dbHandler.TraceSourcesTable().DeleteTraceSourceFromExternalID(params.TraceSourceID.String()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return operations.NewDeleteControlTraceSourcesTraceSourceIDNotFound()
		}
		log.Errorf("Failed to delete Trace Source '%d': %v", params.TraceSourceID, err)
		return operations.NewDeleteControlTraceSourcesTraceSourceIDDefault(http.StatusInternalServerError)
	}
	return operations.NewDeleteControlTraceSourcesTraceSourceIDNoContent()
}

// Check if Trace Source authentication is needed
// If not, returns nil, meaning that there is no Trace Source
// If yes, check that the token is valid and return the corresponding TraceSource ID.
func (s *Server) CheckTraceSourceAuth(token []byte) (uint, error) {
	if !s.needsTraceSourceAuth {
		return 0, nil
	}

	traceSource, err := s.dbHandler.TraceSourcesTable().GetTraceSourceFromToken(token)
	if err != nil {
		return 0, fmt.Errorf("unable to authenticate Trace Source: %v", err)
	}

	return traceSource.ID, nil
}
