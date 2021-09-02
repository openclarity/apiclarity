/*
 *
 * Copyright (c) 2020 Cisco Systems, Inc. and its affiliates.
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package rest

import (
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	_database "github.com/apiclarity/apiclarity/backend/pkg/database"
)

func (s *RESTServer) GetAPIEvents(params operations.GetAPIEventsParams) middleware.Responder {
	var events []*models.APIEvent

	apiEventsFromDb, total, err := _database.GetAPIEventsAndTotal(params)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetAPIEventsDefault(500).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	log.Debugf("GetAPIEvents controller was invoked. params=%+v, apiEventsFromDb=%+v, total=%+v", params, apiEventsFromDb, total)

	for i := range apiEventsFromDb {
		events = append(events, _database.APIEventFromDB(&apiEventsFromDb[i]))
	}

	return operations.NewGetAPIEventsOK().WithPayload(
		&operations.GetAPIEventsOKBody{
			Items: events,
			Total: &total,
		})
}

func (s *RESTServer) GetAPIEvent(params operations.GetAPIEventsEventIDParams) middleware.Responder {
	apiEventFromDb, err := _database.GetAPIEvent(params.EventID)
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
		return operations.NewGetAPIEventsEventIDDefault(500).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	log.Debugf("GetAPIEvent controller was invoked. params=%+v, apiEventFromDb=%+v", params, apiEventFromDb)

	return operations.NewGetAPIEventsEventIDOK().WithPayload(_database.APIEventFromDB(apiEventFromDb))
}

func (s *RESTServer) GetAPIEventsEventIDReconstructedSpecDiff(params operations.GetAPIEventsEventIDReconstructedSpecDiffParams) middleware.Responder {
	specDiffFromDb, err := _database.GetAPIEventReconstructedSpecDiff(params.EventID)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetAPIEventsEventIDReconstructedSpecDiffDefault(500).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	log.Debugf("GetAPIEventsEventIDReconstructedSpecDiff controller was invoked. params=%+v, specDiffFromDb=%+v", params, specDiffFromDb)

	return operations.NewGetAPIEventsEventIDReconstructedSpecDiffOK().WithPayload(
		&models.APIEventSpecDiff{
			NewSpec: &specDiffFromDb.NewReconstructedSpec,
			OldSpec: &specDiffFromDb.OldReconstructedSpec,
		})
}

func (s *RESTServer) GetAPIEventsEventIDProvidedSpecDiff(params operations.GetAPIEventsEventIDProvidedSpecDiffParams) middleware.Responder {
	log.Debugf("GetAPIEventsEventIDProvidedSpecDiff controller was invoked. params=%+v", params)
	specDiffFromDb, err := _database.GetAPIEventProvidedSpecDiff(params.EventID)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetAPIEventsEventIDProvidedSpecDiffDefault(500).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}


	return operations.NewGetAPIEventsEventIDProvidedSpecDiffOK().WithPayload(
		&models.APIEventSpecDiff{
			NewSpec: &specDiffFromDb.NewProvidedSpec,
			OldSpec: &specDiffFromDb.OldProvidedSpec,
		})
}
