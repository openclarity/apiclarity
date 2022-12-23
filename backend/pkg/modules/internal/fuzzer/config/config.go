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

	logging "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	backendConfig "github.com/openclarity/apiclarity/backend/pkg/config"
)

const (
	// List of supported deployment type.
	DeploymentTypeDocker    = "docker"
	DeploymentTypeFake      = "fake"
	DeploymentTypeConfigMap = "configmap"
)

var SupportedDeployment = map[string]bool{
	DeploymentTypeDocker:    true,
	DeploymentTypeFake:      true,
	DeploymentTypeConfigMap: true,
}

const (
	DeploymentType           = DeploymentTypeConfigMap // One of SUPPORTED_DEPLOYMENT[] value.
	ImageName                = "xxx"
	PlatformType             = "API_CLARITY"
	PlatformHost             = "http://apiclarity-apiclarity:8080/api"
	Fuzzers                  = "scn-fuzzer,restler,crud"
	ShowDockerLogs           = false
	FuzzerTestTraceFile      = "fuzzer-demo-test.json"
	DebugMode                = false
	RestlerTimeBudget        = "0.005" // In hours =~ 20s
	TokenInjectorPath        = "/app/"
	TestReportTimeout        = 60      // In seconds
	RestlerQuickTimeBudget   = "0.016" // In hours =~ 1mn
	RestlerDefaultTimeBudget = "0.16"  // In hours =~ 10mn
	RestlerDeepTimeBudget    = "1"     // In hours
	DefaultJobNamespace      = "apiclarity"

	FuzzerImageNameEnvVar          = "FUZZER_IMAGE_NAME"
	FuzzerDeploymentTypeEnvVar     = "FUZZER_DEPLOYMENT_TYPE"
	FuzzerShowDockerLogsEnvVar     = "FUZZER_SHOW_DOCKER_LOGS"
	FuzzerPlatformTypeEnvVar       = "FUZZER_PLATFORM_TYPE"
	FuzzerPlatformHostEnvVar       = "FUZZER_PLATFORM_HOST"
	FuzzerSubFuzzerEnvVar          = "FUZZER_SUBFUZZER"
	FuzzerTestTraceFileEnvVar      = "FUZZER_TEST_TRACE_FILE"
	FuzzerDebugEnvVar              = "FUZZER_DEBUG"
	FuzzerRestlerTimeBudgetEnvVar  = "FUZZER_RESTLER_TIME_BUDGET"
	FuzzerAuthInjectorPathEnvVar   = "FUZZER_RESTLER_TOKEN_INJECTOR_PATH"
	FuzzerTestReportTimeoutEnvVar  = "FUZZER_TESTREPORT_TIMEOUT"
	FuzzerJobNamespaceEnvVar       = "FUZZER_JOB_NAMESPACE"
	PODNamespaceEnvVar             = "POD_NAMESPACE"
	FuzzerJobTemplateConfigMapName = "FUZZER_JOB_TEMPLATE_CONFIG_MAP_NAME"
)

type Config struct {
	imageName         string
	platformType      string
	deploymentType    string
	platformHost      string
	subFuzzer         string
	showDockerLog     bool
	testTraceFile     string
	debug             bool
	restlerTimeBudget string
	tokenInjectorPath string
	testReportTimeout int

	jobNamespace             string
	jobTemplateConfigMapName string
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

func (c *Config) GetJobNamespace() string {
	return c.jobNamespace
}

func (c *Config) GetJobTemplateConfigMapName() string {
	return c.jobTemplateConfigMapName
}

func (c *Config) Dump() {
	/*
	* properly display the config
	 */
	prefix := "[Fuzzer]"
	logging.Debugf("%v ----------------------", prefix)
	logging.Debugf("%v Fuzzer configuration:", prefix)
	logging.Debugf("%v    imageName         (%v)", prefix, c.imageName)
	logging.Debugf("%v    platformType      (%v)", prefix, c.platformType)
	logging.Debugf("%v    deploymentType    (%v)", prefix, c.deploymentType)
	logging.Debugf("%v    showDockerLog     (%v)", prefix, c.showDockerLog)
	logging.Debugf("%v    platformHost      (%v)", prefix, c.platformHost)
	logging.Debugf("%v    subFuzzer         (%v)", prefix, c.subFuzzer)
	logging.Debugf("%v    testTraceFile     (%v)", prefix, c.testTraceFile)
	logging.Debugf("%v    restlerTimeBudget (%v)", prefix, c.restlerTimeBudget)
	logging.Debugf("%v    tokenInjectorPath (%v)", prefix, c.tokenInjectorPath)
	logging.Debugf("%v    testReportTimeout (%v)", prefix, c.testReportTimeout)
	logging.Debugf("%v    jobNamespace      (%v)", prefix, c.jobNamespace)
	logging.Debugf("%v    jobTemplateConfigMapName (%v)", prefix, c.jobTemplateConfigMapName)
	logging.Debugf("%v ----------------------", prefix)
}

func NewFuzzerConfig() *Config {
	// Set default configuration
	viper.SetDefault(FuzzerImageNameEnvVar, ImageName)
	viper.SetDefault(FuzzerPlatformTypeEnvVar, PlatformType)
	viper.SetDefault(FuzzerDeploymentTypeEnvVar, DeploymentType)
	viper.SetDefault(FuzzerPlatformHostEnvVar, PlatformHost)
	viper.SetDefault(FuzzerSubFuzzerEnvVar, Fuzzers)
	viper.SetDefault(FuzzerShowDockerLogsEnvVar, ShowDockerLogs)
	viper.SetDefault(FuzzerTestTraceFileEnvVar, FuzzerTestTraceFile)
	viper.SetDefault(FuzzerDebugEnvVar, DebugMode)
	viper.SetDefault(FuzzerRestlerTimeBudgetEnvVar, RestlerTimeBudget)
	viper.SetDefault(FuzzerAuthInjectorPathEnvVar, TokenInjectorPath)
	viper.SetDefault(FuzzerTestReportTimeoutEnvVar, TestReportTimeout)
	viper.SetDefault(PODNamespaceEnvVar, DefaultJobNamespace)
	viper.SetDefault(FuzzerJobTemplateConfigMapName, "")

	// FuzzerJobNamespaceEnvVar takes priority over PODNamespaceEnvVar
	jobNamespace := viper.GetString(FuzzerJobNamespaceEnvVar)
	if jobNamespace == "" {
		jobNamespace = viper.GetString(PODNamespaceEnvVar)
	}

	// Create a Fuzzer Configuration

	config := Config{
		imageName:                viper.GetString(FuzzerImageNameEnvVar),
		platformType:             viper.GetString(FuzzerPlatformTypeEnvVar),
		deploymentType:           viper.GetString(FuzzerDeploymentTypeEnvVar),
		platformHost:             viper.GetString(FuzzerPlatformHostEnvVar),
		subFuzzer:                viper.GetString(FuzzerSubFuzzerEnvVar),
		showDockerLog:            viper.GetBool(FuzzerShowDockerLogsEnvVar),
		testTraceFile:            viper.GetString(FuzzerTestTraceFileEnvVar),
		debug:                    viper.GetBool(FuzzerDebugEnvVar),
		restlerTimeBudget:        viper.GetString(FuzzerRestlerTimeBudgetEnvVar),
		tokenInjectorPath:        viper.GetString(FuzzerAuthInjectorPathEnvVar),
		testReportTimeout:        viper.GetInt(FuzzerTestReportTimeoutEnvVar),
		jobNamespace:             jobNamespace,
		jobTemplateConfigMapName: viper.GetString(FuzzerJobTemplateConfigMapName),
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
