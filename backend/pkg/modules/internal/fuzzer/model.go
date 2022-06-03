package fuzzer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/logging"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/restapi"
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
}

type API struct {
	id        uint
	name      string
	port      uint
	namespace string
	fuzzed    bool
	inFuzzing bool
	tests     []restapi.TestWithReport // List of tests as displayed on Tests Subtab
}

type Model struct {
	db       []API
	accessor core.BackendAccessor
}

/*
* Factories
 */

func NewAPI(id uint, name string, port uint, namespace string) API {
	return API{
		id:        id,
		name:      name,
		port:      port,
		namespace: namespace,
		fuzzed:    false,
		inFuzzing: false,
		tests:     []restapi.TestWithReport{},
	}
}

func NewTest() restapi.TestWithReport {
	now := time.Now()
	starttime := now.Unix()
	progress, zero := 0, 0
	return restapi.TestWithReport{
		Starttime:       &starttime,
		Progress:        &progress,
		Vulnerabilities: &(restapi.Vulnerabilities{Total: &zero, Critical: &zero, High: &zero, Medium: &zero, Low: &zero}),
	}
}

func CopyTestFromTestWithReport(fullTest restapi.TestWithReport) restapi.Test {
	return restapi.Test{
		Starttime:       fullTest.Starttime,
		Progress:        fullTest.Progress,
		Vulnerabilities: fullTest.Vulnerabilities,
	}
}

func NewFinding(name string, description string, risk string) restapi.Finding {
	return restapi.Finding{
		Name:        &name,
		Description: &description,
		Risk:        &risk,
	}
}

func (api *API) GetLastStatus() *restapi.FuzzingStatusAndReport {
	if len(api.tests) > 0 {
		index := len(api.tests) - 1
		return api.tests[index].Report
	}
	return nil
}

