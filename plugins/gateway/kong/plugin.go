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

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"github.com/go-openapi/strfmt"

	"github.com/apiclarity/apiclarity/plugins/api/client/client"
	"github.com/apiclarity/apiclarity/plugins/api/client/client/operations"
	"github.com/apiclarity/apiclarity/plugins/api/client/models"
	"github.com/apiclarity/apiclarity/plugins/common"
)

type ServiceToTrace struct {
	Service string `json:"service"`
	Port    int    `json:"port"`
}

type Config struct {
	EnableTLS       bool             `json:"enable_tls"`
	Host            string           `json:"host"`
	ServicesToTrace []ServiceToTrace `json:"services_to_trace"`
	apiClient       *client.APIClarityPluginsTelemetriesAPI
}

func New() interface{} {
	return &Config{}
}

func (conf Config) Access(kong *pdk.PDK) {
	if err := kong.Ctx.SetShared(common.RequestTimeContextKey, time.Now().UTC().UnixMilli()); err != nil {
		_ = kong.Log.Err(fmt.Sprintf("Failed to set request time on shared context: %v", err))
	}
}

func (conf Config) Response(kong *pdk.PDK) {
	_ = kong.Log.Info("Handling telemetry")
	if !conf.dumpTelemetry(kong) {
		_ = kong.Log.Debug("Not dumping telemetry because tracing is not enabled for this service")
		return
	}
	if conf.apiClient == nil {
		var tlsOptions *common.ClientTLSOptions
		if conf.EnableTLS {
			tlsOptions = &common.ClientTLSOptions{
				RootCAFileName: common.CACertFile,
			}
		}
		apiClient, err := common.NewAPIClient(conf.Host, tlsOptions)
		if err != nil {
			_ = kong.Log.Err(fmt.Sprintf("Failed to create new api client: %v", err))
			return
		}
		conf.apiClient = apiClient
	}
	telemetry, err := createTelemetry(kong)
	if err != nil {
		_ = kong.Log.Err(fmt.Sprintf("Failed to create telemetry: %v", err))
		return
	}

	params := operations.NewPostTelemetryParams().WithBody(telemetry)

	_, err = conf.apiClient.Operations.PostTelemetry(params)
	if err != nil {
		_ = kong.Log.Err(fmt.Sprintf("Failed to post telemetry: %v", err))
		return
	}
	_ = kong.Log.Info(fmt.Sprintf("Telemetry has been sent: %v", telemetry))
}

// dumpTelemetry returns true if telemetry must be sent.
// if ServicesToTrace is an empty list, nothing is sent.
// if ServicesToTrace contains {"*", 0} everything is sent.
// if ServicesToTrace contains {"example.com", 1234}, only traces to services "example.com" and port 1234 are sent.
func (conf Config) dumpTelemetry(kong *pdk.PDK) bool {
	traceAllServices := ServiceToTrace{
		Service: "*",
		Port:    0,
	}

	if len(conf.ServicesToTrace) == 0 {
		return false
	}

	routedService, err := kong.Router.GetService()
	if err != nil {
		return false
	}

	parsedHost, _ := parseHost(routedService.Host)
	targetService := ServiceToTrace{
		Service: strings.ToLower(parsedHost),
		Port:    routedService.Port,
	}

	for _, s := range conf.ServicesToTrace {
		if s == traceAllServices || s == targetService {
			return true
		}
	}

	return false
}

const MaxBodySize = 1000 * 1000

