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

package apiclarityexporter

import (
	"errors"

	otelconfig "go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

// Config defines configuration for APIClarity HTTP exporter.
type Config struct {
	otelconfig.ExporterSettings   `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct
	confighttp.HTTPClientSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.
	exporterhelper.QueueSettings  `mapstructure:"sending_queue"`
	exporterhelper.RetrySettings  `mapstructure:"retry_on_failure"`
	PreferHostNames               bool `mapstructure:"prefer_hostnames"`
}

var _ otelconfig.Exporter = (*Config)(nil)

// Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {
	if cfg.HTTPClientSettings.Endpoint == "" {
		return errors.New("at least one endpoint must be specified")
	}

	return nil
}
