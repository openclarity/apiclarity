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

package trace_sampling_client

import (
	"github.com/openclarity/trace-sampling-manager/api/client/client"
	"gotest.tools/assert"
	"sync"
	"testing"
	"time"
)

func TestTraceSamplingManager_setHosts(t1 *testing.T) {
	type fields struct {
		TraceSamplingManagerClient *client.TraceSamplingManager
		Hosts                      map[string]bool
		samplingInterval           time.Duration
		lock                       sync.RWMutex
	}
	type args struct {
		hosts []string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantHosts map[string]bool
	}{
		{
			name: "set on empty",
			fields: fields{
				Hosts: map[string]bool{},
			},
			args: args{
				hosts: []string{"host1", "host2"},
			},
			wantHosts: map[string]bool{
				"host1": true,
				"host2": true,
			},
		},
		{
			name: "set on existing",
			fields: fields{
				Hosts: map[string]bool{
					"host3": true,
				},
			},
			args: args{
				hosts: []string{"host1", "host2"},
			},
			wantHosts: map[string]bool{
				"host1": true,
				"host2": true,
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t *testing.T) {
			tsm := &Client{
				Hosts: tt.fields.Hosts,
			}
			tsm.setHosts(tt.args.hosts)
			assert.DeepEqual(t, tt.wantHosts, tsm.Hosts)
		})
	}
}

func TestTraceSamplingManager_ShouldTrace(t1 *testing.T) {
	type fields struct {
		TraceSamplingManagerClient *client.TraceSamplingManager
		Hosts                      map[string]bool
		samplingInterval           time.Duration
		lock                       sync.RWMutex
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
			name: "should trace - no port",
			fields: fields{
				Hosts: map[string]bool{
					"host1": true,
					"host2": true,
					"host3": true,
				},
			},
			args: args{
				host: "host1",
				port: "",
			},
			want: true,
		},
		{
			name: "should not trace - no port",
			fields: fields{
				Hosts: map[string]bool{
					"host1": true,
					"host2": true,
					"host3": true,
				},
			},
			args: args{
				host: "host4",
			},
			want: false,
		},
		{
			name: "should trace - with port",
			fields: fields{
				Hosts: map[string]bool{
					"host1:8080": true,
					"host2":      true,
					"host3":      true,
				},
			},
			args: args{
				host: "host1",
				port: "8080",
			},
			want: true,
		},
		{
			name: "should not trace - with port",
			fields: fields{
				Hosts: map[string]bool{
					"host1:9000": true,
					"host2":      true,
					"host3":      true,
				},
			},
			args: args{
				host: "host1:8000",
			},
			want: false,
		},
		{
			name: "should trace by wildcard",
			fields: fields{
				Hosts: map[string]bool{
					"host1": true,
					"host2": true,
					"*":     true,
				},
			},
			args: args{
				host: "host4",
			},
			want: true,
		},
		{
			name: "should not trace - empty list",
			fields: fields{
				Hosts: map[string]bool{},
			},
			args: args{
				host: "host4",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Client{
				Hosts: tt.fields.Hosts,
			}
			if got := t.ShouldTrace(tt.args.host, tt.args.port); got != tt.want {
				t1.Errorf("ShouldTrace() = %v, want %v", got, tt.want)
			}
		})
	}
}
