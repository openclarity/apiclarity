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

package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"

	backendConfig "github.com/openclarity/apiclarity/backend/pkg/config"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/logging"
)

const (
	// List of supported deployment type.
	DeploymentTypeDocker     = "docker"
	DeploymentTypeKubernetes = "kubernetes"
	DeploymentTypeFake       = "fake"
)

var SupportedDeployment = map[string]bool{
	DeploymentTypeDocker:     true,
	DeploymentTypeKubernetes: true,
	DeploymentTypeFake:       true,
}

const (
	DeploymentType           = DeploymentTypeFake // One of SUPPORTED_DEPLOYMENT[] value.
	ImageName                = "xxx"
	PlatformType             = "API_CLARITY"
	PlatformHost             = "http://localhost:8080/api"
	PlatformHostFromDocker   = "http://apiclarity-apiclarity:8080/api"
	Fuzzers                  = "scn-fuzzer,restler,crud"
	ShowDockerLogs           = false
	FuzzerTestTraceFile      = "data.txt"
	DebugMode                = false
	RestlerTimeBudget        = "0.005" // In hours =~ 20s
	TokenInjectorPath        = "/app/"
	TestReportTimeout        = 30      // In seconds
	RestlerQuickTimeBudget   = "0.016" // In hours =~ 1mn
	RestlerDefaultTimeBudget = "0.16"  // In hours =~ 10mn
	RestlerDeepTimeBudget    = "1"     // In hours

	FuzzerImageNameEnvVar            = "FUZZER_IMAGE_NAME"
	FuzzerDeploymentTypeEnvVar       = "FUZZER_DEPLOYMENT_TYPE"
	FuzzerShowDockerLogsEnvVar       = "FUZZER_SHOW_DOCKER_LOGS"
	FuzzerPlatformTypeEnvVar         = "FUZZER_PLATFORM_TYPE"
	FuzzerPlatformHostEnvVar         = "FUZZER_PLATFORM_HOST"
	FuzzerPlatformHostFromFuzzEnvVar = "FUZZER_PLATFORM_HOST_FROM_FUZZER"
	FuzzerSubFuzzerEnvVar            = "FUZZER_SUBFUZZER"
	FuzzerTestTraceFileEnvVar        = "FUZZER_TEST_TRACE_FILE"
	FuzzerDebugEnvVar                = "FUZZER_DEBUG"
	FuzzerRestlerTimeBudgetEnvVar    = "FUZZER_RESTLER_TIME_BUDGET"
	FuzzerAuthInjectorPathEnvVar     = "FUZZER_RESTLER_TOKEN_INJECTOR_PATH"
	FuzzerTestReportTimeoutEnvVar    = "FUZZER_TESTREPORT_TIMEOUT"
)

type Config struct {
	imageName              string
	platformType           string
	deploymentType         string
	platformHost           string
	platformHostFromFuzzer string
	subFuzzer              string
	showDockerLog          bool
	testTraceFile          string
	debug                  bool
	restlerTimeBudget      string
	tokenInjectorPath      string
	testReportTimeout      int
}

// the Mutex to restrict access to configSingleton.
var lock sync.Mutex

// config singleton.
var configSingleton *Config

// return the singleton instance.
func GetConfig() *Config {
	lock.Lock()
	defer lock.Unlock()

	if configSingleton == nil {
		configSingleton = NewFuzzerConfig()
	}
	return configSingleton
}

// Accessors as member are not visible from external.
func (c *Config) GetImageName() string {
	return c.imageName
}

func (c *Config) GetDeploymentType() string {
	return c.deploymentType
}

func (c *Config) GetShowDockerLogFlag() bool {
	return c.showDockerLog
}

func (c *Config) GetPlatformType() string {
	return c.platformType
}

func (c *Config) GetPlatformHost() string {
	return c.platformHost
}

func (c *Config) GetPlatformHostFromFuzzer() string {
	return c.platformHostFromFuzzer
}

func (c *Config) GetSubFuzzerList() string {
	return c.subFuzzer
}

func (c *Config) GetFakeFileName() string {
	return c.testTraceFile
}

func (c *Config) GetRestlerTimeBudget() string {
	return c.restlerTimeBudget
}

func (c *Config) GetRestlerTokenInjectorPath() string {
	return c.tokenInjectorPath
}

