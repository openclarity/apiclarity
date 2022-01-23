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

package common

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	uuid "github.com/satori/go.uuid"

	"github.com/apiclarity/apiclarity/plugins/api/client/client"
	"github.com/apiclarity/apiclarity/plugins/api/client/models"
)

const (
	MaxBodySize           = 1000 * 1000
	RequestIDHeaderKey    = "X-Request-Id"
	RequestTimeContextKey = "request_time"
)

func ReadBody(body io.ReadCloser) ([]byte, bool, error) {
	ret, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read body: %v", err)
	}
	if len(ret) > MaxBodySize {
		return []byte{}, true, nil
	}
	return ret, false, nil
}

func GetRequestIDFromHeadersOrGenerate(reqHeaders http.Header) string {
	if reqID, ok := reqHeaders[RequestIDHeaderKey]; ok {
		return reqID[0]
	}
	// no request id header, generate request id
	requestID := uuid.NewV4()

	return requestID.String()
}

func CreateHeaders(headers map[string][]string) []*models.Header {
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

func GetPathWithQuery(reqURL *url.URL) string {
	pathAndQuery := reqURL.Path
	if !strings.HasPrefix(pathAndQuery, "/") {
		pathAndQuery = "/" + pathAndQuery
	}
	if reqURL.RawQuery != "" {
		pathAndQuery += "?" + reqURL.RawQuery
	}
	return pathAndQuery
}

func NewAPIClient(host string) *client.APIClarityPluginsTelemetriesAPI {
	cfg := client.DefaultTransportConfig()
	transport := httptransport.New(host, "/api", cfg.Schemes)
	apiClient := client.New(transport, strfmt.Default)
	return apiClient
}

func GetTimeNowRFC3339Nano() (time.Time, error) {
	tNowStr := time.Now().UTC().Format(time.RFC3339Nano)
	tNow, err := time.Parse(time.RFC3339Nano, tNowStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse time: %v", err)
	}
	return tNow, nil
}
