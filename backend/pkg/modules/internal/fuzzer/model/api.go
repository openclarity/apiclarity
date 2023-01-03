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
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	logging "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/config"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/restapi"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/tools"
	"github.com/openclarity/apiclarity/backend/pkg/modules/utils"
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
	ReportNameCRUDPrefix      = "definitions:"
	ReportNameSCNFuzzerPrefix = "path:"
	ReportNameRestlerPrefix   = "restler"
	MinLocationTokensNumber   = 4
	DefaultErrorMsg           = "An error occurred during the test"
)

/*
* A TestItem will link a report with the corresponding spec that generate the report.
 */
type TestItem struct {
	Test      *restapi.TestWithReport
	SpecsInfo *tools.FuzzerSpecsInfo
}

type API struct {
	ID        uint
	Name      string
	Port      uint
	Namespace string
	Fuzzed    bool
	InFuzzing bool
	TestsList []*TestItem
}

/*
* Factories
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
			Report: &(restapi.FuzzingStatusAndReport{
				Progress: 0,
				Report:   map[string]restapi.FuzzingReportItem{},
				Status:   restapi.INPROGRESS,
			}),
		},
		SpecsInfo: &(tools.FuzzerSpecsInfo{}),
	}
}

func ConvertRawFindingToAPIFinding(finding restapi.RawFindings) *common.APIFinding {
	var additionalInfo map[string]interface{}

	if finding.AdditionalInfo != nil {
		err := json.Unmarshal([]byte(*finding.AdditionalInfo), &additionalInfo)
		if err != nil {
			additionalInfo = map[string]interface{}{
				"Details": finding.AdditionalInfo,
			}
		}
	}
	result := common.APIFinding{
		Type:                 *finding.Type,
		Name:                 typeToNameMap[*finding.Type],
		Source:               *finding.Namespace,
		Description:          *finding.Description,
		Severity:             convertSeverity(*finding.Request.Severity),
		AdditionalInfo:       &additionalInfo,
		ProvidedSpecLocation: nil,
	}
	return &result
}

func convertSeverity(severity string) common.Severity {
	switch strings.ToLower(severity) {
	case "critical":
		return common.CRITICAL
	case "high":
		return common.HIGH
	case "medium":
		return common.MEDIUM
	case "low":
		return common.LOW
	case "info":
		return common.INFO
	default:
		logging.Warningf("[Fuzzer] unexpected severity level (%s) using info.", strings.ToLower(severity))
		return common.INFO
	}
}

/*
* API
 */

func (api *API) GetLastReport() *restapi.FuzzingStatusAndReport {
	if len(api.TestsList) > 0 {
		index := len(api.TestsList) - 1
		return api.TestsList[index].Test.Report
	}
	return nil
}

func (api *API) GetLastStatus() (restapi.FuzzingStatusEnum, error) {
	if len(api.TestsList) > 0 {
		index := len(api.TestsList) - 1
		report := api.TestsList[index].Test.Report
		if report == nil {
			// Must not happen as we create a default empty report at test creation
			return restapi.ERROR, fmt.Errorf("no report for last test for the api")
		}
		return report.Status, nil
	}
	return restapi.ERROR, fmt.Errorf("no test for the api")
}

func (api *API) SetErrorForLastStatus(msg string) error {
	if len(api.TestsList) > 0 {
		index := len(api.TestsList) - 1
		lastTest := api.TestsList[index].Test
		report := lastTest.Report
		if report == nil {
			// Must not happen as we create a default empty report at test creation
			return fmt.Errorf("no report for last test for the api")
		}
		report.Status = restapi.ERROR
		lastTest.ErrorMessage = &msg
		return nil
	}
	return fmt.Errorf("no test available")
}

func (api *API) GetLastStatusError() (string, error) {
	if len(api.TestsList) > 0 {
		index := len(api.TestsList) - 1
		lastTest := api.TestsList[index].Test
		return *lastTest.ErrorMessage, nil
	}
	return "", fmt.Errorf("no test available")
}

func (api *API) GetShortStatusOnError(msg string) (*restapi.ShortTestReport, error) {
	// Retrieve a ShortStatus, even if there is no valid report, with error status inside
	// Use by default msg as error message. If empty, try to use lastTest.ErrorMessage is any
	if len(api.TestsList) > 0 {
		index := len(api.TestsList) - 1
		lastTest := api.TestsList[index].Test

		msgToSend := msg
		if len(msgToSend) == 0 && lastTest.ErrorMessage != nil && len(*lastTest.ErrorMessage) > 0 {
			msgToSend = *lastTest.ErrorMessage
		}

		// Create the shortreport structure to fill
		shortReport := restapi.ShortTestReport{
			Starttime:     *lastTest.Starttime,
			Status:        restapi.ERROR,
			StatusMessage: &msgToSend,
			Tags:          &[]restapi.FuzzingReportTag{},
		}
		return &shortReport, nil
	}
	return nil, fmt.Errorf("no existing tests for api(%v)", api.ID)
}

