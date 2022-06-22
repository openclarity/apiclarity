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

package model

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/config"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/logging"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/restapi"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/tools"
)

var typeToNameMap = map[string]string{
	"INTERNAL_SERVER_ERROR":        "Internal Server Error",
	"NOT_IMPLEMENTED_ERROR":        "Not implemented function",
	"AUTH_ISSUE":                   "Authentication Issue",
	"USE_AFTER_FREE":               "Use After Free",
	"RESOURCE_HIERARCHY":           "Resource Hierarchy",
	"LEAKAGE":                      "Leakage",
	"INVALID_DYNAMIC_OBJECT":       "Invalid Dynamic Object",
	"PAYLOAD_BODY":                 "Payload Body",
	"PAYLOAD_BODY_NOT_IMPLEMENTED": "Not implemented function",
	"CRUD_NOT_ENOUGH_DATA":         "Not enough info for object fuzzing",
	"CRUD_DELETE_AGAIN":            "Access to deleted object detected by Fuzzer",
	"CRUD_GET_AFTER_DELETE":        "Access to deleted object detected by Fuzzer",
	"CRUD_LIFE_CYCLE":              "Fuzzer failed to process object",
	"CRUD_PUT_AFTER_DELETE":        "Access to deleted object detected by Fuzzer",
	"FUZZER_INTERNAL_ERROR":        "Fuzzer internal error",
}

const (
	AnnotationReportName      = "fuzzer_report"
	AnnotationFindingsName    = "fuzzer_findings"
	OneHundredPercentConstant = 100
)

type TestItem struct {
	Test              *restapi.TestWithReport
	ProvidedSpec      *models.SpecInfo
	ReconstructedSpec *models.SpecInfo
}

type API struct {
	ID        uint
	Name      string
	Port      uint
	Namespace string
	Fuzzed    bool
	InFuzzing bool
	TestsList []*TestItem
	//tests     []restapi.TestWithReport // List of tests as displayed on Tests Subtab
}

/*
* Factories.
 */

func NewAPI(id uint, name string, port uint, namespace string) API {
	return API{
		ID:        id,
		Name:      name,
		Port:      port,
		Namespace: namespace,
		Fuzzed:    false,
		InFuzzing: false,
		TestsList: []*TestItem{},
		//tests:     []restapi.TestWithReport{},
	}
}

func NewTest() *TestItem {
	// Create a new empty test struct with timestamp=Now
	now := time.Now()
	starttime := now.Unix()
	lastReportTime := now.Unix()
	return &TestItem{
		Test: &restapi.TestWithReport{
			Starttime:       &starttime,
			Progress:        new(int),
			Vulnerabilities: &(restapi.Vulnerabilities{Total: new(int), Critical: new(int), High: new(int), Medium: new(int), Low: new(int)}),
			LastReportTime:  &lastReportTime,
			ErrorMessage:    new(string),
		},
		ProvidedSpec:      &(models.SpecInfo{}),
		ReconstructedSpec: &(models.SpecInfo{}),
	}
}

/*
* API.
 */

func (api *API) GetLastStatus() *restapi.FuzzingStatusAndReport {
	if len(api.TestsList) > 0 {
		index := len(api.TestsList) - 1
		return api.TestsList[index].Test.Report
	}
	return nil
}