func (c *Config) GetTestReportTimeout() int {
	return c.testReportTimeout
}

func (c *Config) Dump() {
	/*
	* properly display the config
	 */
	prefix := "[Fuzzer]"
	logging.Logf("%v ----------------------", prefix)
	logging.Logf("%v Fuzzer configuration:", prefix)
	logging.Logf("%v    imageName         (%v)", prefix, c.imageName)
	logging.Logf("%v    platformType      (%v)", prefix, c.platformType)
	logging.Logf("%v    deploymentType    (%v)", prefix, c.deploymentType)
	logging.Logf("%v    showDockerLog     (%v)", prefix, c.showDockerLog)
	logging.Logf("%v    platformHost      (%v)", prefix, c.platformHost)
	logging.Logf("%v    platformHostFromFuzzer (%v)", prefix, c.platformHostFromFuzzer)
	logging.Logf("%v    subFuzzer         (%v)", prefix, c.subFuzzer)
	logging.Logf("%v    testTraceFile     (%v)", prefix, c.testTraceFile)
	logging.Logf("%v    restlerTimeBudget (%v)", prefix, c.restlerTimeBudget)
	logging.Logf("%v    tokenInjectorPath (%v)", prefix, c.tokenInjectorPath)
	logging.Logf("%v    testReportTimeout (%v)", prefix, c.testReportTimeout)
	logging.Logf("%v ----------------------", prefix)
}

func NewFuzzerConfig() *Config {
	// Set default configuration
	viper.SetDefault(FuzzerImageNameEnvVar, ImageName)
	viper.SetDefault(FuzzerPlatformTypeEnvVar, PlatformType)
	viper.SetDefault(FuzzerDeploymentTypeEnvVar, DeploymentType)
	viper.SetDefault(FuzzerPlatformHostEnvVar, PlatformHost)
	viper.SetDefault(FuzzerPlatformHostFromFuzzEnvVar, PlatformHostFromDocker)
	viper.SetDefault(FuzzerSubFuzzerEnvVar, Fuzzers)
	viper.SetDefault(FuzzerShowDockerLogsEnvVar, ShowDockerLogs)
	viper.SetDefault(FuzzerTestTraceFileEnvVar, FuzzerTestTraceFile)
	viper.SetDefault(FuzzerDebugEnvVar, DebugMode)
	viper.SetDefault(FuzzerRestlerTimeBudgetEnvVar, RestlerTimeBudget)
	viper.SetDefault(FuzzerAuthInjectorPathEnvVar, TokenInjectorPath)
	viper.SetDefault(FuzzerTestReportTimeoutEnvVar, TestReportTimeout)

	// Create a Fuzzer Configuration

	config := Config{
		imageName:              viper.GetString(FuzzerImageNameEnvVar),
		platformType:           viper.GetString(FuzzerPlatformTypeEnvVar),
		deploymentType:         viper.GetString(FuzzerDeploymentTypeEnvVar),
		platformHost:           viper.GetString(FuzzerPlatformHostEnvVar),
		platformHostFromFuzzer: viper.GetString(FuzzerPlatformHostFromFuzzEnvVar),
		subFuzzer:              viper.GetString(FuzzerSubFuzzerEnvVar),
		showDockerLog:          viper.GetBool(FuzzerShowDockerLogsEnvVar),
		testTraceFile:          viper.GetString(FuzzerTestTraceFileEnvVar),
		debug:                  viper.GetBool(FuzzerDebugEnvVar),
		restlerTimeBudget:      viper.GetString(FuzzerRestlerTimeBudgetEnvVar),
		tokenInjectorPath:      viper.GetString(FuzzerAuthInjectorPathEnvVar),
		testReportTimeout:      viper.GetInt(FuzzerTestReportTimeoutEnvVar),
	}

	// ... then override by env, if present

	modulesAssets := viper.GetString(backendConfig.ModulesAssetsEnvVar)
	if modulesAssets != "" {
		// check if data.txt exists
		dataFilename := filepath.Join(modulesAssets, "fuzzer", config.testTraceFile)
		if s, err := os.Stat(dataFilename); err == nil && s.Mode().IsRegular() {
			config.testTraceFile = dataFilename
		}
	}

	return &config
}
