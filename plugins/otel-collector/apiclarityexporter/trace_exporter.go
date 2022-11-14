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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"

	openapiclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/openclarity/apiclarity/plugins/api/client/client"
	apiclientops "github.com/openclarity/apiclarity/plugins/api/client/client/operations"
	apiclientmodels "github.com/openclarity/apiclarity/plugins/api/client/models"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type exporter struct {
	// Input configuration.
	config   *Config
	logger   *zap.Logger
	settings component.TelemetrySettings
	// Default user-agent header.
	userAgent string
	client    *http.Client
	service   *apiclient.APIClarityPluginsTelemetriesAPI
}

// Create new exporter.
func newExporter(oCfg *Config, set component.ExporterCreateSettings) (*exporter, error) {
	if err := oCfg.Validate(); err != nil {
		set.Logger.Error("configuration error", zap.Error(err))
		return nil, err
	}

	userAgent := fmt.Sprintf("%s/%s (%s/%s)",
		set.BuildInfo.Description, set.BuildInfo.Version, runtime.GOOS, runtime.GOARCH)

	// client construction is deferred to start
	return &exporter{
		config:    oCfg,
		logger:    set.Logger,
		userAgent: userAgent,
		settings:  set.TelemetrySettings,
		service:   nil,
	}, nil
}

// start actually creates the HTTP client. The client construction is deferred till this point as this
// is the only place we get hold of Extensions which are required to construct auth round tripper.
func (e *exporter) start(_ context.Context, host component.Host) error {
	// Add base path specific to endpoint b/c otel endpoint doesn't include path
	urlInfo, err := url.Parse(e.config.HTTPClientSettings.Endpoint + apiclient.DefaultBasePath)
	if err != nil {
		return fmt.Errorf("HTTP endpoint must be a valid URL: %w", err)
	}
	e.client, err = e.config.HTTPClientSettings.ToClient(host, e.settings)
	if err != nil {
		return fmt.Errorf("cannot create HTTP client: %w", err)
	}
	runtime := openapiclient.NewWithClient(urlInfo.Host, urlInfo.Path, []string{urlInfo.Scheme}, e.client)
	e.service = apiclient.New(runtime, strfmt.Default)
	//e.logger.Debug("started client for telemetry", zap.String("url", urlInfo.String()), zap.String("endpoint", e.config.Endpoint))

	return nil
}

// https://pkg.go.dev/go.opentelemetry.io/collector@v0.56.0/consumer#ConsumeTracesFunc
func (e *exporter) pushTraces(ctx context.Context, td ptrace.Traces) error {
	if e.service == nil {
		return errors.New("cannot process traces: client is not initialized")
	}

	rspans := td.ResourceSpans()
	for i := 0; i < rspans.Len(); i++ {
		rspan := rspans.At(i)
		e.logger.Debug("Processing resource span in trace",
			zap.Int("index", i),
		)
		res := rspan.Resource()

		sspans := rspan.ScopeSpans()
		for j := 0; j < sspans.Len(); j++ {
			sspan := sspans.At(j)
			e.logger.Debug("Processing scope span",
				zap.Int("index", j),
				zap.String("scope.name", sspan.Scope().Name()),
				zap.String("scope.version", sspan.Scope().Version()),
			)
			scope := sspan.Scope()

			spans := sspan.Spans()
			e.logger.Debug("Processing spans", zap.Int("total", spans.Len()))
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				e.logger.Debug("Processing span",
					zap.Int("index", j),
				)

				// TODO: process rest of spans on error, then return errors
				// generate permanent errors during processing and check here
				actel, perr := e.processOTelSpan(res, scope, span)
				if perr != nil {
					return consumererror.NewPermanent(perr)
				}

				err := e.export(ctx, actel)
				if err != nil {
					return consumererror.NewPermanent(err)
				}
			}
		}
	}

	return nil
}

func (e *exporter) export(ctx context.Context, actelemetry *apiclientmodels.Telemetry) error {
	//e.logger.Debug("Preparing to make APIClarity telemetry request")

	params := apiclientops.NewPostTelemetryParamsWithContext(ctx).WithBody(actelemetry).WithHTTPClient(e.client)
	_, err := e.service.Operations.PostTelemetry(params)
	if err != nil {
		formattedErr := fmt.Errorf("failed to post telemetry: %w", err)
		e.logger.Error("Failed to post telemetry",
			zap.Error(err),
		)
		return formattedErr
	}

	// All other errors are retryable, so don't wrap them in consumererror.NewPermanent().
	return nil
}
