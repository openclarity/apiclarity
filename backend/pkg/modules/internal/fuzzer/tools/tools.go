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

package tools

import (
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	logging "github.com/sirupsen/logrus"

	oapicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/api3/global"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/config"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/restapi"
)

// SeverityToNumber A map to used to compare Severity.
var SeverityToNumber = map[string]int{string(oapicommon.INFO): 1, string(oapicommon.LOW): 2, string(oapicommon.MEDIUM): 3, string(oapicommon.HIGH): 4, string(oapicommon.CRITICAL): 5} // nolint:gomnd

// Create new FuzzingReportPath.
func NewFuzzingReportPath(result int, verb string, uri string) restapi.FuzzingReportPath {
	return restapi.FuzzingReportPath{
		Result: &result,
		Uri:    &uri,
		Verb:   &verb,
	}
}

// Retrieve HTTP result code associated to restler item.
func GetHTTPCodeFromFindingType(findingtype string) int {
	result := 200
	// Note that for restler we haven't the result code. We can dedude the code from findingtype .
	switch {
	case findingtype == "PAYLOAD_BODY_NOT_IMPLEMENTED":
		result = 501
	case findingtype == "INTERNAL_SERVER_ERROR":
		result = 500
	}
	return result
}

/*
Create a string that contain the user auth material, on Fuzzer format.
This will be usedto send in ENV parameter to fuzzer.
*/
func GetAuthStringFromParam(params *restapi.AuthorizationScheme) (string, error) {
	ret := ""

	if params == nil {
		return ret, nil
	}

	discriminator, err := params.Discriminator()
	if err != nil {
		return ret, fmt.Errorf("failed to get discriminator, err=(%v)", err)
	}

	switch discriminator {
	case "ApiToken":

		apiToken, err := params.AsApiToken()
		if err != nil {
			msg := fmt.Sprintf("Bad ApiToken auth format (%v)", params)
			logging.Infof(msg)
			return ret, errors.New(msg)
		}
		sEncKey := b64.StdEncoding.EncodeToString([]byte(apiToken.Key))
		sEncValue := b64.StdEncoding.EncodeToString([]byte(apiToken.Value))
		ret = fmt.Sprintf("APIKey:%s:%s:Header", sEncKey, sEncValue)

	case "BasicAuth":

		basicAuth, err := params.AsBasicAuth()
		if err != nil {
			msg := fmt.Sprintf("Bad BasicAuth auth format (%v)", params)
			logging.Infof(msg)
			return ret, errors.New(msg)
		}
		sEncUser := b64.StdEncoding.EncodeToString([]byte(basicAuth.Username))
		sEncPass := b64.StdEncoding.EncodeToString([]byte(basicAuth.Password))
		ret = fmt.Sprintf("BasicAuth:%s:%s:Header", sEncUser, sEncPass)

	case "BearerToken":

		bearerToken, err := params.AsBearerToken()
		if err != nil {
			msg := fmt.Sprintf("Bad BearerToken auth format (%v)", params)
			logging.Infof(msg)
			return ret, errors.New(msg)
		}
		sEncToken := b64.StdEncoding.EncodeToString([]byte(bearerToken.Token))
		ret = fmt.Sprintf("BearerToken:%s", sEncToken)

	default:
		return ret, fmt.Errorf("unknown discriminator value: (%v)", discriminator)
	}

	return ret, nil
}

func GetTimeBudgetFromParam(param restapi.TestInputDepthEnum) (string, error) {
	ret := config.RestlerDefaultTimeBudget

	switch param {
	case restapi.QUICK:
		ret = config.RestlerQuickTimeBudget
	case restapi.DEFAULT:
		ret = config.RestlerDefaultTimeBudget
	case restapi.DEEP:
		ret = config.RestlerDeepTimeBudget
	default:
		return ret, fmt.Errorf("unknown test depth value: (%v)", string(param))
	}

	return ret, nil
}

func GetBasePathFromURL(URL string) string {
	if URL == "" || URL == "/" {
		return ""
	}

	// strip scheme if exits
	urlNoScheme := URL
	schemeSplittedURL := strings.Split(URL, "://")
	if len(schemeSplittedURL) > 1 {
		urlNoScheme = schemeSplittedURL[1]
	}

	// get path
	var path string
	splittedURLNoScheme := strings.SplitN(urlNoScheme, "/", 2) // nolint:gomnd
	if len(splittedURLNoScheme) > 1 {
		path = splittedURLNoScheme[1]
	}
	if path == "" {
		return ""
	}

	return "/" + path
}

func ConvertLocalToGlobalReportTag(from *[]restapi.FuzzingReportTag) *[]global.FuzzingReportTag {
	if from == nil {
		// If there is no tags on input, no need to convert, result must be null. It is not an error.
		return nil
	}

	/*
	* We need to convert restapi.FuzzingReportTag to global.FuzzingReportTag
	* It is the same struc, because global.FuzzingReportTag is created from restapi.FuzzingReportTag.
	* But: We need to use restapi.FuzzingReportTag on restapi.gen.go and Fuzzer functions that use it,
	* and we need to use global.FuzzingReportTag for notifications.
	* As it is same struct, it is valid here to unmarshall then marshall things.
	 */
	result := []global.FuzzingReportTag{}
	bytes, err := json.Marshal(from)
	if err != nil {
		logging.Errorf("[Fuzzer] ConvertLocalToGlobalReportTag(): Failed to Marshal []restapi.FuzzingReportTag (%v)", err)
		return nil
	}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		logging.Errorf("[Fuzzer] ConvertLocalToGlobalReportTag(): Failed to Unmarshal []global.FuzzingReportTag (%v)", err)
		return nil
	}
	return &result
}

func IsGreaterSeverity(severity1 oapicommon.Severity, severity2 oapicommon.Severity) bool {
	/*
	* Severity comparison operator.
	* Return true is s1>s2, false otherwise
	 */
	s1AsNumber := SeverityToNumber[string(severity1)]
	s2AsNumber := SeverityToNumber[string(severity2)]
	return s1AsNumber > s2AsNumber
}

func IsDone(report *restapi.FuzzingStatusAndReport) bool {
	if report == nil || report.Status == restapi.INPROGRESS {
		return false
	}
	return true
}
