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
	UpstreamAddressEnv          = "UPSTREAM_TELEMETRY_HOST_NAME"
	NamespacesToTapEnv          = "NAMESPACES_TO_TAP"
	TapLogLevelEnv              = "TAP_LOG_LEVEL"
	EnableTLSEnv                = "ENABLE_TLS"
	RootCertFilePathEnv         = "ROOT_CERT_FILE_PATH"
	TraceSamplingManagerAddress = "TRACE_SAMPLING_HOST_NAME"
	TraceSamplingEnabled        = "TRACE_SAMPLING_ENABLED"
)

type Config struct {
	NamespaceToTap              []string
	UpstreamAddress             string
	MizuLogLevel                logging.Level
	EnableTLS                   bool
	RootCertFilePath            string
	TraceSamplingManagerAddress string
	TraceSamplingEnabled        bool
}

func LoadConfig() *Config {
	return &Config{
		NamespaceToTap:              viper.GetStringSlice(NamespacesToTapEnv),
		UpstreamAddress:             viper.GetString(UpstreamAddressEnv),
		EnableTLS:                   viper.GetBool(EnableTLSEnv),
		RootCertFilePath:            viper.GetString(RootCertFilePathEnv),
		TraceSamplingManagerAddress: viper.GetString(TraceSamplingManagerAddress),
		TraceSamplingEnabled:        viper.GetBool(TraceSamplingEnabled),
	}
}