func createTelemetry(kong *pdk.PDK) (*models.Telemetry, error) {
	truncatedBodyReq := false
	truncatedBodyRes := false

	requestTime, err := getRequestTimeFromContext(kong)
	if err != nil {
		return nil, fmt.Errorf("failed to get request time from context: %v", err)
	}
	responseTime := time.Now().UTC().UnixMilli()

	routedService, err := kong.Router.GetService()
	if err != nil {
		return nil, fmt.Errorf("failed to get routed service: %v", err)
	}
	clientIP, err := kong.Client.GetForwardedIp()
	if err != nil {
		_ = kong.Log.Warn(fmt.Sprintf("Failed to get client forwarded ip: %v", err))
	}

	destPort := routedService.Port
	host := routedService.Host

	// Will get the actual path that the request was sent to, not the routed one
	path, err := kong.Request.GetPathWithQuery()
	if err != nil {
		return nil, fmt.Errorf("failed to get request path: %v", err)
	}
	reqBody, err := kong.Request.GetRawBody()
	if err != nil {
		return nil, fmt.Errorf("failed to get request body: %v", err)
	}
	if len(reqBody) > MaxBodySize {
		_ = kong.Log.Info("Request body is too long, ignoring")
		reqBody = ""
		truncatedBodyReq = true
	}
	resBody, err := kong.ServiceResponse.GetRawBody()
	if err != nil {
		return nil, fmt.Errorf("failed to get response body: %v", err)
	}
	if len(resBody) > MaxBodySize {
		_ = kong.Log.Info("Response body is too long, ignoring")
		resBody = ""
		truncatedBodyRes = true
	}
	method, err := kong.Request.GetMethod()
	if err != nil {
		return nil, fmt.Errorf("failed to get request method: %v", err)
	}

	statusCode, err := kong.ServiceResponse.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get response status code: %v", err)
	}
	scheme, err := kong.Request.GetScheme()
	if err != nil {
		return nil, fmt.Errorf("failed to get reuqest scheme: %v", err)
	}
	version, err := kong.Request.GetHttpVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get request http version: %v", err)
	}
	reqHeaders, err := kong.Request.GetHeaders(-1) // default limit of 100 headers
	if err != nil {
		return nil, fmt.Errorf("failed to get request headers: %v", err)
	}
	resHeaders, err := kong.Response.GetHeaders(-1) // default limit of 100 headers
	if err != nil {
		return nil, fmt.Errorf("failed to get response headers: %v", err)
	}
	parsedHost, namespace := parseHost(host)

	telemetry := models.Telemetry{
		DestinationAddress:   ":" + strconv.Itoa(destPort), // No destination ip for now
		DestinationNamespace: namespace,
		Request: &models.Request{
			Common: &models.Common{
				TruncatedBody: truncatedBodyReq,
				Body:          strfmt.Base64(reqBody),
				Headers:       createHeaders(reqHeaders),
				Version:       fmt.Sprintf("%f", version),
				Time:          requestTime,
			},
			Host:   parsedHost,
			Method: method,
			Path:   path,
		},
		RequestID: common.GetRequestIDFromHeadersOrGenerate(reqHeaders),
		Response: &models.Response{
			Common: &models.Common{
				TruncatedBody: truncatedBodyRes,
				Body:          strfmt.Base64(resBody),
				Headers:       createHeaders(resHeaders),
				Version:       fmt.Sprintf("%f", version),
				Time:          responseTime,
			},
			StatusCode: strconv.Itoa(statusCode),
		},
		Scheme:        scheme,
		SourceAddress: clientIP + ":",
	}

	return &telemetry, nil
}

func getRequestTimeFromContext(kong *pdk.PDK) (int64, error) {
	requestTime, err := kong.Ctx.GetSharedInt(common.RequestTimeContextKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get request time from shared context: %v", err)
	}

	return int64(requestTime), nil
}

// KongHost: <svc-name>.<namespace>.8000.svc
// convert to name.namespace.
func parseHost(kongHost string) (host, namespace string) {
	sp := strings.Split(kongHost, ".")

	// nolint:gomnd
	if len(sp) < 2 {
		return kongHost, ""
	}
	host = sp[0] + "." + sp[1]
	namespace = sp[1]

	return
}

func createHeaders(headers map[string][]string) []*models.Header {
	ret := []*models.Header{}

	// TODO support multiple values for a header
	for header, values := range headers {
		ret = append(ret, &models.Header{
			Key:   header,
			Value: values[0],
		})
	}
	return ret
}

var (
	Version  = "0.2"
	Priority = 1
)

func main() {
	_ = server.StartServer(New, Version, Priority)
}