func (api *API) GetLastShortStatus() (*restapi.ShortTestReport, error) {
	if len(api.TestsList) > 0 {
		index := len(api.TestsList) - 1
		lastTest := api.TestsList[index].Test
		lastReport := lastTest.Report

		// Create the shortreport structure
		shortReport := restapi.ShortTestReport{
			Starttime:     *lastTest.Starttime,
			Status:        lastReport.Status,
			StatusMessage: lastTest.ErrorMessage,
			Tags:          []restapi.FuzzingReportTag{},
		}

		// Prepare on the shortreport structure the list of tags/operations from the provided spec content
		specInfo := api.TestsList[index].ProvidedSpec
		if specInfo.Tags != nil {
			for _, tag := range specInfo.Tags {
				logging.Logf("[Fuzzer] API(%v).StartFuzzing(): ... tag (%v)", api.ID, tag.Name)
				fuzzingReportTag := restapi.FuzzingReportTag{Name: tag.Name, Operations: []restapi.FuzzingReportOperation{}}
				for _, op := range tag.MethodAndPathList {
					logging.Logf("[Fuzzer] API(%v).StartFuzzing(): ... ... method %v %v", api.ID, op.Method, op.Path)
					fuzzingReportTag.Operations = append(fuzzingReportTag.Operations, restapi.FuzzingReportOperation{
						Operation: &common.MethodAndPath{
							Method: (*common.HttpMethod)(&op.Method),
							Path:   &op.Path,
						},
						RequestsCount: 0})
				}
				shortReport.Tags = append(shortReport.Tags, fuzzingReportTag)
			}
		}

		// Then iterate on the regular report items and verse it on the shortdemo structure
		for _, reportItem := range lastTest.Report.Report {
			if strings.HasPrefix(*reportItem.Name, "definitions:") {
				/*tokens := strings.Split(*reportItem.Name, ":")
				if len(tokens) > 1 {
					desc := fmt.Sprintf("Tests for the object '%v'", tokens[1])
					reportItem.Description = &desc
				}*/
				//Todo
			} else if strings.HasPrefix(*reportItem.Name, "path:") {
				tokens := strings.Split(*reportItem.Name, ":")
				if len(tokens) > 1 {
					//desc := fmt.Sprintf("Tests on path '%v'", tokens[1])
					//reportItem.Description = &desc
					opPath := tokens[1]
					for _, path := range *reportItem.Paths {
						// Report this path in shortreport
						err := updateRequestCounter(&shortReport, opPath, *path.Verb)
						if err != nil {
							panic(err)
						}
					}
				}
			} else if strings.HasPrefix(*reportItem.Name, "restler") {
				//desc := "Set of tests made automatically by Restler based on the specs"
				//reportItem.Description = &desc
				//Todo
			}
		}

		return &shortReport, nil
	}
	return nil, fmt.Errorf("No existing tests for api(%v)", api.ID)
}

func updateRequestCounter(shortReport *restapi.ShortTestReport, path string, verb string) error {
	for idx1 := range shortReport.Tags {
		tag := &shortReport.Tags[idx1]
		for idx2 := range tag.Operations {
			ops := &tag.Operations[idx2]
			if *ops.Operation.Path == path && *ops.Operation.Method == common.HttpMethod(verb) {
				ops.RequestsCount++
				return nil
			}
		}
	}
	// Not found
	logging.Errorf("[Fuzzer] Can't find operation(%v %v) in spec", verb, path)
	return nil
}

func (api *API) AddNewStatusReport(report restapi.FuzzingStatusAndReport) {
	if !api.InFuzzing {
		logging.Logf("[Fuzzer] AddNewStatusReport():: API id (%v) not in Fuzzing... did you triggered it from HTTP?", api.ID)
		return
	}

	// Logf("[Fuzzer] AddNewStatusReport():: api.inFuzzing=%v", api.inFuzzing)
	// Logf("[Fuzzer] AddNewStatusReport():: len(api.tests)=%v", len(api.tests))

	// Add report contet on test data for the said API
	if api.InFuzzing && len(api.TestsList) > 0 {
		now := time.Now().Unix()
		index := len(api.TestsList) - 1
		lastTest := api.TestsList[index].Test
		lastTest.Progress = &report.Progress
		lastTest.Report = &report
		lastTest.LastReportTime = &now

		// Update main vulnerabilities for the test
		total, critical, high, medium, low := 0, 0, 0, 0, 0
		for _, reportItem := range report.Report {
			for _, finding := range *reportItem.Findings {
				// update severity counters
				switch *finding.Request.Severity {
				case "critical":
					critical++
				case "high":
					high++
				case "medium":
					medium++
				case "low":
					low++
				}
			}
		}
		total = critical + high + medium + low
		lastTest.Vulnerabilities.Critical = &critical
		lastTest.Vulnerabilities.High = &high
		lastTest.Vulnerabilities.Medium = &medium
		lastTest.Vulnerabilities.Low = &low
		lastTest.Vulnerabilities.Total = &total

		// If restler data on report, format is on form:
		// "paths": [],
		// "findings": [
		//  	...
		// ]
		// extract paths from findings.additionalInfo param
		for _, reportItem := range report.Report {
			if *reportItem.Name == "restler" && *reportItem.Source == "RESTLER" {
				for _, finding := range *reportItem.Findings {
					tokens := strings.Split(*finding.AdditionalInfo, " ")
					// logging.Logf("[Fuzzer] AddNewStatusReport():: #### AdditionalInfo=%v", *finding.AdditionalInfo)
					if len(tokens) > 3 && strings.HasPrefix(tokens[2], "HTTP") {
						httpcode := tools.GetHTTPCodeFromFindingType(*finding.Type)
						*reportItem.Paths = append(*reportItem.Paths, tools.NewFuzzingReportPath(httpcode, tokens[0], tokens[1]))
						// logging.Logf("[Fuzzer] AddNewStatusReport():: #### ... add new path len(api.tests)=%v", (*reportItem.Paths)[len(*reportItem.Paths)-1])
					}
				}
				// It exists only one ""
				break
			}
		}

		// fill description
		for key, reportItem := range report.Report {
			if strings.HasPrefix(*reportItem.Name, "definitions:") {
				tokens := strings.Split(*reportItem.Name, ":")
				if len(tokens) > 1 {
					desc := fmt.Sprintf("Tests for the object '%v'", tokens[1])
					reportItem.Description = &desc
				}
			} else if strings.HasPrefix(*reportItem.Name, "path:") {
				tokens := strings.Split(*reportItem.Name, ":")
				if len(tokens) > 1 {
					desc := fmt.Sprintf("Tests on path '%v'", tokens[1])
					reportItem.Description = &desc
				}
			} else if strings.HasPrefix(*reportItem.Name, "restler") {
				desc := "Set of tests made automatically by Restler based on the specs"
				reportItem.Description = &desc
			}
			// Save the update
			report.Report[key] = reportItem
		}
	}
}

