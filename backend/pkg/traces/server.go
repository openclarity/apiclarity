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

package traces

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/backend/pkg/common"
	"github.com/openclarity/apiclarity/plugins/api/server/models"
	"github.com/openclarity/apiclarity/plugins/api/server/restapi"
	"github.com/openclarity/apiclarity/plugins/api/server/restapi/operations"
)

type HandleTraceFunc func(ctx context.Context, trace *models.Telemetry) error

type HTTPTracesServer struct {
	traceHandleFunc HandleTraceFunc
	server          *restapi.Server
}

func CreateHTTPTracesServer(port int, traceHandleFunc HandleTraceFunc) (*HTTPTracesServer, error) {
	s := &HTTPTracesServer{}

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger: %v", err)
	}

	api := operations.NewAPIClarityPluginsTelemetriesAPIAPI(swaggerSpec)

	api.PostTelemetryHandler = operations.PostTelemetryHandlerFunc(func(params operations.PostTelemetryParams) middleware.Responder {
		return s.PostTelemetry(params)
	})

	server := restapi.NewServer(api)

	server.ConfigureFlags()
	server.ConfigureAPI()

	server.Port = port
	s.server = server
	s.traceHandleFunc = traceHandleFunc

	return s, nil
}

func (s *HTTPTracesServer) Start(errChan chan struct{}) {
	log.Infof("Starting traces server")

	go func() {
		if err := s.server.Serve(); err != nil {
			log.Errorf("Failed to serve traces server: %v", err)
			errChan <- common.Empty
		}
	}()
}

func (s *HTTPTracesServer) Stop() {
	log.Infof("Stopping traces server")
	if s.server != nil {
		if err := s.server.Shutdown(); err != nil {
			log.Errorf("Failed to shutdown traces server: %v", err)
		}
	}
}

func (s *HTTPTracesServer) PostTelemetry(params operations.PostTelemetryParams) middleware.Responder {
	if err := s.traceHandleFunc(params.HTTPRequest.Context(), params.Body); err != nil {
		log.Errorf("Error from trace handling func: %v", err)
		return operations.NewPostTelemetryDefault(http.StatusInternalServerError)
	}

	return operations.NewPostTelemetryOK()
}