func (api *API) GetTestShortReportByTimestamp(timestamp int64) (*restapi.ShortTestReport, error) {
	for _, testItem := range api.TestsList {
		if *testItem.Test.Starttime == timestamp {
			return api.getTestShortReport(testItem)
		}
	}
	return nil, fmt.Errorf("no existing test for api(%v) with timestamp=%v", api.ID, timestamp)
}

func (api *API) getTestShortReport(testItem *TestItem) (*restapi.ShortTestReport, error) {
	test := testItem.Test
	report := test.Report

	if report == nil {
		// No available report yet
		return nil, fmt.Errorf("no report yet")
	}

	// Create the shortreport structure to fill
	shortReport := restapi.ShortTestReport{
		Starttime:     *test.Starttime,
		Status:        report.Status,
		StatusMessage: test.ErrorMessage,
		Tags:          &[]restapi.FuzzingReportTag{},
	}

	// Get current spec informations
	var specInfo *models.SpecInfo
	logging.Debugf("[Fuzzer] API(%v).GetLastShortStatus(): specInfo Provided(len=%v), Reconstructed(len=%v)", api.ID, len(testItem.SpecsInfo.ProvidedSpec), len(testItem.SpecsInfo.ReconstructedSpec))
	if testItem.SpecsInfo.ProvidedSpec != "" {
		specInfo = testItem.SpecsInfo.ProvidedSpecInfo
	} else if testItem.SpecsInfo.ReconstructedSpec != "" {
		specInfo = testItem.SpecsInfo.ReconstructedSpecInfo
	} else {
		return nil, fmt.Errorf("no spec information")
	}

	// Prepare on the shortreport structure the list of tags/operations from the spec content
	if specInfo.Tags != nil {
		for _, tag := range specInfo.Tags {
			// logging.Debugf("[Fuzzer] API(%v).GetLastShortStatus(): ... tag (%v)", api.ID, tag.Name)
			fuzzingReportTag := restapi.FuzzingReportTag{
				Name:            tag.Name,
				Operations:      []restapi.FuzzingReportOperation{},
				HighestSeverity: nil,
			}
			for _, op := range tag.MethodAndPathList {
				// logging.Debugf("[Fuzzer] API(%v).GetLastShortStatus(): ... ... method %v %v", api.ID, op.Method, op.Path)
				fuzzingReportTag.Operations = append(fuzzingReportTag.Operations, restapi.FuzzingReportOperation{
					Operation: common.MethodAndPath{
						Method: (*common.HttpMethod)(&op.Method),
						Path:   &op.Path,
					},
					RequestsCount:   0,
					Findings:        &[]common.APIFinding{},
					HighestSeverity: nil,
				})
			}
			*shortReport.Tags = append(*shortReport.Tags, fuzzingReportTag)
		}
	} else {
		return nil, fmt.Errorf("invalid or no existing spec content")
	}

	// Then iterate on the regular report items and verse it on the shortdemo structure
	for _, reportItem := range report.Report {
		if strings.HasPrefix(*reportItem.Name, ReportNameCRUDPrefix) {
			// Come from the 'crud' fuzzer
			// TODO
		} else if strings.HasPrefix(*reportItem.Name, ReportNameSCNFuzzerPrefix) {
			// Come from the 'scn-fuzzer' fuzzer
			tokens := strings.Split(*reportItem.Name, ":")
			if len(tokens) > 1 {
				opPath := tokens[1]
				for _, path := range *reportItem.Paths {
					// Report this path in shortreport
					err := updateRequestCounter(&shortReport, opPath, *path.Verb)
					if err != nil {
						// The error has been already logged, then simply skip the current request
						continue
					}
				}
			}
		} else if strings.HasPrefix(*reportItem.Name, ReportNameRestlerPrefix) {
			// The set of tests made automatically by Restler based on the specs
			err := updateRequestCountersForRestler(&shortReport, reportItem, testItem.SpecsInfo.ProvidedSpec)
			if err != nil {
				// The error has been already logged, then simply skip the current report item
				continue
			}
		}
	}

	// Then redo the same for findings (I know, it can be done on the loop above, but I prefer to separate the job)
	for _, reportItem := range report.Report {
		for _, finding := range *reportItem.Findings {
			// finding.Location is something like &[OASv3Spec paths /user/logout get]
			if len(*finding.Location) < MinLocationTokensNumber {
				logging.Errorf("[Fuzzer] API(%v).GetLastShortStatus(): Found an invalid finding location (%v)", api.ID, finding.Location)
				continue
			}
			verb := (*finding.Location)[3]
			method := (*finding.Location)[2]
			verb = strings.ToUpper(verb)
			err := AddFindingOnShortReport(&shortReport, method, verb, finding)
			if err != nil {
				// Log an error, but we continue the process (not blocking)
				logging.Errorf("[Fuzzer] API(%v).GetLastShortStatus(): can't add finding on report (%v, %v, %v), err=(%v)", api.ID, verb, method, finding, err)
			}
		}
	}

	// Remove Tags item if no tags are present
	if shortReport.Tags != nil && len(*shortReport.Tags) == 0 {
		shortReport.Tags = nil
	}

	return &shortReport, nil
}

