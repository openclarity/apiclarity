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
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	_spec "github.com/apiclarity/speculator/pkg/spec"
	_speculator "github.com/apiclarity/speculator/pkg/speculator"
)

const (
	BackendRestPort            = "BACKEND_REST_PORT"
	HTTPTracesPort             = "HTTP_TRACES_PORT"
	HealthCheckAddress         = "HEALTH_CHECK_ADDRESS"
	StateBackupIntervalSec     = "STATE_BACKUP_INTERVAL_SEC"
	DatabaseCleanerIntervalSec = "DATABASE_CLEANER_INTERVAL_SEC"
	StateBackupFileName        = "STATE_BACKUP_FILE_NAME"

	DBNameEnvVar     = "DB_NAME"
	DBUserEnvVar     = "DB_USER"
	DBPasswordEnvVar = "DB_PASS"
	DBHostEnvVar     = "DB_HOST"
	DBPortEnvVar     = "DB_PORT_NUMBER"
	DatabaseDriver   = "DATABASE_DRIVER"
	EnableDBInfoLogs = "ENABLE_DB_INFO_LOGS"

	ResponseHeadersToIgnore = "RESPONSE_HEADERS_TO_IGNORE"
	RequestHeadersToIgnore  = "REQUEST_HEADERS_TO_IGNORE"
)

type Config struct {
	BackendRestPort            int
	HTTPTracesPort             int
	HealthCheckAddress         string
	StateBackupIntervalSec     int
	DatabaseCleanerIntervalSec int
	StateBackupFileName        string
	SpeculatorConfig           _speculator.Config

	// database config
	DatabaseDriver   string
	DBName           string
	DBUser           string
	DBPassword       string
	DBHost           string
	DBPort           string
	EnableDBInfoLogs bool
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	config.BackendRestPort = viper.GetInt(BackendRestPort)
	config.HTTPTracesPort = viper.GetInt(HTTPTracesPort)
	config.HealthCheckAddress = viper.GetString(HealthCheckAddress)
	config.StateBackupIntervalSec = viper.GetInt(StateBackupIntervalSec)
	config.DatabaseCleanerIntervalSec = viper.GetInt(DatabaseCleanerIntervalSec)
	config.StateBackupFileName = viper.GetString(StateBackupFileName)

	config.DatabaseDriver = viper.GetString(DatabaseDriver)
	config.DBPassword = viper.GetString(DBPasswordEnvVar)
	config.DBUser = viper.GetString(DBUserEnvVar)
	config.DBHost = viper.GetString(DBHostEnvVar)
	config.DBPort = viper.GetString(DBPortEnvVar)
	config.DBName = viper.GetString(DBNameEnvVar)
	config.EnableDBInfoLogs = viper.GetBool(EnableDBInfoLogs)

	config.SpeculatorConfig = createSpeculatorConfig()

	configB, _ := json.Marshal(config)
	log.Infof("\n\nconfig=%s\n\n", configB)

	return config, nil
}

func createSpeculatorConfig() _speculator.Config {
	return _speculator.Config{
		OperationGeneratorConfig: _spec.OperationGeneratorConfig{
			ResponseHeadersToIgnore: viper.GetStringSlice(ResponseHeadersToIgnore),
			RequestHeadersToIgnore:  viper.GetStringSlice(RequestHeadersToIgnore),
		},
	}
}
