// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package core

import (
	"testing"
)

func TestShouldTrace(t *testing.T) {
	type args struct {
		hostsToTrace []string
		host         string
		port         int64
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty hosts mismatch",
			args: args{
				hostsToTrace: []string{},
				host:         "h3",
				port:         86,
			},
			want: false,
		},
		{
			name: "host match",
			args: args{
				hostsToTrace: []string{"h1", "h2:82"},
				host:         "h1",
				port:         80,
			},
			want: true,
		},
		{
			name: "hody and port match",
			args: args{
				hostsToTrace: []string{"h1", "h2:82"},
				host:         "h2",
				port:         82,
			},
			want: true,
		},
		{
			name: "port mismatch",
			args: args{
				hostsToTrace: []string{"h1", "h2:82"},
				host:         "h2",
				port:         85,
			},
			want: false,
		},
		{
			name: "host mismatch",
			args: args{
				hostsToTrace: []string{"h1", "h2:82"},
				host:         "h3",
				port:         82,
			},
			want: false,
		},
		{
			name: "host and port mismatch",
			args: args{
				hostsToTrace: []string{"h1", "h2:82"},
				host:         "h3",
				port:         86,
			},
			want: false,
		},
		{
			name: "wildcard match 1",
			args: args{
				hostsToTrace: []string{"h1", "*", "h2:82"},
				host:         "h3",
				port:         86,
			},
			want: true,
		},
		{
			name: "wildcard match 2",
			args: args{
				hostsToTrace: []string{"h1", "*", "h2:82"},
				host:         "h1",
				port:         86,
			},
			want: true,
		},
		{
			name: "wildcard match 3",
			args: args{
				hostsToTrace: []string{"h1", "*:*", "h2:82"},
				host:         "h1",
				port:         86,
			},
			want: true,
		},
		{
			name: "wildcard host match",
			args: args{
				hostsToTrace: []string{"h1", "*:55", "h2:82"},
				host:         "h8",
				port:         55,
			},
			want: true,
		},
		{
			name: "wildcard host mismatch",
			args: args{
				hostsToTrace: []string{"h1", "*:55", "h2:82"},
				host:         "h8",
				port:         53,
			},
			want: false,
		},
		{
			name: "wildcard port match",
			args: args{
				hostsToTrace: []string{"h1", "h8:*", "h2:82"},
				host:         "h8",
				port:         55,
			},
			want: true,
		},
		{
			name: "wildcard port mismmatch",
			args: args{
				hostsToTrace: []string{"h1", "h8:*", "h2:82"},
				host:         "h9",
				port:         55,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := shouldTrace(tt.args.host, tt.args.port, tt.args.hostsToTrace)
			if res != tt.want {
				t.Errorf("Error in shoudTrace function. host: %s:%d, hostsToTrace: %v. Expected: %v", tt.args.host, tt.args.port, tt.args.hostsToTrace, tt.want)
			}
		})
	}
}
