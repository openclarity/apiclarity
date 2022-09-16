// Copyright The OpenTelemetry Authors
// Modifications Copyright © 2022 Cisco Systems, Inc. and its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apiclarityexporter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	otelconfig "go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/ptrace"

	servermodels "github.com/openclarity/apiclarity/plugins/api/server/models"
	"github.com/openclarity/apiclarity/plugins/api/server/restapi"
	"github.com/openclarity/apiclarity/plugins/api/server/restapi/operations"
)

func TestInvalidConfig(t *testing.T) {
	config := &Config{
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: "",
		},
	}
	f := NewFactory()
	set := componenttest.NewNopExporterCreateSettings()
	_, err := f.CreateTracesExporter(context.Background(), set, config)
	require.Error(t, err)
}

func TestTraceNoBackend(t *testing.T) {
	exp := startTracesExporter(t, "http://localhost")
	td := generateTracesOneSpan()
	assert.Error(t, exp.ConsumeTraces(context.Background(), td))
}

func TestTraceInvalidUrl(t *testing.T) {
	exp := startTracesExporter(t, "http:/\\//this_is_an/*/invalid_url")
	td := generateTracesOneSpan()
	assert.Error(t, exp.ConsumeTraces(context.Background(), td))
}

func TestTraceRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		traces ptrace.Traces
	}{
		{
			name:   "client_span",
			traces: GenerateTracesClientSpan(),
		},
		{
			name:   "server_span",
			traces: GenerateTracesServerSpan(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := startTracesReceiver(t)
			exp := startTracesExporter(t, url)
			assert.NoError(t, exp.ConsumeTraces(context.Background(), test.traces))
		})
	}
}

func startTracesExporter(t *testing.T, baseURL string) component.TracesExporter {
	t.Logf("Starting traces exporter with URL: %s", baseURL)
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.Endpoint = baseURL
	cfg.QueueSettings.Enabled = false
	cfg.RetrySettings.Enabled = false

	var err error
	set := componenttest.NewNopExporterCreateSettings()
	set.Logger, err = zap.NewDevelopment()
	require.NoError(t, err)
	t.Logf("Creating traces exporter with URL: %s", cfg.Endpoint)
	exp, err := factory.CreateTracesExporter(context.Background(), set, cfg)
	require.NoError(t, err)
	startAndCleanup(t, exp)
	return exp
}

func validateTelemetry(t *testing.T, telemetry *servermodels.Telemetry) error {
	assert.NotEmpty(t, telemetry.Request.Host)
	assert.NotEmpty(t, telemetry.DestinationAddress)
	t.Logf("trace post telemetry: request from %s -> %s", telemetry.SourceAddress, telemetry.DestinationAddress)
	return nil
}

func MockPostTelemetry(t *testing.T, params operations.PostTelemetryParams) middleware.Responder {
	if err := validateTelemetry(t, params.Body); err != nil {
		t.Errorf("error processing telemetry request: %v", err)
		return operations.NewPostTelemetryDefault(http.StatusInternalServerError)
	} else {
		return operations.NewPostTelemetryOK()
	}
}

func startTracesReceiver(t *testing.T) string {
	// create from default API handler for  plugins API
	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	require.NoError(t, err)
	api := operations.NewAPIClarityPluginsTelemetriesAPIAPI(swaggerSpec)
	api.PostTelemetryHandler = operations.PostTelemetryHandlerFunc(func(params operations.PostTelemetryParams) middleware.Responder {
		return MockPostTelemetry(t, params)
	})
	server := restapi.NewServer(api)
	server.ConfigureAPI()
	/*
		origHandler := server.GetHandler()
		server.SetHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("exporter trace: beginning serve request at: %s", r.URL.String())
			if r.Body != nil {
				t.Log("exporter trace: request has a body")
			}
			recorder := httptest.NewRecorder()
			origHandler.ServeHTTP(recorder, r)
			result := recorder.Result()
			body, _ := io.ReadAll(result.Body)
			result.Body.Close()
			t.Logf("exporter trace: finished serving request with code: %d and body: %s", result.StatusCode, body)
		}))
	*/

	ts := httptest.NewServer(server.GetHandler())
	t.Cleanup(func() {
		ts.Close()
	})
	t.Logf("APIClarity telemetry is ready at: %s", ts.URL)
	return ts.URL
}

func startAndCleanup(t *testing.T, cmp component.Component) {
	require.NoError(t, cmp.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, cmp.Shutdown(context.Background()))
	})
}

