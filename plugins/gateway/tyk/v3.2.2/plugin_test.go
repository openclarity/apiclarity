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
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/TykTechnologies/tyk/apidef"
	"github.com/TykTechnologies/tyk/ctx"
	"github.com/TykTechnologies/tyk/user"

	"github.com/openclarity/apiclarity/plugins/api/client/models"
	"github.com/openclarity/apiclarity/plugins/common"
)

func Test_createTelemetry(t *testing.T) {
	tNow := time.Now().UTC().UnixNano() / int64(time.Millisecond)

	apiDefinition := apidef.APIDefinition{
		Proxy: apidef.ProxyConfig{
			TargetURL: "echo.ns:9000",
		},
	}

	reqBodyJSON := "{Hello: world!}"
	resBodyJSON := "{Foo: Bar}"

	type args struct {
		res *http.Response
		req *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    *models.Telemetry
		wantErr bool
	}{
		{
			name: "",
			args: args{
				res: &http.Response{
					StatusCode: 200,
					Proto:      "HTTP/1.0",
					Header: map[string][]string{
						"Content-Type": {"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(resBodyJSON)),
				},
				req: &http.Request{
					Method: "GET",
					URL: &url.URL{
						Scheme:   "http",
						Path:     "/api",
						RawQuery: "foo=bar",
					},
					Proto: "HTTP/1.0",
					Header: map[string][]string{
						common.RequestIDHeaderKey: {"reqID"},
					},
					Body:       io.NopCloser(strings.NewReader(reqBodyJSON)),
					Host:       "localhost:8080",
					RemoteAddr: "127.0.0.1:54432",
				},
			},
			want: &models.Telemetry{
				DestinationAddress:   ":9000",
				DestinationNamespace: "ns",
				Request: &models.Request{
					Common: &models.Common{
						TruncatedBody: false,
						Body:          []byte(reqBodyJSON),
						Headers: []*models.Header{
							{
								Key:   common.RequestIDHeaderKey,
								Value: "reqID",
							},
						},
						Version: "HTTP/1.0",
						Time:    tNow,
					},
					Host:   "echo.ns",
					Method: "GET",
					Path:   "/api?foo=bar",
				},
				RequestID: "reqID",
				Response: &models.Response{
					Common: &models.Common{
						TruncatedBody: false,
						Body:          []byte(resBodyJSON),
						Headers: []*models.Header{
							{
								Key:   "Content-Type",
								Value: "application/json",
							},
						},
						Version: "HTTP/1.0",
					},
					StatusCode: "200",
				},
				Scheme:        "http",
				SourceAddress: "127.0.0.1:54432",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &user.SessionState{MetaData: map[string]interface{}{common.RequestTimeContextKey: tNow}}
			ctx.SetSession(tt.args.req, session, false, false)

			got, err := createTelemetry(tt.args.res, tt.args.req, &apiDefinition)
			if (err != nil) != tt.wantErr {
				t.Errorf("createTelemetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// no way to predict response time
			got.Response.Common.Time = 0
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createTelemetry() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDestinationNamespaceFromHost(t *testing.T) {
	type args struct {
		host             string
		defaultNamespace string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "namespace exists in host",
			args: args{
				host:             "foo.bar",
				defaultNamespace: "bor",
			},
			want: "bar",
		},
		{
			name: "namespace does not exists in host",
			args: args{
				host:             "foo",
				defaultNamespace: "bor",
			},
			want: "bor",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := common.GetDestinationNamespaceFromHostOrDefault(tt.args.host, tt.args.defaultNamespace); got != tt.want {
				t.Errorf("GetDestinationNamespaceFromHostOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
