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
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

	backendmodels "github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/backend/pkg/common"
	"github.com/openclarity/apiclarity/backend/pkg/sampling"
	"github.com/openclarity/apiclarity/plugins/api/server/models"
	"github.com/openclarity/apiclarity/plugins/api/server/restapi"
	"github.com/openclarity/apiclarity/plugins/api/server/restapi/operations"
)

const (
	traceSourceAuthTokenHeaderName = "X-Trace-Source-Token" //nolint:gosec
)

type (
	HandleTraceFunc         func(ctx context.Context, trace *models.Telemetry, traceSource *backendmodels.TraceSource) error
	HandleNewDiscoveredAPIs func(ctx context.Context, hosts []string, traceSource *backendmodels.TraceSource) error
	TraceSourceAuthFunc     func(ctx context.Context, token string) (*backendmodels.TraceSource, error)
)

type HTTPTracesServer struct {
	traceHandleFunc             HandleTraceFunc
	newDiscoveredAPIsHandleFunc HandleNewDiscoveredAPIs
	traceSourceAuthFunc         TraceSourceAuthFunc

	server               *restapi.Server
	TraceSamplingManager *sampling.TraceSamplingManager
}

type HTTPTracesServerConfig struct {
	EnableTLS             bool
	Port                  int
	TLSPort               int
	TLSServerCertFilePath string
	TLSServerKeyFilePath  string
	TraceHandleFunc       HandleTraceFunc
	NewDiscoveredAPIsFunc HandleNewDiscoveredAPIs
	TraceSourceAuthFunc   TraceSourceAuthFunc
	TraceSamplingManager  *sampling.TraceSamplingManager
}

func CreateHTTPTracesServer(config *HTTPTracesServerConfig) (*HTTPTracesServer, error) {
	s := &HTTPTracesServer{}

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger: %v", err)
	}

	api := operations.NewAPIClarityPluginsTelemetriesAPIAPI(swaggerSpec)

	api.PostTelemetryHandler = operations.PostTelemetryHandlerFunc(func(params operations.PostTelemetryParams) middleware.Responder {
		return s.PostTelemetry(params)
	})
	api.GetHostsToTraceHandler = operations.GetHostsToTraceHandlerFunc(func(params operations.GetHostsToTraceParams) middleware.Responder {
		return s.getHostsToTrace(params)
	})
	api.PostControlNewDiscoveredAPIsHandler = operations.PostControlNewDiscoveredAPIsHandlerFunc(func(params operations.PostControlNewDiscoveredAPIsParams) middleware.Responder {
		return s.newDiscoveredAPIs(params)
	})

	if config.TraceSourceAuthFunc != nil {
		s.traceSourceAuthFunc = config.TraceSourceAuthFunc
		api.AddMiddlewareFor("POST", "/telemetry", s.traceSourceAuthMiddleware)
		api.AddMiddlewareFor("GET", "/hostsToTrace", s.traceSourceAuthMiddleware)
		api.AddMiddlewareFor("POST", "/control/newDiscoveredAPIs", s.traceSourceAuthMiddleware)
	}

	server := restapi.NewServer(api)

	server.ConfigureFlags()
	server.ConfigureAPI()
	server.Port = config.Port

	// We want to serve both http and https, except if the authencation is wanted,
	// in that case, we are on the public facing server, and we ONLY want HTTPS.
	// TODO: need to use istio to secure the http port when the wasm is sending traces
	if config.EnableTLS {
		if config.TraceSourceAuthFunc != nil { // If authentication is enabled, only enable https
			server.EnabledListeners = []string{"https"}
		} else {
			server.EnabledListeners = []string{"https", "http"}
		}
		server.TLSCertificate = flags.Filename(config.TLSServerCertFilePath)
		server.TLSCertificateKey = flags.Filename(config.TLSServerKeyFilePath)
		server.TLSPort = config.TLSPort
	}

	s.server = server
	s.traceHandleFunc = config.TraceHandleFunc
	s.newDiscoveredAPIsHandleFunc = config.NewDiscoveredAPIsFunc
	s.TraceSamplingManager = config.TraceSamplingManager

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

const traceSourceKey = "contextTraceSource"

type traceServerContextKey string

func WithTraceSource(ctx context.Context, traceSource *backendmodels.TraceSource) context.Context {
	return context.WithValue(ctx, traceServerContextKey(traceSourceKey), traceSource)
}

func TraceSourceFromContext(ctx context.Context) *backendmodels.TraceSource {
	v := ctx.Value(traceServerContextKey(traceSourceKey))
	if v == nil {
		return nil
	}
	return v.(*backendmodels.TraceSource)
}

func (s *HTTPTracesServer) traceSourceAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(traceSourceAuthTokenHeaderName)
		if token == "" { // Header is not set.
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		traceSource, err := s.traceSourceAuthFunc(r.Context(), token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		r = r.WithContext(WithTraceSource(r.Context(), traceSource))
		next.ServeHTTP(w, r)
	})
}

func (s *HTTPTracesServer) PostTelemetry(params operations.PostTelemetryParams) middleware.Responder {
	traceSource := TraceSourceFromContext(params.HTTPRequest.Context())
	if err := s.traceHandleFunc(params.HTTPRequest.Context(), params.Body, traceSource); err != nil {
		log.Errorf("Error from trace handling func: %v", err)
		return operations.NewPostTelemetryDefault(http.StatusInternalServerError)
	}

	return operations.NewPostTelemetryOK()
}

func (s *HTTPTracesServer) getHostsToTrace(params operations.GetHostsToTraceParams) middleware.Responder {
	if s.TraceSamplingManager == nil {
		log.Errorf("No trace sampling manager configured")
		return operations.NewGetHostsToTraceDefault(http.StatusInternalServerError).
			WithPayload(&models.APIResponse{Message: "No trace sampling manager configured"})
	}

	traceSourceID := common.DefaultTraceSourceID
	traceSource := TraceSourceFromContext(params.HTTPRequest.Context())
	if traceSource != nil {
		traceSourceID = uint(traceSource.ID)
	}

	hosts, err := s.TraceSamplingManager.GetHostsToTrace("*", traceSourceID)
	if err != nil {
		log.Errorf("Error from trace handling func: %v", err)
		return operations.NewGetHostsToTraceDefault(http.StatusInternalServerError)
	}

	return operations.NewGetHostsToTraceOK().WithPayload(
		&models.HostsToTrace{
			Hosts: hosts,
		})
}

func (s *HTTPTracesServer) newDiscoveredAPIs(params operations.PostControlNewDiscoveredAPIsParams) middleware.Responder {
	log.Infof("newDiscoveredAPIs was invoked")
	traceSource := TraceSourceFromContext(params.HTTPRequest.Context())
	hosts := params.Body.Hosts

	if err := s.newDiscoveredAPIsHandleFunc(params.HTTPRequest.Context(), hosts, traceSource); err != nil {
		return operations.NewPostControlNewDiscoveredAPIsDefault(http.StatusInternalServerError).WithPayload(&models.APIResponse{
			Message: "Unable to process all new discovered APIs",
		})
	}

	return operations.NewPostControlNewDiscoveredAPIsOK().WithPayload(&models.APIResponse{
		Message: "New APIs will be processed",
	})
}