func (api *API) GetLastShortStatus() (*restapi.ShortTestReport, error) {
	if len(api.TestsList) > 0 {
		index := len(api.TestsList) - 1
		shortReport, err := api.getTestShortReport(api.TestsList[index])
		if err != nil {
			return nil, fmt.Errorf("unable to fetch short report from last test for api(%v): %v", api.ID, err)
		}
		return shortReport, nil
	}
	return nil, fmt.Errorf("no existing tests for api(%v)", api.ID)
}

func updateRequestCounter(shortReport *restapi.ShortTestReport, path string, verb string) error {
	if *shortReport.Tags == nil {
		// No tags, then no operations and no request counter to update
		return nil
	}

	for idx1 := range *shortReport.Tags {
		tag := &(*shortReport.Tags)[idx1]
		for idx2 := range tag.Operations {
			ops := &tag.Operations[idx2]
			if *ops.Operation.Path == path && *ops.Operation.Method == common.HttpMethod(verb) {
				ops.RequestsCount++
				return nil
			}
		}
	}
	// Not found
	logging.Errorf("[Fuzzer] Can't find operation(%v %v) in spec to update requests counter", verb, path)
	return fmt.Errorf("can't find operation(%v %v) in spec to update requests counter", verb, path)
}

func getLocation(path string, method string) *string {
	s := utils.JSONPointer("paths", path, strings.ToLower(method))
	return &s
}

func AddFindingOnShortReport(shortReport *restapi.ShortTestReport, path string, verb string, finding restapi.RawFindings) error {
	if *shortReport.Tags == nil {
		// No tags, then no operations on which to add findings
		return nil
	}

	for idx1 := range *shortReport.Tags {
		tag := &(*shortReport.Tags)[idx1]
		for idx2 := range tag.Operations {
			ops := &tag.Operations[idx2]
			if *ops.Operation.Path == path && *ops.Operation.Method == common.HttpMethod(verb) {
				// Add the finding
				commonFinding := ConvertRawFindingToAPIFinding(finding)
				commonFinding.ProvidedSpecLocation = getLocation(path, verb)
				*ops.Findings = append(*ops.Findings, *commonFinding)
				// Update higestSeverity for operation
				if ops.HighestSeverity == nil || tools.IsGreaterSeverity(commonFinding.Severity, *ops.HighestSeverity) {
					ops.HighestSeverity = &commonFinding.Severity
					// Check for higestSeverity at tags level
					if tag.HighestSeverity == nil || tools.IsGreaterSeverity(commonFinding.Severity, *tag.HighestSeverity) {
						tag.HighestSeverity = &commonFinding.Severity
						// Lastly... test at report level
						if shortReport.HighestSeverity == nil || tools.IsGreaterSeverity(commonFinding.Severity, *shortReport.HighestSeverity) {
							shortReport.HighestSeverity = &commonFinding.Severity
						}
					}
				}
				return nil
			}
		}
	}
	// Not found
	logging.Errorf("[Fuzzer] Can't find operation(%v %v) in spec to store the finding", verb, path)
	return fmt.Errorf("can't find operation(%v %v) in spec to store the finding", verb, path)
}

