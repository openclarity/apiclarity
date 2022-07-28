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
	"flag"
	"os"

	logutils "github.com/Portshift/go-utils/log"
	"github.com/op/go-logging"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli"

	"github.com/openclarity/apiclarity/plugins/taper/config"
	"github.com/openclarity/apiclarity/plugins/taper/version"
)

// We need to set this here since mizu code is using a global flags to define the flags, otherwise it will not work.
var _ = flag.String(logutils.LogLevelFlag, logutils.LogLevelDefaultValue, logutils.LogLevelFlagUsage)

func main() {
	viper.SetDefault(config.UpstreamAddressEnv, "apiclarity-apiclarity.apiclarity:9000")
	viper.SetDefault(config.TraceSamplingManagerAddress, "apiclarity-apiclarity.apiclarity:9990")
	viper.SetDefault(config.TraceSamplingEnabled, false)
	viper.SetDefault(config.NamespacesToTapEnv, "default")
	viper.SetDefault(config.EnableTLSEnv, false)
	viper.SetDefault(config.TapLogLevelEnv, logging.INFO)

	viper.AutomaticEnv()

	app := cli.NewApp()
	app.Name = "APIClarity packet source tap"
	app.Version = version.Version

	app.Usage = ""
	app.UsageText = ""
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "i",
			Usage: "interface to tap",
			Value: "any",
		},
		cli.BoolFlag{
			Name: "nodefrag",
		},
		cli.StringFlag{
			Name:  logutils.LogLevelFlag,
			Value: logutils.LogLevelDefaultValue,
			Usage: logutils.LogLevelFlagUsage,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