func (api *API) AddErrorOnLastTest(fuzzerError error) {
	if len(api.TestsList) > 0 {
		errorMessage := fuzzerError.Error()
		index := len(api.TestsList) - 1
		lastTest := api.TestsList[index].Test
		lastTest.ErrorMessage = &(errorMessage)
		if lastTest.Report == nil {
			report := NewReport()
			lastTest.Report = &report
		}
		report := lastTest.Report
		report.Progress = 100
		report.Status = restapi.ERROR
	}
}

func (api *API) GetTestContent(timestamp int64) *restapi.TestWithReport {
	for _, testItem := range api.TestsList {
		if *testItem.Test.Starttime == timestamp {
			return testItem.Test
		}
	}
	return nil
}

// Return a list of tests with reduced informations.
func (api *API) GetTestsList() *[]restapi.Test {
	var testList []restapi.Test
	if api.InFuzzing && len(api.TestsList) > 0 {
		index := len(api.TestsList) - 1
		lastTestItem := api.TestsList[index].Test
		currentTime := time.Now().Unix()
		secondsSinceLastReport := (currentTime - *lastTestItem.LastReportTime)
		if secondsSinceLastReport > int64(config.GetConfig().GetTestReportTimeout()) {
			// This can be an issue
			fuzzerError := fmt.Errorf("a timeout occurred: it seems we can't receive response from Fuzzer workload")
			err := api.StopFuzzing(fuzzerError)
			if err != nil {
				logging.Errorf("[Fuzzer] API(%v).GetTestsList(): error occurred when trying to stop fuzzing, err=%v", api.ID, err)
			}
		}
	}
	for _, testItem := range api.TestsList {
		testItem := CopyTestFromTestWithReport(*testItem.Test)
		testList = append(testList, testItem)
	}
	return &(testList)
}

func (api *API) GetLastFindings() *[]restapi.Finding {
	var findingList []restapi.Finding

	if len(api.TestsList) > 0 {
		index := len(api.TestsList) - 1
		lastTestItem := api.TestsList[index].Test
		if lastTestItem.Report != nil {
			for _, reportItem := range lastTestItem.Report.Report {
				for _, finding := range *reportItem.Findings {
					findingName := typeToNameMap[*finding.Type]
					findingDescription := ""
					if finding.Description != nil {
						findingDescription = *finding.Description
					}
					risk := *(finding.Request.Severity)
					findingList = append(findingList, NewFinding(findingName, findingDescription, risk))
				}
			}
		}
	}

	return &(findingList)
}

func (api *API) ForceProgressForLastTest(progress int) error {
	if len(api.TestsList) > 0 {
		index := len(api.TestsList) - 1
		lastTestItem := api.TestsList[index].Test
		lastTestItem.Progress = &progress
	}
	return nil
}

