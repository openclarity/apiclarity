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

package database

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/go-openapi/strfmt"
)

var (
	oldSpec = "{\n  \"swagger\": \"2.0\",\n  \"info\": {\n    \"title\": \"Sample API\",\n    \"description\": \"API description in Markdown.\",\n    \"version\": \"1.0.0\"\n  },\n  \"host\": \"api.example.com\",\n  \"basePath\": \"/v1\",\n  \"schemes\": [\n    \"https\"\n  ],\n  \"paths\": {\n    \"/cats\": {\n      \"get\": {\n        \"summary\": \"Returns a list of cats.\",\n        \"description\": \"Optional extended description in Markdown.\",\n        \"produces\": [\n          \"application/json\"\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"OK\"\n          }\n        }\n      }\n    }\n  }\n}"
	newSpec = "{\n  \"swagger\": \"2.0\",\n  \"info\": {\n    \"title\": \"Sample API\",\n    \"description\": \"API description in Markdown.\",\n    \"version\": \"1.0.0\"\n  },\n  \"host\": \"api.example.com\",\n  \"basePath\": \"/v1\",\n  \"schemes\": [\n    \"https\"\n  ],\n  \"paths\": {\n    \"/cats\": {\n      \"get\": {\n        \"summary\": \"Returns a list of cats.\",\n        \"produces\": [\n          \"application/json\"\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"OK\"\n          }\n        }\n      }\n    },\n    \"/dogs\": {\n      \"get\": {\n        \"summary\": \"Returns a list of dogs.\",\n        \"produces\": [\n          \"application/json\"\n        ],\n        \"responses\": {\n          \"200\": {\n            \"description\": \"OK\"\n          }\n        }\n      }\n    }\n  }\n}"
)

func genRandIPAddr() string {
	ip := fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
	return ip
}

func customGenerator() {
	_ = faker.AddProvider("sourceIP", func(v reflect.Value) (interface{}, error) {
		return genRandIPAddr(), nil
	})
	_ = faker.AddProvider("destinationIP", func(v reflect.Value) (interface{}, error) {
		return genRandIPAddr(), nil
	})
}

func createAPIEvent() *APIEvent {
	var event APIEvent

	if err := faker.FakeData(&event); err != nil {
		panic(err)
	}

	event.Time = strfmt.DateTime(time.Now().Add(-(time.Duration(rand.Int()%48) * time.Hour)))

	return &event
}

func createAPIInfo() *APIInfo {
	var event APIInfo

	if err := faker.FakeData(&event); err != nil {
		panic(err)
	}

	return &event
}

func (db *Handler) CreateFakeData() {
	rand.Seed(time.Now().Unix())
	time.Sleep(1 * time.Second)
	customGenerator()

	for i := 0; i < 10; i++ {
		apiInfo := createAPIInfo()
		if apiInfo.HasProvidedSpec {
			apiInfo.ProvidedSpec = newSpec
		}
		if apiInfo.HasReconstructedSpec {
			apiInfo.ReconstructedSpec = oldSpec
		}
		// put in table to get ID
		db.APIInventoryTable().CreateAPIInfo(apiInfo)
		for i := 0; i < rand.Int()%50; i++ {
			apiEvent := createAPIEvent()
			apiEvent.APIInfoID = apiInfo.ID
			// set api event spec name & port to be the same as api info spec name & port
			apiEvent.HostSpecName = apiInfo.Name
			apiEvent.EventType = apiInfo.Type
			apiEvent.DestinationPort = apiInfo.Port
			if apiEvent.HasReconstructedSpecDiff {
				apiEvent.OldReconstructedSpec = oldSpec
				apiEvent.NewReconstructedSpec = newSpec
			}

			db.APIEventsTable().CreateAPIEvent(apiEvent)
		}
	}

	// Create Non APIs
	for i := 0; i < 50; i++ {
		apiEvent := createAPIEvent()
		apiEvent.HasReconstructedSpecDiff = false
		apiEvent.IsNonAPI = true
		apiEvent.Path = "/images/image.png"
		apiEvent.Method = "GET"

		db.APIEventsTable().CreateAPIEvent(apiEvent)
	}
}
