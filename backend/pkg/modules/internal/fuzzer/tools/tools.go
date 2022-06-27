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
	"errors"
	"fmt"
	"strings"

	"encoding/json"

	"github.com/openclarity/apiclarity/api3/global"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/config"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/logging"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/restapi"
)

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
		return ret, err
	}

	switch discriminator {
	case "ApiToken":

		apiToken, err := params.AsApiToken()
		if err != nil {
			msg := fmt.Sprintf("Bad ApiToken auth format (%v)", params)
			logging.Logf(msg)
			return ret, errors.New(msg)
		}
		sEncKey := b64.StdEncoding.EncodeToString([]byte(apiToken.Key))
		sEncValue := b64.StdEncoding.EncodeToString([]byte(apiToken.Value))
		ret = fmt.Sprintf("APIKey:%s:%s:Header", sEncKey, sEncValue)

	case "BasicAuth":

		basicAuth, err := params.AsBasicAuth()
		if err != nil {
			msg := fmt.Sprintf("Bad BasicAuth auth format (%v)", params)
			logging.Logf(msg)
			return ret, errors.New(msg)
		}
		sEncUser := b64.StdEncoding.EncodeToString([]byte(basicAuth.Username))
		sEncPass := b64.StdEncoding.EncodeToString([]byte(basicAuth.Password))
		ret = fmt.Sprintf("BasicAuth:%s:%s:Header", sEncUser, sEncPass)

	case "BearerToken":

		bearerToken, err := params.AsBearerToken()
		if err != nil {
			msg := fmt.Sprintf("Bad BearerToken auth format (%v)", params)
			logging.Logf(msg)
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

func GetAuthSchemeFromFuzzTargetParams(params restapi.FuzzTargetParams) (*restapi.AuthorizationScheme, error) {

	if params.Type == nil || *params.Type == "NONE" {
		return nil, nil
	}

	authScheme := restapi.AuthorizationScheme{}

	switch {
	case *params.Type == "apikey":

		if params.Key == nil || params.Value == nil {
			msg := fmt.Sprintf("Bad (%v) auth format (%v)", *params.Type, params)
			logging.Logf(msg)
			return nil, errors.New(msg)
		}
		ret := authScheme.FromApiToken(
			restapi.ApiToken{
				Key:   *params.Key,
				Value: *params.Value,
				Type:  restapi.APITOKEN,
			},
		)
		if ret != nil {
			return nil, ret
		}

	case *params.Type == "bearertoken":

		if params.Token == nil {
			msg := fmt.Sprintf("Bad (%v) auth format (%v)", *params.Type, params)
			logging.Logf(msg)
			return nil, errors.New(msg)
		}
		ret := authScheme.FromBearerToken(
			restapi.BearerToken{
				Token: *params.Token,
				Type:  restapi.BEARERTOKEN,
			},
		)
		if ret != nil {
			return nil, ret
		}

	case *params.Type == "basicauth":

		if params.Username == nil || params.Password == nil {
			msg := fmt.Sprintf("Bad (%v) auth format (%v)", *params.Type, params)
			logging.Logf(msg)
			return nil, errors.New(msg)
		}
		ret := authScheme.FromBasicAuth(
			restapi.BasicAuth{
				Username: *params.Username,
				Password: *params.Password,
				Type:     restapi.BASICAUTH,
			},
		)
		if ret != nil {
			return nil, ret
		}

	default:

		msg := fmt.Sprintf("Not supported auth type (%v) auth format (%v)", *params.Type, params)
		logging.Logf(msg)
		return nil, errors.New(msg)
	}

	return &authScheme, nil
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

func ConvertLocalToGlobalReportTag(from []restapi.FuzzingReportTag) ([]global.FuzzingReportTag, error) {
	/*
	* We need to convert restapi.FuzzingReportTag to global.FuzzingReportTag
	* It is the same struc, because global.FuzzingReportTag is created from restapi.FuzzingReportTag.
	* But: We need to use restapi.FuzzingReportTag on restapi.gen.go and Fuzzer functions that use it,
	* and we need to use global.FuzzingReportTag for notifications.
	* As it is same struct, it is valid here to unmarshall then marshall things.
	 */
	to := []global.FuzzingReportTag{}
	bytes, err := json.Marshal(from)
	if err != nil {
		return to, err
	}
	err = json.Unmarshal(bytes, &to)
	if err != nil {
		return to, err
	}
	return to, nil
}
