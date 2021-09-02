/*
 *
 * Copyright (c) 2020 Cisco Systems, Inc. and its affiliates.
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package config

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
)

const (
	BackendRestPort            = "BACKEND_REST_PORT"
	HttpTracesPort             = "HTTP_TRACES_PORT"
	HealthCheckAddress         = "HEALTH_CHECK_ADDRESS"
	StateBackupIntervalSec     = "STATE_BACKUP_INTERVAL_SEC"
	DatabaseCleanerIntervalSec = "DATABASE_CLEANER_INTERVAL_SEC"
	StateBackupFileName        = "STATE_BACKUP_FILE_NAME"
)

type Config struct {
	BackendRestPort            int
	HttpTracesPort             int
	HealthCheckAddress         string
	StateBackupIntervalSec     int
	DatabaseCleanerIntervalSec int
	StateBackupFileName        string
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	config.BackendRestPort = viper.GetInt(BackendRestPort)
	config.HttpTracesPort = viper.GetInt(HttpTracesPort)
	config.HealthCheckAddress = viper.GetString(HealthCheckAddress)
	config.StateBackupIntervalSec = viper.GetInt(StateBackupIntervalSec)
	config.DatabaseCleanerIntervalSec = viper.GetInt(DatabaseCleanerIntervalSec)
	config.StateBackupFileName = viper.GetString(StateBackupFileName)

	configB, _ := json.Marshal(config)
	fmt.Printf("\n\nconfig=%s\n\n", configB)

	return config, nil
}
