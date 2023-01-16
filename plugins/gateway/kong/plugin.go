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
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"github.com/go-openapi/strfmt"

	"github.com/openclarity/apiclarity/plugins/api/client/models"
	"github.com/openclarity/apiclarity/plugins/common"
	"github.com/openclarity/apiclarity/plugins/common/apiclarity_client"
)

const (
	hostnameSeparator = "."
	MaxBodySize       = 1000 * 1000
)

var (
	apiclarityClient *apiclarity_client.Client
	discoveredApis   []string
	lock             sync.RWMutex
)

type Config struct {
	EnableTLS            bool   `json:"enable_tls"`
	Host                 string `json:"host"`
	TraceSamplingEnabled bool   `json:"trace_sampling_enabled"`
	Token                string `json:"trace_source_token"`
}

func New() interface{} {
	return &Config{}
}

func (conf Config) Access(kong *pdk.PDK) {
	if conf.TraceSamplingEnabled && apiclarityClient == nil {
		_ = kong.Log.Info(fmt.Sprintf("Host: '%v'", conf.Host))
		_ = kong.Log.Info(fmt.Sprintf("TraceSamplingEnabled: '%v'", conf.TraceSamplingEnabled))
		_ = kong.Log.Info(fmt.Sprintf("Token: '%v'", conf.Token))
		_ = kong.Log.Info(fmt.Sprintf("SamplingInterval: '%v'", common.SamplingInterval))
		discoveredApis = []string{}
		if conf.EnableTLS {
			if _, err := os.Stat("/etc/traces/certs"); os.IsNotExist(err) {
				_ = kong.Log.Err("Path '/etc/traces/certs' does not exists")
				return
			}
		}
		_ = kong.Log.Info("Creating ApiClarity client")
		client, err := apiclarity_client.Create(true, conf.Host, conf.Token, common.SamplingInterval)
		if err != nil {
			_ = kong.Log.Err(fmt.Sprintf("Failed to create ApiClarity client: %v", err))
			return
		}
		apiclarityClient = client
		if err := apiclarityClient.RefreshHostsToTrace(); err != nil {
			_ = kong.Log.Err(fmt.Sprintf("Failed to get hosts to trace: %v", err))
		}
		apiclarityClient.Start()
	}

	// set request time on shared context
	if err := kong.Ctx.SetShared(common.RequestTimeContextKey, time.Now().UTC().UnixMilli()); err != nil {
		_ = kong.Log.Err(fmt.Sprintf("Failed to set request time on shared context: %v", err))
	}
}

func (conf Config) Response(kong *pdk.PDK) {
	if apiclarityClient == nil {
		_ = kong.Log.Err(fmt.Sprintf("ApiClarity Client does not exists"))
		return
	}

	// First, manage new discovery APIs
	if err := processNewDiscoveredApi(kong); err != nil {
		_ = kong.Log.Err(fmt.Sprintf("Failed to processNewDiscoveredApi: '%v'", err))
	}

	if conf.TraceSamplingEnabled {
		shouldTrace, err := shouldTrace(kong)
		if err != nil {
			_ = kong.Log.Err(fmt.Sprintf("Failed to get should trace host: %v", err))
		}
		if !shouldTrace {
			return
		}
	}
	telemetry, err := createTelemetry(kong)
	if err != nil {
		_ = kong.Log.Err(fmt.Sprintf("Failed to create telemetry: %v", err))
		return
	}

	err = apiclarityClient.PostTelemetry(telemetry)
	if err != nil {
		_ = kong.Log.Err(fmt.Sprintf("Failed to post telemetry: %v", err))
		return
	}
	_ = kong.Log.Info(fmt.Sprintf("Telemetry has been sent: %v", telemetry))
}

func processNewDiscoveredApi(kong *pdk.PDK) error {
	host, err := getHost(kong)
	if err != nil {
		return fmt.Errorf("failed to get routed service for processNewDiscoveredApi: %v", err)
	}
	// Check for newDiscoveredApi
	if !common.Contains(discoveredApis, host) {
		appendNewDiscoveredApi(host)
		hosts := []string{host}
		if err := apiclarityClient.PostNewDiscoveredAPIs(hosts); err != nil {
			return fmt.Errorf("failed to send newDiscoveredApi request: %v", err)
		}
		_ = kong.Log.Info("Sent PostNewDiscoveredAPIs with success")
	}
	return nil
}

func appendNewDiscoveredApi(host string) {
	lock.RLock()
	defer lock.RUnlock()

	discoveredApis = append(discoveredApis, host)
}

func shouldTrace(kong *pdk.PDK) (bool, error) {
	routedService, err := kong.Router.GetService()
	if err != nil {
		return false, fmt.Errorf("failed to get routed service for shouldTrace: %v", err)
	}
	host, port, _ := parseKongHost(routedService.Host)
	// Note: here 'host' contain already the namespace. i.e. for service 'catalog' on namespace 'sock-shop', host will be 'catalog.sock-shop'
	if apiclarityClient.ShouldTrace(host, port) {
		return true, nil
	}
	_ = kong.Log.Info("Ignoring host: %v:%v", host, port)
	return false, nil
}

func getHost(kong *pdk.PDK) (string, error) {
	routedService, err := kong.Router.GetService()
	if err != nil {
		return "", fmt.Errorf("failed to get routed service: %v", err)
	}
	host, port, _ := parseKongHost(routedService.Host)
	if port != "" {
		host = host + ":" + port
	}
	return host, nil
}

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
		return nil, fmt.Errorf("failed to get routed serivce: %v", err)
	}
	clientIP, err := kong.Client.GetForwardedIp()
	if err != nil {
		_ = kong.Log.Warn(fmt.Sprintf("Failed to get client forwarded ip: %v", err))
	}

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
	host, port, namespace := parseKongHost(routedService.Host)

	// Note by Axel. Currently, we populate Host with (host+namespace), and don't put namespace on DestinationNamespace because
	// on an issue on the way we create a service in ApiClarity either with newDiscoveredApi either with Telemetry
	if len(namespace) > 0 {
		host = host + hostnameSeparator + namespace
	}
	telemetry := models.Telemetry{
		DestinationAddress:   ":" + port, // No destination ip for now
		DestinationNamespace: "",         // No namespace here for now as it is concatenated on 'host'
		Request: &models.Request{
			Common: &models.Common{
				TruncatedBody: truncatedBodyReq,
				Body:          strfmt.Base64(reqBody),
				Headers:       createHeaders(reqHeaders),
				Version:       fmt.Sprintf("%f", version),
				Time:          requestTime,
			},
			Host:   host,
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

// KongHost format: <svc-name>.<namespace>.8000.svc.
func parseKongHost(kongHost string) (host, port, namespace string) {
	sp := strings.Split(kongHost, ".")

	// nolint:gomnd
	if len(sp) < 2 {
		host = kongHost
		return
	}
	host = sp[0] + "." + sp[1]
	namespace = sp[1]

	// nolint:gomnd
	if len(sp) < 3 {
		return
	}

	port = sp[2]

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