func (api *API) AddNewStatusReport(report restapi.FuzzingStatusAndReport) {
	if !api.inFuzzing {
		logging.Logf("[Fuzzer] AddNewStatusReport():: API id (%v) not in Fuzzing... did you triggered it from HTTP?", api.id)
		// Debug only, must not happen in production
		api.tests = append(api.tests, NewTest())
		api.inFuzzing = true
	}

	// Logf("[Fuzzer] AddNewStatusReport():: api.inFuzzing=%v", api.inFuzzing)
	// Logf("[Fuzzer] AddNewStatusReport():: len(api.tests)=%v", len(api.tests))
	// Add vulnerabilities data
	if api.inFuzzing && len(api.tests) > 0 {
		index := len(api.tests) - 1
		api.tests[index].Progress = &report.Progress
		api.tests[index].Report = &report

		// Update main vulnerabilities for the test
		total, critical, high, medium, low := 0, 0, 0, 0, 0
		for _, reportItem := range report.Report.AdditionalProperties {
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
		api.tests[index].Vulnerabilities.Critical = &critical
		api.tests[index].Vulnerabilities.High = &high
		api.tests[index].Vulnerabilities.Medium = &medium
		api.tests[index].Vulnerabilities.Low = &low
		api.tests[index].Vulnerabilities.Total = &total

		// If restler data on report, format is on form:
		// "paths": [],
		// "findings": [
		//  	...
		// ]
		// extract paths from findings.additionalInfo param
		for _, reportItem := range report.Report.AdditionalProperties {
			if *reportItem.Name == "restler" && *reportItem.Source == "RESTLER" {
				for _, finding := range *reportItem.Findings {
					tokens := strings.Split(*finding.AdditionalInfo, " ")
					// Logf("[Fuzzer] AddNewStatusReport():: #### AdditionalInfo=%v", *finding.AdditionalInfo)
					if len(tokens) > 3 && strings.HasPrefix(tokens[2], "HTTP") {
						httpcode := GetHTTPCodeFromFindingType(*finding.Type)
						*reportItem.Paths = append(*reportItem.Paths, NewFuzzingReportPath(httpcode, tokens[0], tokens[1]))
						// Logf("[Fuzzer] AddNewStatusReport():: #### ... add new path len(api.tests)=%v", (*reportItem.Paths)[len(*reportItem.Paths)-1])
					}
				}
				// It exists only one ""
				break
			}
		}

		// fill description
		for key, reportItem := range report.Report.AdditionalProperties {
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
			report.Report.Set(key, reportItem)
		}
	}
}

func (api *API) GetReport(timestamp int64) *restapi.TestWithReport {
	for _, test := range api.tests {
		if *test.Starttime == timestamp {
			return &test
		}
	}
	return nil
}

func (api *API) GetTestsList() *[]restapi.Test {
	var testList []restapi.Test
	for _, testWithReport := range api.tests {
		testList = append(testList, CopyTestFromTestWithReport(testWithReport))
	}
	return &(testList)
}

func (api *API) GetLastFindings() *[]restapi.Finding {
	var findingList []restapi.Finding

	if len(api.tests) > 0 {
		index := len(api.tests) - 1
		for _, reportItem := range api.tests[index].Report.Report.AdditionalProperties {
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

	return &(findingList)
}

func (m *Model) Init(ctx context.Context) error {
	// Nothing to do, for now
	return nil
}

func (m *Model) AddAPI(id uint, name string, port uint, service string) *API {
	api := NewAPI(id, name, port, service)
	logging.Logf("[model.AddAPI] Add new API %v", api)
	m.db = append(m.db, api)
	return &m.db[len(m.db)-1]
}

func (m *Model) GetAPI(ctx context.Context, apiID uint) (*API, error) {
	/*
	* Try to retrieve it from the cache
	 */
	for index, api := range m.db {
		if api.id == apiID {
			return &m.db[index], nil
		}
	}

	logging.Logf("[model.GetAPI] API %v not found, try to retrieve it from backend", apiID)

	/*
	* Try to retrieve it from backend
	 */
	apiInfo, err := m.accessor.GetAPIInfo(ctx, apiID)
	logging.Logf("[model.GetAPI(%v)]: get apiInfo=(%v)", apiID, apiInfo)
	if err != nil {
		log.Fatalln(err)
		return nil, fmt.Errorf("Error when retrieve api %v: %v", apiID, err)
	}

	// TODO: replace with apiInfo.DestinationNamespace when available
	newAPI := NewAPI(apiInfo.ID, apiInfo.Name, uint(apiInfo.Port), "")
	logging.Logf("[model.AddAPI] Add new API %v", newAPI)
	m.db = append(m.db, newAPI)
	return &m.db[len(m.db)-1], nil
}

func (m *Model) AddAPITest(apiID uint, message string) error {
	return nil
}

func (m *Model) StartAPIFuzzing(apiID uint) error {
	for index, api := range m.db {
		if api.id == apiID {
			logging.Logf("[Fuzzer] Model.StartAPIFuzzing(): Start API fuzzing for api %v", api)
			if m.db[index].inFuzzing {
				logging.Errorf("[Fuzzer] Model.StartAPIFuzzing(): A fuzzing is already started for api %v", api)
				return fmt.Errorf("a fuzzing is already started for: %v", apiID)
			}
			m.db[index].inFuzzing = true
			// Add a new job
			m.db[index].tests = append(m.db[index].tests, NewTest())
			dumpSlice(m.db)
			return nil
		}
	}
	return fmt.Errorf("API not found (%v)", apiID)
}

func (m *Model) StopAPIFuzzing(apiID uint) error {
	for index, api := range m.db {
		if api.id == apiID {
			logging.Logf("[Fuzzer] Model.StopAPIFuzzing(): Stop API fuzzing for api %v", api)
			m.db[index].inFuzzing = false
			m.db[index].fuzzed = true
			return nil
		}
	}
	return fmt.Errorf("API not found (%v)", apiID)
}

func (m *Model) ReceiveFullReport(ctx context.Context, apiID uint, body []byte) error {
	/*
	* Decode the result
	 */
	var data restapi.FuzzingStatusAndReport
	err := json.Unmarshal(body, &data)
	if err != nil {
		logging.Logf("Can't decode request body, error=%v", err)
		return fmt.Errorf("can't decode request body, error=%v", err)
	}
	logging.Logf("body=%v", data)

	/*
	 * Add the new status to the last Test
	 */
	api, err := m.GetAPI(ctx, apiID)
	if err != nil {
		logging.Errorf("[Fuzzer] ReceiveFullReport(): Can't retrieve api_id=(%v), error=(%v)", apiID, err)
		return fmt.Errorf("API not found (%v)", apiID)
	}

	logging.Logf("[Fuzzer] ReceiveFullReport(): API_id (%v) => API (%v)", apiID, api)
	api.AddNewStatusReport(data)
	// If the status indicate a completion, close the job
	if data.Progress == 100 && data.Status != "IN_PROGRESS" {
		err = m.StopAPIFuzzing(apiID) // TODO handle error
		if err != nil {
			logging.Errorf("[Fuzzer] ReceiveFullReport(): failed to stop fuzzing status, error=%v", err)
		}
	}
	return nil
}

func dumpSlice(s []API) {
	logging.Logf("len=%d cap=%d %v", len(s), cap(s), s)
}

func NewModel(ctx core.BackendAccessor) *Model {
	p := new(Model)
	p.accessor = ctx
	p.db = []API{}
	return p
}
