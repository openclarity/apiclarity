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

package config

import (
	"os"
	"reflect"
	"testing"

	"github.com/spf13/viper"
	"gotest.tools/v3/assert"
)

func TestLoadConfig(t *testing.T) {
	viper.AutomaticEnv()
	tests := []struct {
		name               string
		namespacesToTapEnv string
		upstreamAddressEnv string
		want               *Config
	}{
		{
			name:               "good",
			namespacesToTapEnv: "ns1 ns2",
			upstreamAddressEnv: "addr:80",
			want: &Config{
				NamespaceToTap:  []string{"ns1", "ns2"},
				UpstreamAddress: "addr:80",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv(UpstreamAddressEnv, tt.upstreamAddressEnv)
			assert.NilError(t, err)
			err = os.Setenv(NamespacesToTapEnv, tt.namespacesToTapEnv)
			assert.NilError(t, err)

			got := LoadConfig()
			os.Unsetenv(NamespacesToTapEnv)
			os.Unsetenv(UpstreamAddressEnv)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
