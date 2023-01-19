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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/TykTechnologies/tyk/apidef"
	"github.com/TykTechnologies/tyk/ctx"
	"github.com/TykTechnologies/tyk/log"
	"github.com/TykTechnologies/tyk/user"
	"github.com/go-openapi/strfmt"

	"github.com/openclarity/apiclarity/plugins/api/client/models"
	"github.com/openclarity/apiclarity/plugins/common"
	"github.com/openclarity/apiclarity/plugins/common/apiclarity_client"
)

var logger = log.Get()

var (
	telemetryHost        string
	gatewayNamespace     string
	enableTLS            bool
	traceSamplingEnabled bool
	token                string
	apiclarityClient     *apiclarity_client.Client
	discoveredApis       []string
	lock                 sync.RWMutex
)

//nolint:gochecknoinits
func init() {
	telemetryHost = os.Getenv("APICLARITY_HOST")
	gatewayNamespace = os.Getenv("TYK_GATEWAY_NAMESPACE")
	token = os.Getenv("TRACE_SOURCE_TOKEN")
	enableTLS = false
	if os.Getenv("ENABLE_TLS") == "true" {
		enableTLS = true
		if _, err := os.Stat(common.CACertFile); os.IsNotExist(err) {
			logger.Errorf("path %s does not exists", common.CACertFile)
			return
		}
	}
	if os.Getenv("TRACE_SAMPLING_ENABLED") == "true" {
		traceSamplingEnabled = true
	}
	if apiclarityClient == nil {
		discoveredApis = []string{}
		client, err := apiclarity_client.Create(enableTLS, telemetryHost, token, common.SamplingInterval)
		if err != nil {
			logger.Errorf("failed to create ApiClarity client: %v", err)
			return
		}
		apiclarityClient = client
		if err := apiclarityClient.RefreshHostsToTrace(); err != nil {
			logger.Errorf("failed to get hosts to trace: %v", err)
		}
		apiclarityClient.Start()
	}
}

// Called during post phase.
//
//nolint:deadcode
func PostGetAPIDefinition(_ http.ResponseWriter, r *http.Request) {
	apiDefinition := ctx.GetDefinition(r)
	if apiDefinition == nil {
		apiDefinition = apiDefinitionRetriever(r.Context())
	}

	if apiDefinition == nil {
		logger.Error("failed to get api definition")
		return
	}

	session := ctx.GetSession(r)
	session = setRequestTimeOnSession(session)

	ctx.SetSession(r, session, false, false)

	// set the apiDefinition since we dont get it in the response phase
	ctx.SetDefinition(r, apiDefinition)
}

// Called during response phase.
//
//nolint:deadcode
func ResponseSendTelemetry(_ http.ResponseWriter, res *http.Response, req *http.Request) {
	logger.Info("handling telemetry")

	apiDefinition := ctx.GetDefinition(req)
	if apiDefinition == nil {
		logger.Error("failed to get api definition")
		return
	}
	if traceSamplingEnabled && apiclarityClient != nil {
		host, port := common.GetHostAndPortFromURL(apiDefinition.Proxy.TargetURL, gatewayNamespace)

		err := processNewDiscoveredAPI(host, port)
		if err != nil {
			logger.Errorf("failed to processNewDiscoveredAPI: '%v'", err)
		}

		if !apiclarityClient.ShouldTrace(host, port) {
			logger.Infof("ignored host: %v:%v", host, port)
			return
		}
	}

	telemetry, err := createTelemetry(res, req, apiDefinition)
	if err != nil {
		logger.Errorf("failed to create telemetry: %v", err)
		return
	}

	err = apiclarityClient.PostTelemetry(telemetry)
	if err != nil {
		logger.Errorf("failed to post telemetry: %v", err)
		return
	}

	logger.Infof("telemetry has been sent")
}

func processNewDiscoveredAPI(hostPart, port string) error {
	if len(hostPart) == 0 {
		return fmt.Errorf("host value is not available")
	}

	host := hostPart

	if len(port) > 0 {
		host = fmt.Sprintf("%s:%s", hostPart, port)
	}

	// Check for newDiscoveredApi
	if !common.Contains(discoveredApis, host) {
		appendNewDiscoveredAPI(host)
		hosts := []string{host}
		err := apiclarityClient.PostNewDiscoveredAPIs(hosts)
		if err != nil {
			return fmt.Errorf("failed to send newDiscoveredApi request: %v", err)
		}
		logger.Infof("sent PostNewDiscoveredAPIs with success")
	}
	return nil
}

func appendNewDiscoveredAPI(host string) {
	lock.RLock()
	defer lock.RUnlock()

	discoveredApis = append(discoveredApis, host)
}

func createTelemetry(res *http.Response, req *http.Request, apiDefinition *apidef.APIDefinition) (*models.Telemetry, error) {
	metadata := ctx.GetSession(req).MetaData
	requestTime, ok := metadata[common.RequestTimeContextKey].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to get request time from metadata")
	}

	responseTime := time.Now().UTC().UnixNano() / int64(time.Millisecond)

	host, port := common.GetHostAndPortFromURL(apiDefinition.Proxy.TargetURL, gatewayNamespace)
	// TODO this is assuming internal service. for external services it will be wrong.
	destinationNamespace := common.GetDestinationNamespaceFromHostOrDefault(host, gatewayNamespace)

	reqBody, truncatedBodyReq, err := common.ReadBody(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	// Restore the content to the request body (since we read it)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

	resBody, truncatedBodyRes, err := common.ReadBody(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	// Restore the content to the response body (since we read it)
	res.Body = ioutil.NopCloser(bytes.NewBuffer(resBody))

	pathAndQuery := common.GetPathWithQuery(req.URL)

	telemetry := models.Telemetry{
		DestinationAddress:   ":" + port, // No destination ip for now
		DestinationNamespace: destinationNamespace,
		Request: &models.Request{
			Common: &models.Common{
				TruncatedBody: truncatedBodyReq,
				Body:          strfmt.Base64(reqBody),
				Headers:       common.CreateHeaders(req.Header),
				Version:       req.Proto,
				Time:          requestTime,
			},
			Host:   host,
			Method: req.Method,
			Path:   pathAndQuery,
		},
		RequestID: common.GetRequestIDFromHeadersOrGenerate(req.Header),
		Response: &models.Response{
			Common: &models.Common{
				TruncatedBody: truncatedBodyRes,
				Body:          strfmt.Base64(resBody),
				Headers:       common.CreateHeaders(res.Header),
				Version:       res.Proto,
				Time:          responseTime,
			},
			StatusCode: strconv.Itoa(res.StatusCode),
		},
		Scheme:        req.URL.Scheme,
		SourceAddress: req.RemoteAddr,
	}

	return &telemetry, nil
}

func setRequestTimeOnSession(session *user.SessionState) *user.SessionState {
	requestTime := time.Now().UTC().UnixNano() / int64(time.Millisecond) // UnixMilli supported only from go 1.17
	if session == nil {
		session = &user.SessionState{MetaData: map[string]interface{}{common.RequestTimeContextKey: requestTime}}
	} else if session.MetaData == nil {
		session.MetaData = map[string]interface{}{common.RequestTimeContextKey: requestTime}
	} else {
		session.MetaData[common.RequestTimeContextKey] = requestTime
	}
	return session
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
