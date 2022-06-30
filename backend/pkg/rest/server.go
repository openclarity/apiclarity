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

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api/server/restapi"
	"github.com/openclarity/apiclarity/api/server/restapi/operations"
	"github.com/openclarity/apiclarity/backend/pkg/common"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules"
	_speculator "github.com/openclarity/speculator/pkg/speculator"
	"github.com/openclarity/trace-sampling-manager/manager/pkg/manager"
)

type Server struct {
	server          *restapi.Server
	dbHandler       database.Database
	speculator      *_speculator.Speculator
	modulesWrapper  modules.Module
	samplingManager *manager.Manager
	features        []modules.ModuleInfo
}

func CreateRESTServer(port int, speculator *_speculator.Speculator, dbHandler *database.Handler, modulesWrapper modules.Module, samplingManager *manager.Manager, features []modules.ModuleInfo) (*Server, error) {
	s := &Server{
		speculator:      speculator,
		dbHandler:       dbHandler,
		modulesWrapper:  modulesWrapper,
		samplingManager: samplingManager,
		features:        features,
	}

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger spec: %v", err)
	}

	api := operations.NewAPIClarityAPIsAPI(swaggerSpec)

	api.GetAPIEventsHandler = operations.GetAPIEventsHandlerFunc(func(params operations.GetAPIEventsParams) middleware.Responder {
		return s.GetAPIEvents(params)
	})

	api.GetAPIEventsEventIDHandler = operations.GetAPIEventsEventIDHandlerFunc(func(params operations.GetAPIEventsEventIDParams) middleware.Responder {
		return s.GetAPIEvent(params)
	})

	api.GetAPIEventsEventIDReconstructedSpecDiffHandler = operations.GetAPIEventsEventIDReconstructedSpecDiffHandlerFunc(func(params operations.GetAPIEventsEventIDReconstructedSpecDiffParams) middleware.Responder {
		return s.GetAPIEventsEventIDReconstructedSpecDiff(params)
	})

	api.GetAPIEventsEventIDProvidedSpecDiffHandler = operations.GetAPIEventsEventIDProvidedSpecDiffHandlerFunc(func(params operations.GetAPIEventsEventIDProvidedSpecDiffParams) middleware.Responder {
		return s.GetAPIEventsEventIDProvidedSpecDiff(params)
	})

	api.GetAPIInventoryHandler = operations.GetAPIInventoryHandlerFunc(func(params operations.GetAPIInventoryParams) middleware.Responder {
		return s.GetAPIInventory(params)
	})

	api.PostAPIInventoryHandler = operations.PostAPIInventoryHandlerFunc(func(params operations.PostAPIInventoryParams) middleware.Responder {
		return s.PostAPIInventory(params)
	})

	api.GetAPIInventoryAPIIDSpecsHandler = operations.GetAPIInventoryAPIIDSpecsHandlerFunc(func(params operations.GetAPIInventoryAPIIDSpecsParams) middleware.Responder {
		return s.GetAPIInventoryAPIIDSpecs(params)
	})

	api.GetAPIInventoryAPIIDSpecsHandler = operations.GetAPIInventoryAPIIDSpecsHandlerFunc(func(params operations.GetAPIInventoryAPIIDSpecsParams) middleware.Responder {
		return s.GetAPIInventoryAPIIDSpecs(params)
	})

	api.PutAPIInventoryAPIIDSpecsProvidedSpecHandler = operations.PutAPIInventoryAPIIDSpecsProvidedSpecHandlerFunc(func(params operations.PutAPIInventoryAPIIDSpecsProvidedSpecParams) middleware.Responder {
		return s.PutAPIInventoryAPIIDSpecsProvidedSpec(params)
	})

	api.GetAPIInventoryAPIIDReconstructedSwaggerJSONHandler = operations.GetAPIInventoryAPIIDReconstructedSwaggerJSONHandlerFunc(func(params operations.GetAPIInventoryAPIIDReconstructedSwaggerJSONParams) middleware.Responder {
		return s.GetAPIReconstructedSwaggerJSON(params)
	})

	api.GetAPIInventoryAPIIDProvidedSwaggerJSONHandler = operations.GetAPIInventoryAPIIDProvidedSwaggerJSONHandlerFunc(func(params operations.GetAPIInventoryAPIIDProvidedSwaggerJSONParams) middleware.Responder {
		return s.GetAPIProvidedSwaggerJSON(params)
	})

	api.GetAPIUsageHitCountHandler = operations.GetAPIUsageHitCountHandlerFunc(func(params operations.GetAPIUsageHitCountParams) middleware.Responder {
		return s.GetAPIUsageHitCount(params)
	})

	api.GetDashboardAPIUsageHandler = operations.GetDashboardAPIUsageHandlerFunc(func(params operations.GetDashboardAPIUsageParams) middleware.Responder {
		return s.GetDashboardAPIUsage(params)
	})

	api.GetDashboardAPIUsageMostUsedHandler = operations.GetDashboardAPIUsageMostUsedHandlerFunc(func(params operations.GetDashboardAPIUsageMostUsedParams) middleware.Responder {
		return s.GetDashboardAPIUsageMostUsed(params)
	})

	api.GetDashboardAPIUsageLatestDiffsHandler = operations.GetDashboardAPIUsageLatestDiffsHandlerFunc(func(params operations.GetDashboardAPIUsageLatestDiffsParams) middleware.Responder {
		return s.GetDashboardAPIUsageLatestDiffs(params)
	})

	api.GetAPIInventoryAPIIDSuggestedReviewHandler = operations.GetAPIInventoryAPIIDSuggestedReviewHandlerFunc(func(params operations.GetAPIInventoryAPIIDSuggestedReviewParams) middleware.Responder {
		return s.GetAPIInventoryAPIIDSuggestedReview(params)
	})

	api.GetAPIInventoryAPIIDAPIInfoHandler = operations.GetAPIInventoryAPIIDAPIInfoHandlerFunc(func(params operations.GetAPIInventoryAPIIDAPIInfoParams) middleware.Responder {
		return s.GetAPIInventoryAPIIDAPIInfo(params)
	})

	api.GetAPIInventoryAPIIDFromHostAndPortHandler = operations.GetAPIInventoryAPIIDFromHostAndPortHandlerFunc(func(params operations.GetAPIInventoryAPIIDFromHostAndPortParams) middleware.Responder {
		return s.GetAPIInventoryAPIIDFromHostAndPort(params)
	})

	api.PostAPIInventoryReviewIDApprovedReviewHandler = operations.PostAPIInventoryReviewIDApprovedReviewHandlerFunc(func(params operations.PostAPIInventoryReviewIDApprovedReviewParams) middleware.Responder {
		return s.PostAPIInventoryReviewIDApprovedReview(params)
	})

	api.DeleteAPIInventoryAPIIDSpecsProvidedSpecHandler = operations.DeleteAPIInventoryAPIIDSpecsProvidedSpecHandlerFunc(func(params operations.DeleteAPIInventoryAPIIDSpecsProvidedSpecParams) middleware.Responder {
		return s.DeleteAPIInventoryAPIIDSpecsProvidedSpec(params)
	})

	api.DeleteAPIInventoryAPIIDSpecsReconstructedSpecHandler = operations.DeleteAPIInventoryAPIIDSpecsReconstructedSpecHandlerFunc(func(params operations.DeleteAPIInventoryAPIIDSpecsReconstructedSpecParams) middleware.Responder {
		return s.DeleteAPIInventoryAPIIDSpecsReconstructedSpec(params)
	})

	api.GetFeaturesHandler = operations.GetFeaturesHandlerFunc(func(params operations.GetFeaturesParams) middleware.Responder {
		return s.GetFeatures(params)
	})

	server := restapi.NewServer(api)

	server.ConfigureFlags()
	server.ConfigureAPI()
	server.Port = port

	origHandler := server.GetHandler()
	newHandler := http.NewServeMux()

	// Enhance the default handler with modules apis handlers
	newHandler.Handle("/api/modules/", modulesWrapper.HTTPHandler())
	newHandler.Handle("/", origHandler)
	server.SetHandler(newHandler)
	s.server = server

	return s, nil
}

func (s *Server) Start(errChan chan struct{}) {
	log.Infof("Starting REST server")
	go func() {
		if err := s.server.Serve(); err != nil {
			log.Errorf("Failed to serve REST server: %v", err)
			errChan <- common.Empty
		}
	}()
	log.Infof("REST server is running")
}

func (s *Server) Stop() {
	log.Infof("Stopping REST server")
	if s.server != nil {
		if err := s.server.Shutdown(); err != nil {
			log.Errorf("Failed to shutdown REST server: %v", err)
		}
	}
}