func TestErrorResponses(t *testing.T) {
	addr := GetAvailableLocalAddress(t, "tcp")
	errMsgPrefix := fmt.Sprintf("error exporting items, request to http://%s/v1/traces responded with HTTP Status Code ", addr)

	tests := []struct {
		name           string
		responseStatus int
		responseBody   *status.Status
		err            error
		isPermErr      bool
		headers        map[string]string
	}{
		{
			name:           "400",
			responseStatus: http.StatusBadRequest,
			responseBody:   status.New(codes.InvalidArgument, "Bad field"),
			isPermErr:      true,
		},
		{
			name:           "404",
			responseStatus: http.StatusNotFound,
			err:            errors.New(errMsgPrefix + "404"),
		},
		{
			name:           "419",
			responseStatus: http.StatusTooManyRequests,
			responseBody:   status.New(codes.InvalidArgument, "Quota exceeded"),
			err: exporterhelper.NewThrottleRetry(
				errors.New(errMsgPrefix+"429, Message=Quota exceeded, Details=[]"),
				time.Duration(0)*time.Second),
		},
		{
			name:           "503",
			responseStatus: http.StatusServiceUnavailable,
			responseBody:   status.New(codes.InvalidArgument, "Server overloaded"),
			err: exporterhelper.NewThrottleRetry(
				errors.New(errMsgPrefix+"503, Message=Server overloaded, Details=[]"),
				time.Duration(0)*time.Second),
		},
		{
			name:           "503-Retry-After",
			responseStatus: http.StatusServiceUnavailable,
			responseBody:   status.New(codes.InvalidArgument, "Server overloaded"),
			headers:        map[string]string{"Retry-After": "30"},
			err: exporterhelper.NewThrottleRetry(
				errors.New(errMsgPrefix+"503, Message=Server overloaded, Details=[]"),
				time.Duration(30)*time.Second),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/traces", func(writer http.ResponseWriter, request *http.Request) {
				for k, v := range test.headers {
					writer.Header().Add(k, v)
				}
				writer.WriteHeader(test.responseStatus)
				if test.responseBody != nil {
					msg, err := proto.Marshal(test.responseBody.Proto())
					require.NoError(t, err)
					_, err = writer.Write(msg)
					require.NoError(t, err)
				}
			})
			srv := http.Server{
				Addr:    addr,
				Handler: mux,
			}
			ln, err := net.Listen("tcp", addr)
			require.NoError(t, err)
			go func() {
				_ = srv.Serve(ln)
			}()

			cfg := &Config{
				ExporterSettings: otelconfig.NewExporterSettings(otelconfig.NewComponentID(typeStr)),
				// Create without QueueSettings and RetrySettings so that ConsumeTraces
				// returns the errors that we want to check immediately.
			}
			set := componenttest.NewNopExporterCreateSettings()
			set.Logger, err = zap.NewDevelopment()
			require.NoError(t, err)
			exp, err := CreateTracesExporter(context.Background(), set, cfg)
			require.NoError(t, err)

			// start the exporter
			err = exp.Start(context.Background(), componenttest.NewNopHost())
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, exp.Shutdown(context.Background()))
			})

			// generate traces
			traces := ptrace.NewTraces()
			err = exp.ConsumeTraces(context.Background(), traces)
			assert.Error(t, err)

			if test.isPermErr {
				assert.True(t, consumererror.IsPermanent(err))
			} else {
				assert.EqualValues(t, test.err, err)
			}

			srv.Close()
		})
	}
}

func TestUserAgent(t *testing.T) {
	var err error
	addr := GetAvailableLocalAddress(t, "tcp")
	set := componenttest.NewNopExporterCreateSettings()
	set.BuildInfo.Description = "Collector"
	set.BuildInfo.Version = "1.2.3test"
	set.Logger, err = zap.NewDevelopment()
	require.NoError(t, err)

	tests := []struct {
		name       string
		headers    map[string]string
		expectedUA string
	}{
		{
			name:       "default_user_agent",
			expectedUA: "Collector/1.2.3test",
		},
		{
			name:       "custom_user_agent",
			headers:    map[string]string{"User-Agent": "My Custom Agent"},
			expectedUA: "My Custom Agent",
		},
		{
			name:       "custom_user_agent_lowercase",
			headers:    map[string]string{"user-agent": "My Custom Agent"},
			expectedUA: "My Custom Agent",
		},
	}

	t.Run("traces", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				mux := http.NewServeMux()
				mux.HandleFunc("/v1/traces", func(writer http.ResponseWriter, request *http.Request) {
					assert.Contains(t, request.Header.Get("user-agent"), test.expectedUA)
					writer.WriteHeader(200)
				})
				srv := http.Server{
					Addr:    addr,
					Handler: mux,
				}
				ln, err := net.Listen("tcp", addr)
				require.NoError(t, err)
				go func() {
					_ = srv.Serve(ln)
				}()

				factory := NewFactory()
				cfg := factory.CreateDefaultConfig().(*Config)
				cfg.HTTPClientSettings = confighttp.HTTPClientSettings{
					Headers:  test.headers,
					Endpoint: fmt.Sprintf("http://%s", addr),
				}

				exp, err := CreateTracesExporter(context.Background(), set, cfg)
				require.NoError(t, err)

				// start the exporter
				err = exp.Start(context.Background(), componenttest.NewNopHost())
				require.NoError(t, err)
				t.Cleanup(func() {
					require.NoError(t, exp.Shutdown(context.Background()))
				})

				// generate data
				traces := ptrace.NewTraces()
				err = exp.ConsumeTraces(context.Background(), traces)
				require.NoError(t, err)

				srv.Close()
			})
		}
	})
}