func (api *API) StartFuzzing(specsInfo *models.OpenAPISpecs) error {
	logging.Logf("[Fuzzer] API(%v).StartFuzzing(): Start fuzzing", api.ID)
	if api.InFuzzing {
		logging.Errorf("[Fuzzer] API(%v).StartFuzzing(): A fuzzing is already started", api.ID)
		return fmt.Errorf("a fuzzing is already started for api(%v)", api.ID)
	}
	api.InFuzzing = true
	// Add a new Test item with progress 0% and No report
	testItem := NewTest()
	testItem.ProvidedSpec = specsInfo.ProvidedSpec
	testItem.ReconstructedSpec = specsInfo.ReconstructedSpec
	/*logging.Logf("[Fuzzer] API(%v).StartFuzzing(): specsInfo.ProvidedSpec=(%v)", api.Id, specsInfo.ProvidedSpec)
	for _, tag := range specsInfo.ProvidedSpec.Tags {
		logging.Logf("[Fuzzer] API(%v).StartFuzzing(): ... tag (%v)", api.Id, tag.Name)
		for _, op := range tag.MethodAndPathList {
			logging.Logf("[Fuzzer] API(%v).StartFuzzing(): ... ... method %v %v", api.Id, op.Method, op.Path)
		}
	}*/
	api.TestsList = append(api.TestsList, testItem)
	return nil
}

func (api *API) StopFuzzing(fuzzerError error) error {
	logging.Logf("[Fuzzer] API(%v).StopFuzzing(): Stop fuzzing, with error(%v)", api.ID, fuzzerError)
	api.InFuzzing = false
	api.Fuzzed = true
	// Force the last test progress to 100%
	err := api.ForceProgressForLastTest(OneHundredPercentConstant)
	if fuzzerError != nil {
		api.AddErrorOnLastTest(fuzzerError)
	}
	if err != nil {
		log.Fatalln(err)
		return fmt.Errorf("can't set the progress status for last test of api (%v)", api.ID)
	}
	return nil
}

func (api *API) StoreReportData(ctx context.Context, accessor core.BackendAccessor, moduleName string, data restapi.FuzzingStatusAndReport) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("can't decode report data for api(%v), err=%v", api.ID, err.Error())
	}
	err = accessor.StoreAPIInfoAnnotations(ctx, moduleName, api.ID, core.Annotation{Name: AnnotationReportName, Annotation: bytes})
	if err != nil {
		return fmt.Errorf("can't store report data for api(%v), err=%v", api.ID, err.Error())
	}
	return nil
}

func (api *API) StoreLastFindingsData(ctx context.Context, accessor core.BackendAccessor, moduleName string, data []byte) error {
	err := accessor.StoreAPIInfoAnnotations(ctx, moduleName, api.ID, core.Annotation{Name: AnnotationFindingsName, Annotation: data})
	if err != nil {
		return fmt.Errorf("can't store report data for api(%v), err=%v", api.ID, err.Error())
	}
	return nil
}

func (api *API) RetrieveInfoFromStore(ctx context.Context, accessor core.BackendAccessor, moduleName string) error {
	dbAnns, err := accessor.ListAPIInfoAnnotations(ctx, moduleName, api.ID)
	if err != nil {
		return fmt.Errorf("can't retrieve annotation for api(%v), err=%v", api.ID, err.Error())
	}
	for _, annotation := range dbAnns {
		if annotation.Name == AnnotationReportName {
			logging.Logf("[Fuzzer] API(%v).RetrieveInfoFromStore(): Found Annotation Name=(%v), size=(%v)", api.ID, annotation.Name, len(annotation.Annotation))
			var data restapi.FuzzingStatusAndReport
			err = json.Unmarshal(annotation.Annotation, &data)
			if err != nil {
				logging.Errorf("[Fuzzer] API(%v).RetrieveInfoFromStore(): failed to decode the annotation body, error=%v", api.ID, err)
				break
			}
			// Before ingest any report, we must be "in fuzzing" mode
			api.InFuzzing = true
			if len(api.TestsList) == 0 {
				// Add the report in a new test
				api.TestsList = append(api.TestsList, NewTest())
				api.AddNewStatusReport(data)
			}
			api.AddNewStatusReport(data)
			api.InFuzzing = false
		}
		if annotation.Name == "Fuzzer report" || annotation.Name == "Fuzzer findings" {
			logging.Logf("[Fuzzer] API(%v).RetrieveInfoFromStore(): Found Annotation Name=(%v), size=(%v)", api.ID, annotation.Name, len(annotation.Annotation))
			// Nothing to do for now, we don't use it
		}
	}
	return nil
}
