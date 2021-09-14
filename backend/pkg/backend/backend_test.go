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

package backend

import (
	"testing"

	"github.com/apiclarity/apiclarity/api/server/models"
	_spec "github.com/apiclarity/speculator/pkg/spec"
)

func Test_isNonAPI(t *testing.T) {
	type args struct {
		trace *_spec.SCNTelemetry
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "content type is not application/json expected to classify as non API",
			args: args{
				trace: &_spec.SCNTelemetry{
					SCNTResponse: _spec.SCNTResponse{
						SCNTCommon: _spec.SCNTCommon{
							Headers: [][2]string{{contentTypeHeaderName, "non-api"}},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "REST API",
			args: args{
				trace: &_spec.SCNTelemetry{
					SCNTResponse: _spec.SCNTResponse{
						SCNTCommon: _spec.SCNTCommon{
							Headers: [][2]string{{contentTypeHeaderName, contentTypeApplicationJSON}},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "no headers expected to classify as API",
			args: args{
				trace: &_spec.SCNTelemetry{
					SCNTResponse: _spec.SCNTResponse{
						SCNTCommon: _spec.SCNTCommon{},
					},
				},
			},
			want: false,
		},
		{
			name: "content type is application/hal+json - classify as API",
			args: args{
				trace: &_spec.SCNTelemetry{
					SCNTResponse: _spec.SCNTResponse{
						SCNTCommon: _spec.SCNTCommon{
							Headers: [][2]string{{contentTypeHeaderName, "application/hal+json"}},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNonAPI(tt.args.trace); got != tt.want {
				t.Errorf("isNonAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getHostname(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "no scheme",
			args: args{
				host: "example.com:8080",
			},
			want: "example.com",
		},
		{
			name: "with scheme",
			args: args{
				host: "acap://example.com:8080",
			},
			want: "example.com",
		},
		{
			name: "only host",
			args: args{
				host: "example.com",
			},
			want: "example.com",
		},
		{
			name: "hostname is empty",
			args: args{
				host: "https://",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "failed to parse host",
			args: args{
				host: "1 2 3",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getHostname(tt.args.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHostname() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getHostname() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAPIDiffType(t *testing.T) {
	type args struct {
		diffType _spec.DiffType
	}
	tests := []struct {
		name string
		args args
		want models.DiffType
	}{
		{
			name: "unknown type - default DiffTypeNODIFF",
			args: args{
				diffType: "unknown type",
			},
			want: models.DiffTypeNODIFF,
		},
		{
			name: "DiffTypeNoDiff",
			args: args{
				diffType: _spec.DiffTypeNoDiff,
			},
			want: models.DiffTypeNODIFF,
		},
		{
			name: "DiffTypeZombieDiff",
			args: args{
				diffType: _spec.DiffTypeZombieDiff,
			},
			want: models.DiffTypeZOMBIEDIFF,
		},
		{
			name: "DiffTypeShadowDiff",
			args: args{
				diffType: _spec.DiffTypeShadowDiff,
			},
			want: models.DiffTypeSHADOWDIFF,
		},
		{
			name: "DiffTypeGeneralDiff",
			args: args{
				diffType: _spec.DiffTypeGeneralDiff,
			},
			want: models.DiffTypeGENERALDIFF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAPIDiffType(tt.args.diffType); got != tt.want {
				t.Errorf("getAPIDiffType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSpecDiffType(t *testing.T) {
	type args struct {
		providedDiff      models.DiffType
		reconstructedDiff models.DiffType
	}
	tests := []struct {
		name string
		args args
		want models.DiffType
	}{
		{
			name: "Zombie over Shadow",
			args: args{
				providedDiff:      models.DiffTypeZOMBIEDIFF,
				reconstructedDiff: models.DiffTypeSHADOWDIFF,
			},
			want: models.DiffTypeZOMBIEDIFF,
		},
		{
			name: "Same type",
			args: args{
				providedDiff:      models.DiffTypeGENERALDIFF,
				reconstructedDiff: models.DiffTypeGENERALDIFF,
			},
			want: models.DiffTypeGENERALDIFF,
		},
		{
			name: "reconstructed unknown type",
			args: args{
				providedDiff:      models.DiffTypeNODIFF,
				reconstructedDiff: "unknown type",
			},
			want: models.DiffTypeNODIFF,
		},
		{
			name: "provided unknown type",
			args: args{
				providedDiff:      "unknown type",
				reconstructedDiff: models.DiffTypeNODIFF,
			},
			want: models.DiffTypeNODIFF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSpecDiffType(tt.args.providedDiff, tt.args.reconstructedDiff); got != tt.want {
				t.Errorf("getSpecDiffType() = %v, want %v", got, tt.want)
			}
		})
	}
}
