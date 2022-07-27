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
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

const (
	UpstreamTelemetryAddressEnv = "UPSTREAM_TELEMETRY_ADDRESS"
	TapperNamespace             = "TAPPER_NAMESPACE"
	NamespacesToTapEnv          = "NAMESPACES_TO_TAP"
	TapLogLevelEnv              = "TAP_LOG_LEVEL"
	EnableTLSEnv                = "ENABLE_TLS"
	TraceSamplingAddressEnv     = "TRACE_SAMPLING_ADDRESS"
	TraceSamplingEnabled        = "TRACE_SAMPLING_ENABLED"
)

type Config struct {
	NamespaceToTap           []string
	TapperNamespace          string
	UpstreamTelemetryAddress string
	MizuLogLevel             logging.Level
	EnableTLS                bool
	TraceSamplingAddress     string
	TraceSamplingEnabled     bool
}

func LoadConfig() *Config {
	return &Config{
		NamespaceToTap:           viper.GetStringSlice(NamespacesToTapEnv),
		UpstreamTelemetryAddress: viper.GetString(UpstreamTelemetryAddressEnv),
		EnableTLS:                viper.GetBool(EnableTLSEnv),
		TraceSamplingAddress:     viper.GetString(TraceSamplingAddressEnv),
		TraceSamplingEnabled:     viper.GetBool(TraceSamplingEnabled),
		TapperNamespace:          viper.GetString(TapperNamespace),
	}
}
