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

	_spec "github.com/openclarity/speculator/pkg/spec"
	_speculator "github.com/openclarity/speculator/pkg/speculator"
)

const (
	BackendRestPort               = "BACKEND_REST_PORT"
	BackendRestTLSPort            = "BACKEND_REST_TLS_PORT"
	TraceSamplingEnabled          = "TRACE_SAMPLING_ENABLED"
	HTTPTracesPort                = "HTTP_TRACES_PORT"
	HTTPTracesTLSPort             = "HTTP_TRACES_TLS_PORT"
	HTTPTraceSamplingManagerPort  = "HTTP_TRACE_SAMPLING_MANAGER_PORT"
	HTTPSTraceSamplingManagerPort = "HTTPS_TRACE_SAMPLING_MANAGER_PORT"
	GRPCTraceSamplingManagerPort  = "GRPC_TRACE_SAMPLING_MANAGER_PORT"
	HostToTraceSecretName         = "HOST_TO_TRACE_SECRET_NAME"       //nolint:gosec
	HostToTraceSecretNamespace    = "HOST_TO_TRACE_SECRET_NAMESPACE"  //nolint:gosec
	HostToTraceSecretOwnerName    = "HOST_TO_TRACE_SECRET_OWNER_NAME" //nolint:gosec
	HealthCheckAddress            = "HEALTH_CHECK_ADDRESS"
	StateBackupIntervalSec        = "STATE_BACKUP_INTERVAL_SEC"
	DatabaseCleanerIntervalSec    = "DATABASE_CLEANER_INTERVAL_SEC"
	StateBackupFileName           = "STATE_BACKUP_FILE_NAME"
	NoMonitorEnvVar               = "NO_K8S_MONITOR"
	K8sLocalEnvVar                = "K8S_LOCAL"
	EnableK8s                     = "ENABLE_K8S"
	EnableTLS                     = "ENABLE_TLS"
	TLSServerCertFilePath         = "TLS_SERVER_CERT_FILE_PATH"
	TLSServerKeyFilePath          = "TLS_SERVER_KEY_FILE_PATH"
	RootCertFilePath              = "ROOT_CERT_FILE_PATH"

	ExternalHTTPTracesTLSPort = "EXTERNAL_HTTP_TRACES_TLS_PORT"

	DBNameEnvVar     = "DB_NAME"
	DBUserEnvVar     = "DB_USER"
	DBPasswordEnvVar = "DB_PASS"
	DBHostEnvVar     = "DB_HOST"
	DBPortEnvVar     = "DB_PORT_NUMBER"
	DatabaseDriver   = "DATABASE_DRIVER"
	EnableDBInfoLogs = "ENABLE_DB_INFO_LOGS"

	ResponseHeadersToIgnore = "RESPONSE_HEADERS_TO_IGNORE"
	RequestHeadersToIgnore  = "REQUEST_HEADERS_TO_IGNORE"

	ModulesAssetsEnvVar = "MODULES_ASSETS"

	NotificationPrefix = "NOTIFICATION_BACKEND_PREFIX"
)

type Config struct {
	BackendRestPort            int
	BackendRestTLSPort         int
	HTTPTracesPort             int
	HTTPTracesTLSPort          int
	HealthCheckAddress         string
	StateBackupIntervalSec     int
	DatabaseCleanerIntervalSec int
	StateBackupFileName        string
	SpeculatorConfig           _speculator.Config
	K8sLocal                   bool
	EnableK8s                  bool
	EnableTLS                  bool
	TLSServerCertFilePath      string
	TLSServerKeyFilePath       string
	RootCertFilePath           string

	NotificationPrefix string

	// External HTTP Trace server
	ExternalHTTPTracesTLSPort int

	// trace sampling config
	HTTPTraceSamplingManagerPort  int
	HTTPSTraceSamplingManagerPort int
	GRPCTraceSamplingManagerPort  int
	TraceSamplingEnabled          bool
	HostToTraceSecretName         string
	HostToTraceSecretNamespace    string
	HostToTraceSecretOwnerName    string

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
	config.BackendRestTLSPort = viper.GetInt(BackendRestTLSPort)
	config.HTTPTracesPort = viper.GetInt(HTTPTracesPort)
	config.HTTPTracesTLSPort = viper.GetInt(HTTPTracesTLSPort)
	config.HTTPTraceSamplingManagerPort = viper.GetInt(HTTPTraceSamplingManagerPort)
	config.HTTPSTraceSamplingManagerPort = viper.GetInt(HTTPSTraceSamplingManagerPort)
	config.GRPCTraceSamplingManagerPort = viper.GetInt(GRPCTraceSamplingManagerPort)
	config.HostToTraceSecretName = viper.GetString(HostToTraceSecretName)
	config.HostToTraceSecretNamespace = viper.GetString(HostToTraceSecretNamespace)
	config.HostToTraceSecretOwnerName = viper.GetString(HostToTraceSecretOwnerName)
	config.TraceSamplingEnabled = viper.GetBool(TraceSamplingEnabled)
	config.HealthCheckAddress = viper.GetString(HealthCheckAddress)
	config.StateBackupIntervalSec = viper.GetInt(StateBackupIntervalSec)
	config.DatabaseCleanerIntervalSec = viper.GetInt(DatabaseCleanerIntervalSec)
	config.StateBackupFileName = viper.GetString(StateBackupFileName)
	config.TLSServerCertFilePath = viper.GetString(TLSServerCertFilePath)
	config.TLSServerKeyFilePath = viper.GetString(TLSServerKeyFilePath)
	config.RootCertFilePath = viper.GetString(RootCertFilePath)
	config.NotificationPrefix = viper.GetString(NotificationPrefix)

	config.ExternalHTTPTracesTLSPort = viper.GetInt(ExternalHTTPTracesTLSPort)

	config.K8sLocal = viper.GetBool(K8sLocalEnvVar)
	config.EnableK8s = viper.GetBool(EnableK8s)
	config.EnableTLS = viper.GetBool(EnableTLS)
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
