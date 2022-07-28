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
	"testing"

	"github.com/openclarity/apiclarity/plugins/common/trace_sampling_client"
)

func TestAgent_shouldTrace(t *testing.T) {
	type fields struct {
		traceSamplingManager *trace_sampling_client.Client
	}
	type args struct {
		host string
		port string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "traceSamplingManager == nil",
			fields: fields{
				traceSamplingManager: nil,
			},
			args: args{
				host: "host1",
				port: "80",
			},
			want: true,
		},
		{
			name: "found",
			fields: fields{
				traceSamplingManager: &trace_sampling_client.Client{
					Hosts: map[string]bool{
						"host1.ns1:80": true,
						"host2.ns2:80": true,
					},
				},
			},
			args: args{
				host: "host1.ns1",
				port: "80",
			},
			want: true,
		},
		{
			name: "not found",
			fields: fields{
				traceSamplingManager: &trace_sampling_client.Client{
					Hosts: map[string]bool{
						"host1.ns1:80": true,
						"host2.ns2:80": true,
					},
				},
			},
			args: args{
				host: "host3",
				port: "80",
			},
			want: false,
		},
		{
			name: "all",
			fields: fields{
				traceSamplingManager: &trace_sampling_client.Client{
					Hosts: map[string]bool{
						"host1.ns1": true,
						"host2.ns2": true,
						"*":         true,
					},
				},
			},
			args: args{
				host: "host3",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				traceSamplingClient: tt.fields.traceSamplingManager,
			}
			if got := a.shouldTrace(tt.args.host, tt.args.port); got != tt.want {
				t.Errorf("shouldTrace() = %v, want %v", got, tt.want)
			}
		})
	}
}
