package model

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/logging"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/restapi"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/tools"
)

type Model struct {
	db       []API
	accessor core.BackendAccessor
}

/*
* Factories
 */

func CopyTestFromTestWithReport(fullTest restapi.TestWithReport) restapi.Test {
	return restapi.Test{
		Starttime:       fullTest.Starttime,
		Progress:        fullTest.Progress,
		Vulnerabilities: fullTest.Vulnerabilities,
		ErrorMessage:    fullTest.ErrorMessage,
	}
}

func NewFinding(name string, description string, risk string) restapi.Finding {
	return restapi.Finding{
		Name:        &name,
		Description: &description,
		Risk:        &risk,
	}
}

func NewReport() restapi.FuzzingStatusAndReport {
	return restapi.FuzzingStatusAndReport{
		Progress: 0,
		Status:   restapi.DONE,
		Report:   map[string]restapi.FuzzingReportItem{},
	}

}

func NewRawFindings(message string, severity string, findingType string) restapi.RawFindings {
	return restapi.RawFindings{
		AdditionalInfo: new(string),
		Description:    &message,
		Location:       &([]string{}),
		Namespace:      new(string),
		Request: &restapi.RawFindingsSeverity{
			Severity: &severity},
		Type: &findingType,
	}
}

/*
* Model
 */

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
		if api.Id == apiID {
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

	newAPI := NewAPI(apiInfo.ID, apiInfo.Name, uint(apiInfo.Port), apiInfo.DestinationNamespace)
	logging.Logf("[model.AddAPI] Add new API %v", newAPI)
	m.db = append(m.db, newAPI)
	return &m.db[len(m.db)-1], nil
}

func (m *Model) AddAPITest(apiID uint, message string) error {
	return nil
}

func (m *Model) StartAPIFuzzing(ctx context.Context, apiID uint, specsInfo *tools.FuzzerSpecsInfo) error {
	// Get Api
	api, err := m.GetAPI(ctx, apiID)
	if err != nil {
		return fmt.Errorf("API not found (%v)", apiID)
	}
	// Start fuzzing
	err = api.StartFuzzing(specsInfo)
	if err != nil {
		return fmt.Errorf("can't start fuzzing (%v)", apiID)
	}
	//dumpSlice(m.db)
	return nil
}

func (m *Model) StopAPIFuzzing(ctx context.Context, apiID uint, fuzzerError error) error {
	// Get Api
	api, err := m.GetAPI(ctx, apiID)
	if err != nil {
		return fmt.Errorf("API not found (%v)", apiID)
	}
	// Stop fuzzing
	err = api.StopFuzzing(fuzzerError)
	if err != nil {
		return fmt.Errorf("can't stop fuzzing (%v)", apiID)
	}
	//dumpSlice(m.db)
	return nil
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
		err = m.StopAPIFuzzing(ctx, apiID, nil) // TODO handle error
		if err != nil {
			logging.Errorf("[Fuzzer] ReceiveFullReport(): failed to stop fuzzing status, error=%v", err)
		}
	}
	return nil
}

//nolint:unused,deadcode
func dumpSlice(s []API) {
	/*
	* Debug only, dump the list of APIs
	 */
	logging.Logf("len=%d cap=%d", len(s), cap(s))
	for _, api := range s {
		logging.Logf("... API {id(%v), name(%v), port(%d), fuzzed(%v), inFuzzing(%v), namespace(%v), tests(%v)}", api.Id, api.Name, api.Port, api.Fuzzed, api.InFuzzing, api.Namespace, len(api.TestsList))
		for _, testItem := range api.TestsList {
			logging.Logf("... ... test {progress(%v), start(%v), lastReport(%d), vulns(%d/%d/%d/%d/%d)}",
				testItem.Test.Progress,
				testItem.Test.Starttime,
				testItem.Test.LastReportTime,
				testItem.Test.Vulnerabilities.Total, testItem.Test.Vulnerabilities.Critical, testItem.Test.Vulnerabilities.High, testItem.Test.Vulnerabilities.Medium, testItem.Test.Vulnerabilities.Low)
		}
	}
}

func NewModel(ctx core.BackendAccessor) *Model {
	p := new(Model)
	p.accessor = ctx
	p.db = []API{}
	return p
}