func updateRequestCountersForRestler(shortReport *restapi.ShortTestReport, reportItem restapi.FuzzingReportItem, spec string) error {
	if *shortReport.Tags == nil {
		// No tags, then no operations and no request counter to update
		return nil
	}

	logging.Debugf("[Fuzzer] updateRequestCountersForRestler(): spec len=(%v)", len(spec))
	doc, err := tools.LoadSpec([]byte(spec))
	if err != nil {
		logging.Errorf("[Fuzzer] updateRequestCountersForRestler(): Invalid Spec err=(%v)", err)
		return fmt.Errorf("invalid Spec")
	}

	// Find basepaths from servers list, then save it before reset
	basePaths := tools.GetBasePathsFromServers(&doc.Servers)
	logging.Debugf("[Fuzzer] updateRequestCountersForRestler(): basePaths (%v)", basePaths)
	doc.Servers = openapi3.Servers{}

	// Create the router
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return fmt.Errorf("can't create router, err=(%v)", err)
	}

	for _, path := range *reportItem.Paths {
		// Patch for Fuzzer improper verb
		verb := *path.Verb
		if verb[0:1] == "'" {
			verb = strings.TrimPrefix(verb, "'")
		}

		URIsToTest := []string{}
		URIsToTest = append(URIsToTest, *path.Uri)
		for _, basepath := range basePaths {
			if strings.HasPrefix(*path.Uri, basepath) {
				URIsToTest = append(URIsToTest, strings.TrimPrefix(*path.Uri, basepath))
			}
		}
		// logging.Debugf("[Fuzzer] updateRequestCountersForRestler(): process paths (%v %v)", verb, URIsToTest)
		for _, uri := range URIsToTest {
			route, err := tools.FindRoute(&router, verb, uri)
			if err != nil {
				// Not an error, that can occurs, specially when we manage some basepath. Simply skip it.
				logging.Debugf("[Fuzzer] updateRequestCountersForRestler(): ... can't find it err=(%v)", err)
				continue
			}
			err = updateRequestCounter(shortReport, route.Path, route.Method)
			if err != nil {
				logging.Errorf("[Fuzzer] updateRequestCountersForRestler(): can't update request counter err=(%v)", err)
			}
			break
		}
	}

	return nil
}

