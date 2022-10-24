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

package apiclarityexporter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
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
	type rttInfo struct {
		name   string
		traces ptrace.Traces
	}
	tests := []rttInfo{}
	for i, attrs := range spanClientAttributes {
		td := GenerateTracesOneSpan(ptrace.SpanKindClient)
		tests = append(tests, rttInfo{
			name:   fmt.Sprintf("client_span_%d", i),
			traces: TracesOneSpanAddAttributes(td, attrs),
		})
	}
	for i, attrs := range spanServerAttributes {
		td := GenerateTracesOneSpan(ptrace.SpanKindServer)
		tests = append(tests, rttInfo{
			name:   fmt.Sprintf("server_span_%d", i),
			traces: TracesOneSpanAddAttributes(td, attrs),
		})
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
	cfg.HTTPClientSettings.Endpoint = baseURL
	cfg.QueueSettings.Enabled = false
	cfg.RetrySettings.Enabled = false

	var err error
	set := componenttest.NewNopExporterCreateSettings()
	set.Logger, err = zap.NewDevelopment()
	require.NoError(t, err)
	t.Logf("Creating traces exporter with URL: %s", cfg.HTTPClientSettings.Endpoint)
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
