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

package utils

import (
	"reflect"
	"sort"
	"testing"
)

func TestMapToSlice(t *testing.T) {
	type args struct {
		m map[string]bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "sanity",
			args: args{
				m: map[string]bool{
					"1": true,
					"2": true,
					"3": true,
				},
			},
			want: []string{"1", "2", "3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapToSlice(tt.args.m)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapToSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