func (api *API) AddNewStatusReport(report restapi.FuzzingStatusAndReport) error {
	if !api.InFuzzing {
		logging.Infof("[Fuzzer] AddNewStatusReport():: API id (%v) not in Fuzzing... did you triggered it from HTTP?", api.ID)
		return fmt.Errorf("API not in fuzzing")
	}

	logging.Debugf("[Fuzzer] AddNewStatusReport():: Status inFuzzing=(%v), nb of tests=(%v)", api.InFuzzing, len(api.TestsList))

	// Add report contet on test data for the said API
	if api.InFuzzing && len(api.TestsList) > 0 {
		now := time.Now().Unix()
		index := len(api.TestsList) - 1
		lastTest := api.TestsList[index].Test
		lastTest.Progress = &report.Progress
		lastTest.Report = &report
		lastTest.LastReportTime = &now

		if report.Status == restapi.ERROR {
			// Put a default message here. Must be updated when Fuzzer will be able to return a proper error message
			*lastTest.ErrorMessage = DefaultErrorMsg
		}

		// Update main vulnerabilities for the test
		total, critical, high, medium, low := 0, 0, 0, 0, 0
		for _, reportItem := range report.Report {
			for _, finding := range *reportItem.Findings {
				// update severity counters
				switch convertSeverity(*finding.Request.Severity) {
				case common.CRITICAL:
					critical++
				case common.HIGH:
					high++
				case common.MEDIUM:
					medium++
				case common.LOW:
					low++
				case common.INFO:
					// Nothing
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
					if len(tokens) > 3 && strings.HasPrefix(tokens[2], "HTTP") {
						logging.Debugf("[Fuzzer] AddNewStatusReport():: Adding new report item (%v %v)", tokens[0], tokens[1])
						httpcode := tools.GetHTTPCodeFromFindingType(*finding.Type)
						*reportItem.Paths = append(*reportItem.Paths, tools.NewFuzzingReportPath(httpcode, tokens[0], tokens[1]))
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
	return nil
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

func (api *API) GetTestByTimestamp(timestamp int64) *restapi.TestWithReport {
	for _, testItem := range api.TestsList {
		if *testItem.Test.Starttime == timestamp {
			return testItem.Test
		}
	}
	return nil
}

func (api *API) GetLastTest() *restapi.TestWithReport {
	if len(api.TestsList) > 0 {
		index := len(api.TestsList) - 1
		return api.TestsList[index].Test
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

func (api *API) GetLastAPIFindings() *[]common.APIFinding {
	var findingList []common.APIFinding

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
					additionalInfo := map[string]interface{}{
						"Description": finding.AdditionalInfo,
					}
					verb := (*finding.Location)[3]
					path := (*finding.Location)[2]
					APIFinding := common.APIFinding{
						AdditionalInfo:            &additionalInfo,
						Description:               findingDescription,
						Name:                      findingName,
						ProvidedSpecLocation:      getLocation(path, verb),
						ReconstructedSpecLocation: new(string),
						Severity:                  convertSeverity(risk),
						Source:                    *finding.Namespace,
						Type:                      *finding.Type,
					}
					findingList = append(findingList, APIFinding)
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

func (api *API) StartFuzzing(params *FuzzingInput) (FuzzingTimestamp, error) {
	logging.Infof("[Fuzzer] API(%v).StartFuzzing(): Start fuzzing", api.ID)
	if api.InFuzzing {
		logging.Errorf("[Fuzzer] API(%v).StartFuzzing(): A fuzzing is already started", api.ID)
		return ZeroTime, fmt.Errorf("a fuzzing is already started for api(%v)", api.ID)
	}
	api.InFuzzing = true
	// Add a new Test item with progress 0% and No report
	testItem := NewTest()
	testItem.SpecsInfo = params.SpecsInfo
	api.TestsList = append(api.TestsList, testItem)
	return *testItem.Test.Starttime, nil
}

func (api *API) StopFuzzing(fuzzerError error) error {
	if fuzzerError != nil {
		logging.Infof("[Fuzzer] API(%v).StopFuzzing(): Stop fuzzing, with error(%v)", api.ID, fuzzerError)
	} else {
		logging.Infof("[Fuzzer] API(%v).StopFuzzing(): Stop fuzzing", api.ID)
	}
	api.InFuzzing = false
	api.Fuzzed = true
	// Force the last test progress to 100%
	err := api.ForceProgressForLastTest(OneHundredPercentConstant)
	if fuzzerError != nil {
		api.AddErrorOnLastTest(fuzzerError)
	}
	if err != nil {
		return fmt.Errorf("can't set the progress status for last test of api (%v)", api.ID)
	}
	return nil
}

func (api *API) StoreReportData(ctx context.Context, accessor core.BackendAccessor, moduleName string, data restapi.FuzzingStatusAndReport) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("can't decode report data for api(%v), err=%v", api.ID, err)
	}
	err = accessor.StoreAPIInfoAnnotations(ctx, moduleName, api.ID, core.Annotation{Name: AnnotationReportName, Annotation: bytes})
	if err != nil {
		return fmt.Errorf("can't store report data for api(%v), err=%v", api.ID, err)
	}
	return nil
}

func (api *API) StoreLastFindingsData(ctx context.Context, accessor core.BackendAccessor, moduleName string, data []byte) error {
	err := accessor.StoreAPIInfoAnnotations(ctx, moduleName, api.ID, core.Annotation{Name: AnnotationFindingsName, Annotation: data})
	if err != nil {
		return fmt.Errorf("can't store report data for api(%v), err=%v", api.ID, err)
	}
	return nil
}

func (api *API) RetrieveInfoFromStore(ctx context.Context, accessor core.BackendAccessor, moduleName string) error {
	dbAnns, err := accessor.ListAPIInfoAnnotations(ctx, moduleName, api.ID)
	if err != nil {
		return fmt.Errorf("can't retrieve annotation for api(%v), err=%v", api.ID, err)
	}
	for _, annotation := range dbAnns {
		if annotation.Name == AnnotationReportName {
			logging.Infof("[Fuzzer] API(%v).RetrieveInfoFromStore(): Found Annotation Name=(%v), size=(%v)", api.ID, annotation.Name, len(annotation.Annotation))
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
			}
			err := api.AddNewStatusReport(data)
			if err != nil {
				logging.Errorf("[Fuzzer] API(%v).RetrieveInfoFromStore(): failed to add new status report, error=(%v)", api.ID, err)
			}
			api.InFuzzing = false
		}
		if annotation.Name == "Fuzzer report" || annotation.Name == "Fuzzer findings" {
			logging.Infof("[Fuzzer] API(%v).RetrieveInfoFromStore(): Found Annotation Name=(%v), size=(%v)", api.ID, annotation.Name, len(annotation.Annotation))
			// Nothing to do for now, we don't use it
		}
	}
	return nil
}
