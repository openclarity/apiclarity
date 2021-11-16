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
	"os"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/apiclarity/apiclarity/api/server/restapi"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	_database "github.com/apiclarity/apiclarity/backend/pkg/database"
)

const (
	debugEnvVar            = "DEBUG"
	serverPortEnvVar       = "SERVER_PORT"
	serverPortDefaultValue = 8080
)

func initLogs() {
	formatter := log.TextFormatter{
		FullTimestamp:          true,
		DisableTimestamp:       false,
		DisableSorting:         true,
		DisableLevelTruncation: true,
		QuoteEmptyFields:       true,
	}

	log.SetFormatter(&formatter)

	log.SetReportCaller(true)

	log.SetOutput(os.Stdout)

	if viper.GetBool(debugEnvVar) {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	viper.SetDefault(serverPortEnvVar, serverPortDefaultValue)

	initLogs()

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewAPIClarityAPIsAPI(swaggerSpec)

	dbHandler := _database.Init(&_database.DBConfig{DriverType: _database.DBDriverTypeLocal})

	server := restapi.NewServer(api)
	defer func() { _ = server.Shutdown() }()

	server.ConfigureFlags()
	server.ConfigureAPI()
	server.Port = viper.GetInt(serverPortEnvVar)

	if viper.GetBool(_database.FakeDataEnvVar) {
		go dbHandler.CreateFakeData()
	}

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}
