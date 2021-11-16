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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/TykTechnologies/tyk/apidef"
	"github.com/TykTechnologies/tyk/ctx"
	"github.com/TykTechnologies/tyk/log"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/apiclarity/apiclarity/plugins/api/client/client"
	"github.com/apiclarity/apiclarity/plugins/api/client/client/operations"
	"github.com/apiclarity/apiclarity/plugins/api/client/models"
)

const (
	MaxBodySize              = 1000 * 1000
	RequestIDHeaderKey       = "X-Request-Id"
	MinimumSeparatedHostSize = 2
)

var logger = log.Get()

var (
	host             string
	gatewayNamespace string
)

//nolint:gochecknoinits
func init() {
	host = os.Getenv("APICLARITY_HOST")
	gatewayNamespace = os.Getenv("TYK_GATEWAY_NAMESPACE")
}

// Called during post phase for setting the apiDefinition since we dont get it in the response phase.
//nolint:deadcode
func PostGetAPIDefinition(_ http.ResponseWriter, r *http.Request) {
	apiDefinition := ctx.GetDefinition(r)
	if apiDefinition == nil {
		apiDefinition = apiDefinitionRetriever(r.Context())
	}

	if apiDefinition == nil {
		logger.Error("Failed to get api definition")
		return
	}
	ctx.SetDefinition(r, apiDefinition)
}

// Called during response phase.
//nolint:deadcode
func ResponseSendTelemetry(_ http.ResponseWriter, res *http.Response, req *http.Request) {
	logger.Info("Handling telemetry")
	telemetry, err := createTelemetry(res, req)
	if err != nil {
		logger.Errorf("Failed to create telemetry: %v", err)
		return
	}

	apiClient := newAPIClient(host)
	params := operations.NewPostTelemetryParams().WithBody(telemetry)

	_, err = apiClient.Operations.PostTelemetry(params)
	if err != nil {
		logger.Errorf("Failed to post telemetry: %v", err)
		return
	}
	logger.Infof("Telemetry has been sent")
}

func createTelemetry(res *http.Response, req *http.Request) (*models.Telemetry, error) {
	apiDefinition := ctx.GetDefinition(req)
	if apiDefinition == nil {
		return nil, fmt.Errorf("failed to get api definition")
	}

	host, port := getHostAndPortFromTargetURL(apiDefinition.Proxy.TargetURL)
	destinationNamespace := getDestinationNamespaceFromHost(host)

	reqBody, truncatedBodyReq, err := readBody(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	resBody, truncatedBodyRes, err := readBody(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	pathAndQuery := getPathWithQuery(req.URL)

	telemetry := models.Telemetry{
		DestinationAddress:   ":" + port, // No destination ip for now
		DestinationNamespace: destinationNamespace,
		Request: &models.Request{
			Common: &models.Common{
				TruncatedBody: truncatedBodyReq,
				Body:          strfmt.Base64(reqBody),
				Headers:       createHeaders(req.Header),
				Version:       req.Proto,
			},
			Host:   host,
			Method: req.Method,
			Path:   pathAndQuery,
		},
		RequestID: getRequestIDFromHeaders(req.Header),
		Response: &models.Response{
			Common: &models.Common{
				TruncatedBody: truncatedBodyRes,
				Body:          strfmt.Base64(resBody),
				Headers:       createHeaders(res.Header),
				Version:       res.Proto,
			},
			StatusCode: strconv.Itoa(res.StatusCode),
		},
		Scheme:        req.URL.Scheme,
		SourceAddress: req.RemoteAddr,
	}

	return &telemetry, nil
}

func readBody(body io.ReadCloser) ([]byte, bool, error) {
	ret, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read body: %v", err)
	}
	if len(ret) > MaxBodySize {
		return []byte{}, true, nil
	}
	return ret, false, nil
}

func getRequestIDFromHeaders(reqHeaders http.Header) string {
	if reqID, ok := reqHeaders[RequestIDHeaderKey]; ok {
		return reqID[0]
	}
	return ""
}

func getPathWithQuery(reqURL *url.URL) string {
	pathAndQuery := reqURL.Path
	if reqURL.RawQuery != "" {
		pathAndQuery += "?" + reqURL.RawQuery
	}
	return pathAndQuery
}

// Will try to extract the namespace from the host name, and if not found, will use the namespace that the gateway is running in.
func getDestinationNamespaceFromHost(host string) string {
	sp := strings.Split(host, ".")
	if len(sp) >= MinimumSeparatedHostSize {
		return sp[0]
	}
	return gatewayNamespace
}

func getHostAndPortFromTargetURL(targetURL string) (host, port string) {
	if !strings.Contains(targetURL, "://") {
		// need to add scheme to host in order for url.Parse to parse properly
		targetURL = "http://" + targetURL
	}

	parsedHost, err := url.Parse(targetURL)
	if err != nil {
		return targetURL, ""
	}

	host = parsedHost.Hostname()
	port = parsedHost.Port()
	if port == "" {
		if parsedHost.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

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

func newAPIClient(host string) *client.APIClarityPluginsTelemetriesAPI {
	cfg := client.DefaultTransportConfig()
	transport := httptransport.New(host, "/api", cfg.Schemes)
	apiClient := client.New(transport, strfmt.Default)
	return apiClient
}

// This is a hack. Currently there is an open bug in Tyk that the APIDefinition is nil
// https://github.com/TykTechnologies/tyk/issues/3612
// It does not work in the response phase, so need to propagate this information from a previous phase.
func apiDefinitionRetriever(currentCtx interface{}) *apidef.APIDefinition {
	contextValues := reflect.ValueOf(currentCtx).Elem()
	contextKeys := reflect.TypeOf(currentCtx).Elem()

	if contextKeys.Kind() == reflect.Struct {
		for i := 0; i < contextValues.NumField(); i++ {
			rv := contextValues.Field(i)
			reflectValue := reflect.NewAt(rv.Type(), unsafe.Pointer(rv.UnsafeAddr())).Elem().Interface()

			reflectField := contextKeys.Field(i)

			if reflectField.Name == "Context" {
				apiDefinitionRetriever(reflectValue)
			} else if fmt.Sprintf("%T", reflectValue) == "*apidef.APIDefinition" {
				apidefinition := apidef.APIDefinition{}
				b, _ := json.Marshal(reflectValue)
				e := json.Unmarshal(b, &apidefinition)
				if e == nil {
					return &apidefinition
				}
			}
		}
	}

	return nil
}

func main() {}
