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
	"time"

	logging "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/restapi"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/tools"
)

// FuzzingTimestamp the type used for our Timestamp.
type FuzzingTimestamp = int64

// ZeroTime The zero timestamp to use.
var ZeroTime = time.Time{}.Unix()

// Model: The Model struct.
type Model struct {
	db       []API
	accessor core.BackendAccessor
}

// FuzzingInput: a struct to store all input parameters for fuzzing.
type FuzzingInput struct {
	Auth      *restapi.AuthorizationScheme
	Depth     restapi.TestInputDepthEnum
	SpecsInfo *tools.FuzzerSpecsInfo
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
			Severity: &severity,
		},
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
	logging.Infof("[model.AddAPI] Add new API %v", api)
	m.db = append(m.db, api)
	return &m.db[len(m.db)-1]
}

func (m *Model) GetAPI(ctx context.Context, apiID uint) (*API, error) {
	/*
	* Try to retrieve it from the cache
	 */
	for index, api := range m.db {
		if api.ID == apiID {
			return &m.db[index], nil
		}
	}

	logging.Infof("[model.GetAPI] API %v not found, try to retrieve it from backend", apiID)

	/*
	* Try to retrieve it from backend
	 */
	apiInfo, err := m.accessor.GetAPIInfo(ctx, apiID)
	logging.Debugf("[model.GetAPI(%v)]: get apiInfo=(%v)", apiID, apiInfo)
	if err != nil {
		return nil, fmt.Errorf("error when retrieve api %v: %v", apiID, err)
	}

	newAPI := NewAPI(apiInfo.ID, apiInfo.Name, uint(apiInfo.Port), apiInfo.DestinationNamespace)
	logging.Infof("[model.AddAPI] Add new API %v", newAPI)
	m.db = append(m.db, newAPI)
	return &m.db[len(m.db)-1], nil
}

func (m *Model) AddAPITest(apiID uint, message string) error {
	return nil
}

func (m *Model) StartAPIFuzzing(ctx context.Context, apiID uint, params *FuzzingInput) (FuzzingTimestamp, error) {
	// Get Api
	api, err := m.GetAPI(ctx, apiID)
	if err != nil {
		return ZeroTime, fmt.Errorf("API not found (%v)", apiID)
	}
	// Start fuzzing
	timestamp, err := api.StartFuzzing(params)
	if err != nil {
		return ZeroTime, fmt.Errorf("can't start fuzzing (%v)", apiID)
	}
	// dumpSlice(m.db)
	return timestamp, nil
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
		err2 := api.SetErrorForLastStatus("failed to stop Fuzzing")
		if err2 != nil {
			logging.Errorf("[Fuzzer] StopAPIFuzzing(): can't set last status error for API (%v), err=(%v)", apiID, err)
		}
		return fmt.Errorf("can't stop fuzzing (%v)", apiID)
	}
	return nil
}

func (m *Model) ReceiveFullReport(ctx context.Context, apiID uint, body []byte) error {
	/*
	* Decode the result
	 */
	var data restapi.FuzzingStatusAndReport
	err := json.Unmarshal(body, &data)
	if err != nil {
		logging.Infof("Can't decode request body, error=%v", err)
		return fmt.Errorf("can't decode request body, error=%v", err)
	}
	logging.Infof("body=%v", data)

	/*
	 * Add the new status to the last Test
	 */
	api, err := m.GetAPI(ctx, apiID)
	if err != nil {
		logging.Errorf("[Fuzzer] ReceiveFullReport(): Can't retrieve api_id=(%v), error=(%v)", apiID, err)
		return fmt.Errorf("API not found (%v)", apiID)
	}

	logging.Infof("[Fuzzer] ReceiveFullReport(): API_id (%v) => API (%v)", apiID, api)
	err = api.AddNewStatusReport(data)
	if err != nil {
		logging.Errorf("[Fuzzer] ReceiveFullReport(): Can't add new report, error=(%v)", err)
	}

	// If the status indicate a completion, close the job
	if data.Progress == 100 && data.Status != "IN_PROGRESS" {
		err = m.StopAPIFuzzing(ctx, apiID, nil) // TODO handle error
		if err != nil {
			logging.Errorf("[Fuzzer] ReceiveFullReport(): failed to stop fuzzing status, error=%v", err)
		}
	}
	return nil
}

//nolint:unused,deadcode // used for debug only
func dumpSlice(s []API) {
	/*
	* Debug only, dump the list of APIs
	 */
	logging.Infof("len=%d cap=%d", len(s), cap(s))
	for _, api := range s {
		logging.Infof("... API {id(%v), name(%v), port(%d), fuzzed(%v), inFuzzing(%v), namespace(%v), tests(%v)}", api.ID, api.Name, api.Port, api.Fuzzed, api.InFuzzing, api.Namespace, len(api.TestsList))
		for _, testItem := range api.TestsList {
			logging.Infof("... ... test {progress(%v), start(%v), lastReport(%d), vulns(%d/%d/%d/%d/%d)}",
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
