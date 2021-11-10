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
	"fmt"
	"net/http"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/backend/pkg/common"
	"github.com/apiclarity/apiclarity/plugins/api/server/models"
	"github.com/apiclarity/apiclarity/plugins/api/server/restapi"
	"github.com/apiclarity/apiclarity/plugins/api/server/restapi/operations"
	"github.com/apiclarity/speculator/pkg/spec"
)

type HandleTraceFunc func(trace *spec.SCNTelemetry) error

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
	telemetry := convertTelemetry(params.Body)

	if err := s.traceHandleFunc(telemetry); err != nil {
		log.Errorf("Error from trace handling func: %v", err)
		return operations.NewPostTelemetryDefault(http.StatusInternalServerError)
	}

	return operations.NewPostTelemetryOK()
}

func convertTelemetry(telemetry *models.Telemetry) *spec.SCNTelemetry {
	return &spec.SCNTelemetry{
		RequestID:          telemetry.RequestID,
		Scheme:             telemetry.Scheme,
		DestinationAddress: telemetry.DestinationAddress,
		SourceAddress:      telemetry.SourceAddress,
		SCNTRequest: spec.SCNTRequest{
			Method:     telemetry.Request.Method,
			Path:       telemetry.Request.Path,
			Host:       telemetry.Request.Host,
			SCNTCommon: convertCommon(telemetry.Request.Common),
		},
		SCNTResponse: spec.SCNTResponse{
			StatusCode: telemetry.Response.StatusCode,
			SCNTCommon: convertCommon(telemetry.Response.Common),
		},
	}
}

func convertCommon(common *models.Common) spec.SCNTCommon {
	return spec.SCNTCommon{
		Version:       common.Version,
		Headers:       convertHeaders(common.Headers),
		Body:          common.Body,
		TruncatedBody: common.TruncatedBody,
	}
}

func convertHeaders(headers []*models.Header) [][2]string {
	var ret [][2]string

	for _, header := range headers {
		ret = append(ret, [2]string{
			header.Key, header.Value,
		})
	}
	return ret
}
