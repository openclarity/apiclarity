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

package log

import (
	"io"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	LogLevelFlag         = "log-level"
	LogLevelDefaultValue = "warning"
	LogLevelFlagUsage    = "Sets the logging level (debug, info, warning, error, fatal, panic)"
)

func InitLogs(c *cli.Context, output io.Writer) {
	formatter := log.TextFormatter{
		FullTimestamp:          true,
		DisableTimestamp:       false,
		DisableSorting:         true,
		DisableLevelTruncation: true,
		QuoteEmptyFields:       true,
	}
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&formatter)

	log.SetReportCaller(true)

	log.SetOutput(output)

	// use cmd level if provided
	logLevel := c.String(LogLevelFlag)

	// Only logs with this severity or above will be issued.
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Errorf("invalid log level, setting to be warning: %v", err)
		log.SetLevel(log.WarnLevel)
	} else {
		log.SetLevel(level)
	}
}
