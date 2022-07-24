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

package notifier

import "testing"

func Test_setSchemeIfNeeded(t *testing.T) {
	type args struct {
		url    string
		scheme string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "url already have a scheme",
			args: args{
				url:    "https://example.com",
				scheme: "http",
			},
			want: "https://example.com",
		},
		{
			name: "url is missing a scheme",
			args: args{
				url:    "example.com",
				scheme: "http",
			},
			want: "http://example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setSchemeIfNeeded(tt.args.url, tt.args.scheme); got != tt.want {
				t.Errorf("setSchemeIfNeeded() = %v, want %v", got, tt.want)
			}
		})
	}
}
