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
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/runtime/security"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/backend/pkg/common"
	"github.com/openclarity/apiclarity/plugins/api/server/models"
	"github.com/openclarity/apiclarity/plugins/api/server/restapi"
	"github.com/openclarity/apiclarity/plugins/api/server/restapi/operations"
)

type (
	HandleTraceFunc     func(ctx context.Context, trace *models.Telemetry, principal *models.TraceSourcePrincipal) error
	TraceSourceAuthFunc func(ctx context.Context, token []byte) (*models.TraceSourcePrincipal, error)
)

type HTTPTracesServer struct {
	traceHandleFunc     HandleTraceFunc
	traceSourceAuthFunc TraceSourceAuthFunc
	server              *restapi.Server
}

type HTTPTracesServerConfig struct {
	EnableTLS             bool
	Port                  int
	TLSPort               int
	TLSServerCertFilePath string
	TLSServerKeyFilePath  string
	TraceHandleFunc       HandleTraceFunc
	TraceSourceAuthFunc   TraceSourceAuthFunc
}

func CreateHTTPTracesServer(config *HTTPTracesServerConfig) (*HTTPTracesServer, error) {
	s := &HTTPTracesServer{}

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger: %v", err)
	}

	api := operations.NewAPIClarityPluginsTelemetriesAPIAPI(swaggerSpec)

	api.TraceSourceTokenHeaderAuth = s.ValidateTraceSource
	if config.TraceHandleFunc == nil {
		// Why do we override the APIKeyAuthenticator ?
		// Because we want an authentication only when served on HTTPs.
		// The default generated code, always checks for the presence of the
		// authentication header. We don't want to check for authentication header
		// while on the HTTP server (because it's not publicly exposed).
		// This authenticator function returns a default empty principal when no
		// authentication header is provided.
		api.APIKeyAuthenticator = APIKeyNoAuth
	}

	api.PostTelemetryHandler = operations.PostTelemetryHandlerFunc(func(params operations.PostTelemetryParams, principal interface{}) middleware.Responder {
		return s.PostTelemetry(params, principal)
	})

	server := restapi.NewServer(api)

	server.ConfigureFlags()
	server.ConfigureAPI()
	server.Port = config.Port

	// We want to serve both http and https, except if the TraceSourceAuthFunc
	// function is set,
	// in that case, we are on the public facing server, and we ONLY want HTTPS.
	// TODO: need to use istio to secure the http port when the wasm is sending traces
	if config.EnableTLS {
		if config.TraceSourceAuthFunc != nil {
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
	s.traceSourceAuthFunc = config.TraceSourceAuthFunc

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

func (s *HTTPTracesServer) PostTelemetry(params operations.PostTelemetryParams, principal interface{}) middleware.Responder {
	traceSourcePrincipal, ok := principal.(*models.TraceSourcePrincipal)
	if !ok {
		panic("Principal is not of type models.TraceSourcePrincipal")
	}
	if err := s.traceHandleFunc(params.HTTPRequest.Context(), params.Body, traceSourcePrincipal); err != nil {
		log.Errorf("Error from trace handling func: %v", err)
		return operations.NewPostTelemetryDefault(http.StatusInternalServerError)
	}

	return operations.NewPostTelemetryOK()
}

func APIKeyNoAuth(name, in string, authenticate security.TokenAuthentication) runtime.Authenticator {
	return security.HttpAuthenticator(func(r *http.Request) (bool, interface{}, error) {
		prin := models.TraceSourcePrincipal("")
		return true, &prin, nil
	})
}

func (s *HTTPTracesServer) ValidateTraceSource(token string) (interface{}, error) {
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, errors.New(http.StatusUnauthorized, "incorrect api key auth (not a valid base64 token)")
	}
	traceSourcePrinc, err := s.traceSourceAuthFunc(context.TODO(), decodedToken)
	if err == nil && traceSourcePrinc != nil {
		return traceSourcePrinc, nil
	}

	return nil, errors.New(http.StatusUnauthorized, "incorrect api key auth")
}
