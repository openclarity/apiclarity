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

package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli"

	"github.com/openclarity/apiclarity/backend/pkg/backend"
	"github.com/openclarity/apiclarity/backend/pkg/config"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/version"
	log_utils "github.com/openclarity/speculator/pkg/utils/log"
)

func run(c *cli.Context) {
	log_utils.InitLogs(c, os.Stdout)
	backend.Run()
}

func versionCommand(_ *cli.Context) {
	fmt.Printf("Version: %s \nCommit: %s\nBuild Time: %s",
		version.Version, version.CommitHash, version.BuildTimestamp)
}

func main() {
	viper.SetDefault(config.HealthCheckAddress, ":8081")
	viper.SetDefault(config.HTTPTracesPort, "9000")
	viper.SetDefault(config.HTTPTracesTLSPort, "9443")
	viper.SetDefault(config.HTTPTraceSamplingManagerPort, "9990")
	viper.SetDefault(config.HTTPSTraceSamplingManagerPort, "9943")
	viper.SetDefault(config.GRPCTraceSamplingManagerPort, "9991")
	viper.SetDefault(config.HostToTraceSecretName, "hosts-to-trace")
	viper.SetDefault(config.HostToTraceSecretNamespace, os.Getenv("POD_NAMESPACE"))
	viper.SetDefault(config.HostToTraceSecretOwnerName, "apiclarity")
	viper.SetDefault(config.BackendRestPort, "8080")
	viper.SetDefault(config.BackendRestTLSPort, "8443")
	viper.SetDefault(config.StateBackupIntervalSec, "30")
	viper.SetDefault(config.DatabaseCleanerIntervalSec, "30")
	viper.SetDefault(config.StateBackupFileName, "state.gob")
	viper.SetDefault(config.DatabaseDriver, database.DBDriverTypePostgres)
	viper.SetDefault(config.EnableK8s, true)
	viper.SetDefault(config.TLSServerCertFilePath, "/etc/certs/server.crt")
	viper.SetDefault(config.TLSServerKeyFilePath, "/etc/certs/server.key")
	viper.SetDefault(config.RootCertFilePath, "/etc/root-ca/ca.crt")
	viper.SetDefault(config.ExternalHTTPTracesTLSPort, "10443")
	viper.AutomaticEnv()
	app := cli.NewApp()
	app.Usage = ""
	app.Name = "APIClarity"
	app.Version = version.Version

	runCommand := cli.Command{
		Name:   "run",
		Usage:  "Starts APIClarity",
		Action: run,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  log_utils.LogLevelFlag,
				Value: log_utils.LogLevelDefaultValue,
				Usage: log_utils.LogLevelFlagUsage,
			},
		},
	}
	runCommand.UsageText = runCommand.Name

	versionCommand := cli.Command{
		Name:   "version",
		Usage:  "APIClarity Version Details",
		Action: versionCommand,
	}
	versionCommand.UsageText = versionCommand.Name

	app.Commands = []cli.Command{
		runCommand,
		versionCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
