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

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	_database "github.com/apiclarity/apiclarity/backend/pkg/database"
)

func (s *Server) GetAPIEvents(params operations.GetAPIEventsParams) middleware.Responder {
	var events []*models.APIEvent

	apiEventsFromDB, total, err := s.dbHandler.APIEventsTable().GetAPIEventsAndTotal(params)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetAPIEventsDefault(http.StatusInternalServerError).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	log.Debugf("GetAPIEvents controller was invoked. params=%+v, apiEventsFromDB=%+v, total=%+v", params, apiEventsFromDB, total)

	for i := range apiEventsFromDB {
		events = append(events, _database.APIEventFromDB(&apiEventsFromDB[i]))
	}

	return operations.NewGetAPIEventsOK().WithPayload(
		&operations.GetAPIEventsOKBody{
			Items: events,
			Total: &total,
		})
}

func (s *Server) GetAPIEvent(params operations.GetAPIEventsEventIDParams) middleware.Responder {
	apiEventFromDB, err := s.dbHandler.APIEventsTable().GetAPIEvent(params.EventID)
	if err != nil {
		// TODO: need to handle errors (ex. what should we return when record not found - id is missing)
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		// should it be configured in swagger?
		//switch err {
		//case gorm.ErrRecordNotFound:
		//	return operations.NewGetAPIEventsEventIDDefault(404).WithPayload(&models.APIResponse{
		//		Message: "ID is missing",
		//	})
		//}
		// //errors.Is(err, gorm.ErrRecordNotFound)
		return operations.NewGetAPIEventsEventIDDefault(http.StatusInternalServerError).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	log.Debugf("GetAPIEvent controller was invoked. params=%+v, apiEventFromDB=%+v", params, apiEventFromDB)

	return operations.NewGetAPIEventsEventIDOK().WithPayload(_database.APIEventFromDB(apiEventFromDB))
}

func (s *Server) GetAPIEventsEventIDReconstructedSpecDiff(params operations.GetAPIEventsEventIDReconstructedSpecDiffParams) middleware.Responder {
	specDiffFromDB, err := s.dbHandler.APIEventsTable().GetAPIEventReconstructedSpecDiff(params.EventID)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetAPIEventsEventIDReconstructedSpecDiffDefault(http.StatusInternalServerError).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	log.Debugf("GetAPIEventsEventIDReconstructedSpecDiff controller was invoked. params=%+v, specDiffFromDB=%+v", params, specDiffFromDB)

	return operations.NewGetAPIEventsEventIDReconstructedSpecDiffOK().WithPayload(
		&models.APIEventSpecDiff{
			DiffType: &specDiffFromDB.SpecDiffType,
			NewSpec:  &specDiffFromDB.NewReconstructedSpec,
			OldSpec:  &specDiffFromDB.OldReconstructedSpec,
		})
}

func (s *Server) GetAPIEventsEventIDProvidedSpecDiff(params operations.GetAPIEventsEventIDProvidedSpecDiffParams) middleware.Responder {
	log.Debugf("GetAPIEventsEventIDProvidedSpecDiff controller was invoked. params=%+v", params)
	specDiffFromDB, err := s.dbHandler.APIEventsTable().GetAPIEventProvidedSpecDiff(params.EventID)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetAPIEventsEventIDProvidedSpecDiffDefault(http.StatusInternalServerError).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	return operations.NewGetAPIEventsEventIDProvidedSpecDiffOK().WithPayload(
		&models.APIEventSpecDiff{
			DiffType: &specDiffFromDB.SpecDiffType,
			NewSpec:  &specDiffFromDB.NewProvidedSpec,
			OldSpec:  &specDiffFromDB.OldProvidedSpec,
		})
}
