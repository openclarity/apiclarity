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
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/apiclarity/apiclarity/api/server/restapi"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	"github.com/apiclarity/apiclarity/backend/pkg/common"
	_speculator "github.com/apiclarity/speculator/pkg/speculator"
)

type RESTServer struct {
	server *restapi.Server
	speculator *_speculator.Speculator
}

func CreateRESTServer(port int, speculator *_speculator.Speculator) (*RESTServer, error) {
	s := &RESTServer{
		speculator: speculator,
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

	api.GetAPIInventoryAPIIDSpecsHandler = operations.GetAPIInventoryAPIIDSpecsHandlerFunc(func(params operations.GetAPIInventoryAPIIDSpecsParams) middleware.Responder  {
		return s.GetAPIInventoryAPIIDSpecs(params)
	})

	api.GetAPIInventoryAPIIDSpecsHandler = operations.GetAPIInventoryAPIIDSpecsHandlerFunc(func(params operations.GetAPIInventoryAPIIDSpecsParams) middleware.Responder  {
		return s.GetAPIInventoryAPIIDSpecs(params)
	})

	api.PutAPIInventoryAPIIDSpecsProvidedSpecHandler = operations.PutAPIInventoryAPIIDSpecsProvidedSpecHandlerFunc(func(params operations.PutAPIInventoryAPIIDSpecsProvidedSpecParams) middleware.Responder  {
		return s.PutAPIInventoryAPIIDSpecsProvidedSpec(params)
	})

	api.GetAPIInventoryAPIIDReconstructedSwaggerJSONHandler = operations.GetAPIInventoryAPIIDReconstructedSwaggerJSONHandlerFunc(func(params operations.GetAPIInventoryAPIIDReconstructedSwaggerJSONParams) middleware.Responder {
		return s.GetAPIReconstructedSwaggerJSON(params)
	})

	api.GetAPIInventoryAPIIDProvidedSwaggerJSONHandler = operations.GetAPIInventoryAPIIDProvidedSwaggerJSONHandlerFunc(func(params operations.GetAPIInventoryAPIIDProvidedSwaggerJSONParams) middleware.Responder {
		return s.GetAPIProvidedSwaggerJSON(params)
	})

	api.GetAPIUsageHitCountHandler = operations.GetAPIUsageHitCountHandlerFunc(func(params operations.GetAPIUsageHitCountParams) middleware.Responder  {
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

	api.PostAPIInventoryReviewIDApprovedReviewHandler = operations.PostAPIInventoryReviewIDApprovedReviewHandlerFunc(func(params operations.PostAPIInventoryReviewIDApprovedReviewParams) middleware.Responder {
		return s.PostAPIInventoryReviewIDApprovedReview(params)
	})

	server := restapi.NewServer(api)

	server.ConfigureFlags()
	server.ConfigureAPI()
	server.Port = port

	s.server = server

	return s, nil
}

func (s *RESTServer) Start(errChan chan struct{}) {
	log.Infof("Starting REST server")
	go func() {
		if err := s.server.Serve(); err != nil {
			log.Errorf("Failed to serve REST server: %v", err)
			errChan <- common.Empty
		}
	}()
	log.Infof("REST server is running")
}

func (s *RESTServer) Stop() {
	log.Infof("Stopping REST server")
	if s.server != nil {
		if err := s.server.Shutdown(); err != nil {
			log.Errorf("Failed to shutdown REST server: %v", err)
		}
	}
}
