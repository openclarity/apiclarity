package tools

import (
	b64 "encoding/base64"
	"fmt"

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
func GetAuthStringFromParam(params restapi.FuzzTargetParams) (string, error) {
	ret := ""

	if params.Type == nil || *params.Type == "NONE" {
		return ret, nil
	}

	switch {
	case *params.Type == "apikey":

		if params.Key == nil || params.Value == nil {
			logging.Logf("Bad (%v) auth format (%v)", *params.Type, params)
			return ret, nil
		}
		sEncKey := b64.StdEncoding.EncodeToString([]byte(*params.Key))
		sEncValue := b64.StdEncoding.EncodeToString([]byte(*params.Value))
		ret = fmt.Sprintf("APIKey:%s:%s:Header", sEncKey, sEncValue)

	case *params.Type == "bearertoken":

		if params.Token == nil {
			logging.Logf("Bad (%v) auth format (%v)", *params.Type, params)
			return ret, nil
		}
		sEncToken := b64.StdEncoding.EncodeToString([]byte(*params.Token))
		ret = fmt.Sprintf("BearerToken:%s", sEncToken)

	case *params.Type == "basicauth":

		if params.Username == nil || params.Password == nil {
			logging.Logf("Bad (%v) auth format (%v)", *params.Type, params)
			return ret, nil
		}
		sEncUser := b64.StdEncoding.EncodeToString([]byte(*params.Username))
		sEncPass := b64.StdEncoding.EncodeToString([]byte(*params.Password))
		ret = fmt.Sprintf("BasicAuth:%s:%s:Header", sEncUser, sEncPass)

	default:

		logging.Logf("Not supported auth type (%v) auth format (%v)", *params.Type, params)
	}

	return ret, nil
}

func DumpHTTPFuzzParam(params restapi.FuzzTargetParams) string {
	ret := "{"
	// No ternary operator in golang... :-(
	if params.Service == nil {
		ret = ret + "Service=<nil>"
	} else {
		ret = ret + fmt.Sprintf("Service=%s", *params.Service)
	}
	if params.Type == nil {
		ret = ret + ", Authentication=<nil>"
	} else {
		ret = ret + fmt.Sprintf(", Authentication=%s", *params.Type)
	}
	if params.Token == nil {
		ret = ret + ", Token=<nil>"
	} else {
		ret = ret + fmt.Sprintf(", Token=%s", *params.Token)
	}
	if params.Key == nil {
		ret = ret + ", Key=<nil>"
	} else {
		ret = ret + fmt.Sprintf(", Key=%s", *params.Key)
	}
	if params.Value == nil {
		ret = ret + ", Value=<nil>"
	} else {
		ret = ret + fmt.Sprintf(", Value=%s", *params.Value)
	}
	if params.Username == nil {
		ret = ret + ", Username=<nil>"
	} else {
		ret = ret + fmt.Sprintf(", Username=%s", *params.Username)
	}
	if params.Password == nil {
		ret = ret + ", Password=<nil>"
	} else {
		ret = ret + fmt.Sprintf(", Password=%s", *params.Password)
	}
	ret = ret + "}"
	return ret
}

func TrimLeftChars(s string, n int) string {
	m := 0
	for i := range s {
		if m >= n {
			return s[i:]
		}
		m++
	}
	return s[:0]
}
