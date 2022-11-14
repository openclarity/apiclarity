package tracepreprocess

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/openclarity/apiclarity/backend/pkg/analyzesettings"
	"github.com/openclarity/apiclarity/backend/pkg/dataframe"
	analyticscore "github.com/openclarity/apiclarity/backend/pkg/modules/internal/analytics_core"
	"github.com/openclarity/apiclarity/backend/pkg/profiledb"
	"github.com/openclarity/apiclarity/backend/pkg/pubsub"
	"github.com/openclarity/apiclarity/plugins/api/server/models"
)

const (
	authorizationHeader = "authorization"
	bearerAuth          = "Bearer"
)

func newModule(analyticsCore *analyticscore.AnalyticsCore) {

}

type tracePreprocessing struct {
}

func (p tracePreprocessing) GetPriority() int {
	return 10
}

func (p tracePreprocessing) GetName() analyticscore.AnalyticsModuleProccFuncName {
	return "tracePreprocessing"
}

func cookieHeader(rawCookies string) []*http.Cookie {
	header := http.Header{}
	header.Add("Cookie", rawCookies)
	req := http.Request{Header: header}
	return req.Cookies()
}

func setCookieHeader(rawCookies string) []*http.Cookie {
	header := http.Header{}
	header.Add("Set-Cookie", rawCookies)
	req := http.Request{Header: header}
	return req.Cookies()
}

func annotateRespHeader(traceEventToProcess *models.Telemetry, annotatedTrace *analyticscore.AnnotatedTrace) {
	for i := 0; i < len(traceEventToProcess.Response.Common.Headers); i++ {
		headerName := strings.ToLower(traceEventToProcess.Response.Common.Headers[i].Key)
		headerValue := traceEventToProcess.Response.Common.Headers[i].Value
		if headerName == "cookie" || headerName == "set-cookie" {
			var cookiesToParse []*http.Cookie = nil
			if headerName == "cookie" {
				cookiesToParse = cookieHeader(headerValue)
			} else {
				cookiesToParse = setCookieHeader(headerValue)
			}
			for _, cookie := range cookiesToParse {
				cookieName := cookie.Name
				cookieValue := cookie.Value
				_, found := annotatedTrace.RespArgValuesCookie[headerName]

				if !found {
					annotatedTrace.RespArgValuesCookie[cookieName] = analyticscore.ArgInstances{
						IsArray:   false,
						Instances: make(map[string]string),
					}

				}
				annotatedTrace.RespArgValuesCookie[headerName].Instances["."] = cookieValue
			}

		} else {
			_, found := annotatedTrace.RespArgValuesHeader[headerName]

			if !found {
				annotatedTrace.RespArgValuesHeader[headerName] = analyticscore.ArgInstances{
					IsArray:   false,
					Instances: make(map[string]string),
				}

			}
			annotatedTrace.RespArgValuesHeader[headerName].Instances["."] = headerValue

		}

	}
}
func annotateJwt(traceEventToProcess *models.Telemetry, annotatedTrace *analyticscore.AnnotatedTrace, tokenValue string) {
	annotatedTrace.AuthToken = tokenValue
	parser := jwt.Parser{
		UseJSONNumber:        true,
		SkipClaimsValidation: true,
	}
	token, _, err := parser.ParseUnverified(tokenValue, jwt.MapClaims{})
	if err == nil {
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			scopes := claims["scopes"]
			sub := claims["sub"]
			if scopes != nil {
				switch scopes.(type) {
				case string:
					splitScopes := strings.Split(scopes.(string), " ")
					annotatedTrace.ExplicitScopes = append(annotatedTrace.ExplicitScopes, splitScopes...)
				case []string:
					annotatedTrace.ExplicitScopes = append(annotatedTrace.ExplicitScopes, scopes.([]string)...)
				}
			}
			if sub != nil {
				switch sub.(type) {
				case string:
					annotatedTrace.ExplicitUserID = sub.(string)
				}
			}
		}
	}
}
func annotateReqHeader(traceEventToProcess *models.Telemetry, annotatedTrace *analyticscore.AnnotatedTrace) {
	for i := 0; i < len(traceEventToProcess.Request.Common.Headers); i++ {
		headerName := strings.ToLower(traceEventToProcess.Request.Common.Headers[i].Key)
		headerValue := traceEventToProcess.Request.Common.Headers[i].Value
		if headerName == "cookie" {
			var cookiesToParse []*http.Cookie = nil
			cookiesToParse = cookieHeader(headerValue)

			for _, cookie := range cookiesToParse {
				cookieName := cookie.Name
				cookieValue := cookie.Value
				_, found := annotatedTrace.ReqArgValuesCookie[headerName]

				if !found {
					annotatedTrace.ReqArgValuesCookie[cookieName] = analyticscore.ArgInstances{
						IsArray:   false,
						Instances: make(map[string]string),
					}

				}
				annotatedTrace.ReqArgValuesCookie[headerName].Instances["."] = cookieValue
			}
		} else if headerName == authorizationHeader {
			splitValue := strings.Split(headerValue, " ")
			if len(splitValue) == 2 && splitValue[0] == bearerAuth {
				annotateJwt(traceEventToProcess, annotatedTrace, splitValue[1])
			}
		} else {
			_, found := annotatedTrace.ReqArgValuesHeader[headerName]

			if !found {
				annotatedTrace.ReqArgValuesHeader[headerName] = analyticscore.ArgInstances{
					IsArray:   false,
					Instances: make(map[string]string),
				}

			}
			annotatedTrace.ReqArgValuesHeader[headerName].Instances["."] = headerValue

		}

	}
}

func annotateUrlPath(traceEventToProcess *models.Telemetry, annotatedTrace *analyticscore.AnnotatedTrace) {

}

func annotateUrlQuery(traceEventToProcess *models.Telemetry, annotatedTrace *analyticscore.AnnotatedTrace) {
	u, err := url.Parse(traceEventToProcess.Request.Path)
	if err == nil {
		query := u.RawQuery
		m, _ := url.ParseQuery(query)

		for variableKey, variableValue := range m {
			variableKeyLower := strings.ToLower(variableKey)

			_, found := annotatedTrace.ReqArgValuesURLQuery[variableKeyLower]

			if !found {
				annotatedTrace.ReqArgValuesURLQuery[variableKeyLower] = analyticscore.ArgInstances{
					IsArray:   false,
					Instances: make(map[string]string),
				}

			}

			if len(variableValue) == 1 {
				annotatedTrace.ReqArgValuesHeader[variableKeyLower].Instances["."] = variableValue[0]
			} else {
				for i := 0; i < len(variableValue); i++ {
					index := "["
					index = index + strconv.Itoa(i) + "]"
					annotatedTrace.ReqArgValuesHeader[variableKeyLower].Instances[index] = variableValue[i]
				}
			}

		}

	}
}

func annotateDecodedJsonParams(args *map[string]analyticscore.ArgInstances, val interface{}, prefix string, prefixArrays string) {

	prefixCurrent := prefix

	if prefixCurrent == "" {
		prefixCurrent = "body_raw"
	}

	switch val := val.(type) {
	case bool:
		_, found := (*args)[prefix]

		if !found {
			(*args)[prefix] = analyticscore.ArgInstances{
				IsArray:   false,
				Instances: make(map[string]string),
			}

		}
		(*args)[prefixCurrent].Instances[prefixArrays] = "True"
	case json.Number:
		_, found := (*args)[prefix]

		if !found {
			(*args)[prefix] = analyticscore.ArgInstances{
				IsArray:   false,
				Instances: make(map[string]string),
			}

		}
		(*args)[prefixCurrent].Instances[prefixArrays] = string(val)

	case string:
		_, found := (*args)[prefix]

		if !found {
			(*args)[prefixCurrent] = analyticscore.ArgInstances{
				IsArray:   false,
				Instances: make(map[string]string),
			}

		}
		(*args)[prefixCurrent].Instances[prefixArrays] = val
	case []interface{}:
		for i, v := range val {
			annotateDecodedJsonParams(args, v, prefixCurrent+"[]", prefixArrays+"["+strconv.Itoa(i)+"]")
		}
	case map[string]interface{}:
		for keyJson, valueJson := range val {
			annotateDecodedJsonParams(args, valueJson, prefix+"."+keyJson, prefixArrays+".")
		}
	case nil:
		_, found := (*args)[prefix]

		if !found {
			(*args)[prefix] = analyticscore.ArgInstances{
				IsArray:   false,
				Instances: make(map[string]string),
			}

		}
		(*args)[prefix].Instances[prefixArrays] = ""
		// nothing
	default:
		// nothing
	}
}

func annotateReqBody(traceEventToProcess *models.Telemetry, annotatedTrace *analyticscore.AnnotatedTrace) {
	var parsed interface{}

	// We deserialize the JSON object this way because json.Unmarshal doesn't
	// distinguish between int and floats. Here, thanks to d.UseNumber(), we
	// can switch on json.Number and then try to cast to Int64.
	d := json.NewDecoder(bytes.NewReader(traceEventToProcess.Request.Common.Body))
	d.UseNumber()
	canBeUrlEncoded := false

	instancesContentType, foundContentType := annotatedTrace.ReqArgValuesHeader["content-type"]
	if foundContentType {
		for _, data := range instancesContentType.Instances {
			if strings.Contains(data, "wwwurlencoded") {
				canBeUrlEncoded = true
			}
		}
	}
	var parsedJson interface{}
	if err := d.Decode(&parsed); err == nil {
		annotateDecodedJsonParams(&annotatedTrace.ReqArgValuesBody, parsedJson, "", "")
		// here we are in json body
	} else if canBeUrlEncoded {
		m, errURL := url.ParseQuery(traceEventToProcess.Request.Common.Body.String())
		if errURL == nil {
			for variableKey, variableValue := range m {
				variableKeyLower := strings.ToLower(variableKey)

				_, found := annotatedTrace.ReqArgValuesBody[variableKeyLower]

				if !found {
					annotatedTrace.ReqArgValuesBody[variableKeyLower] = analyticscore.ArgInstances{
						IsArray:   false,
						Instances: make(map[string]string),
					}

				}

				if len(variableValue) == 1 {
					annotatedTrace.ReqArgValuesBody[variableKeyLower].Instances["."] = variableValue[0]
				} else {
					for i := 0; i < len(variableValue); i++ {
						index := "["
						index = index + strconv.Itoa(i) + "]"
						annotatedTrace.ReqArgValuesBody[variableKeyLower].Instances[index] = variableValue[i]
					}
				}

			}
		}

	}
}

func annotateRespBody(traceEventToProcess *models.Telemetry, annotatedTrace *analyticscore.AnnotatedTrace) {
	var parsed interface{}

	// We deserialize the JSON object this way because json.Unmarshal doesn't
	// distinguish between int and floats. Here, thanks to d.UseNumber(), we
	// can switch on json.Number and then try to cast to Int64.
	d := json.NewDecoder(bytes.NewReader(traceEventToProcess.Response.Common.Body))
	d.UseNumber()
	canBeUrlEncoded := false

	instancesContentType, foundContentType := annotatedTrace.RespArgValuesHeader["content-type"]
	if foundContentType {
		for _, data := range instancesContentType.Instances {
			if strings.Contains(data, "wwwurlencoded") {
				canBeUrlEncoded = true
			}
		}
	}
	var parsedJson interface{}
	if err := d.Decode(&parsed); err == nil {
		annotateDecodedJsonParams(&annotatedTrace.RespArgValuesBody, parsedJson, "", "")
		// here we are in json body
	} else if canBeUrlEncoded {
		m, errURL := url.ParseQuery(traceEventToProcess.Response.Common.Body.String())
		if errURL == nil {
			for variableKey, variableValue := range m {
				variableKeyLower := strings.ToLower(variableKey)

				_, found := annotatedTrace.RespArgValuesBody[variableKeyLower]

				if !found {
					annotatedTrace.RespArgValuesBody[variableKeyLower] = analyticscore.ArgInstances{
						IsArray:   false,
						Instances: make(map[string]string),
					}

				}

				if len(variableValue) == 1 {
					annotatedTrace.RespArgValuesBody[variableKeyLower].Instances["."] = variableValue[0]
				} else {
					for i := 0; i < len(variableValue); i++ {
						index := "["
						index = index + strconv.Itoa(i) + "]"
						annotatedTrace.ReqArgValuesBody[variableKeyLower].Instances[index] = variableValue[i]
					}
				}

			}
		}

	}
}

func (p tracePreprocessing) ProccFunc(topicName analyticscore.TopicType, dataframes map[analyticscore.DataFrameID]dataframe.DataFrame, partitionID int, message pubsub.MessageForBroker, annotations []interface{}, handler *analyticscore.AnalyticsCore) (newAnnotations []interface{}) {
	switch m := message.(type) {
	case analyticscore.TraceMessageForBroker:
		traceEventToProcess := m.Event.Telemetry
		annotatedTrace := &analyticscore.AnnotatedTrace{
			ApiName:              analyzesettings.GetAPIName(traceEventToProcess.Request.Host, traceEventToProcess.Request.Path),
			ApiEndpointName:      traceEventToProcess.Request.Method + " " + traceEventToProcess.Request.Path,
			EnvironmentID:        "",
			SamplingSourceID:     "",
			ExplicitUserID:       "",
			ExplicitScopes:       nil,
			ReqArgValuesHeader:   nil,
			ReqArgValuesCookie:   nil,
			ReqArgValuesURLPath:  nil,
			ReqArgValuesURLQuery: nil,
			ReqArgValuesBody:     nil,
			RespArgValuesHeader:  nil,
			RespArgValuesCookie:  nil,
			RespArgValuesBody:    nil,
			RespCode:             0,
			SourceAddress:        traceEventToProcess.SourceAddress,
			DestinationAddress:   traceEventToProcess.DestinationAddress,
			RequestID:            traceEventToProcess.RequestID,
			Host:                 traceEventToProcess.Request.Host,
			Method:               traceEventToProcess.Request.Method,
			Path:                 traceEventToProcess.Request.Path,
			TimeRequest:          traceEventToProcess.Request.Common.Time,
			TimeResponse:         traceEventToProcess.Response.Common.Time,
			Profile:              nil,
		}
		annotatedTrace.RespCode, _ = strconv.Atoi(traceEventToProcess.Response.StatusCode)

		annotateRespHeader(traceEventToProcess, annotatedTrace)
		annotateReqHeader(traceEventToProcess, annotatedTrace)
		annotateRespBody(traceEventToProcess, annotatedTrace)
		annotateReqBody(traceEventToProcess, annotatedTrace)
		annotateUrlPath(traceEventToProcess, annotatedTrace)
		annotateUrlQuery(traceEventToProcess, annotatedTrace)
		foundNavArg := false
		for name, value := range annotatedTrace.ReqArgValuesURLQuery {
			instance, found := value.Instances["."]

			if found {
				if !foundNavArg && analyzesettings.IsNavArg(analyzesettings.ArgTypeReqUrlQuery, name, annotatedTrace.ApiName) {
					annotatedTrace.ApiEndpointName = annotatedTrace.ApiEndpointName + "URL:" + name + "=" + instance
					foundNavArg = true
				}

				ruleType := analyzesettings.GetAuthType(analyzesettings.ArgTypeReqUrlQuery, name, annotatedTrace.ApiName)
				if ruleType == analyzesettings.TokenAuthType {
					annotateJwt(traceEventToProcess, annotatedTrace, instance)
				} else if ruleType == analyzesettings.UserIDAuthType {
					annotatedTrace.ExplicitUserID = instance
				} else if ruleType == analyzesettings.ScopesAuthType {
					valueSplited := strings.Split(instance, " ")
					annotatedTrace.ExplicitScopes = append(annotatedTrace.ExplicitScopes, valueSplited...)
				}
			}

		}
		for name, value := range annotatedTrace.ReqArgValuesBody {
			instance, found := value.Instances["."]

			if found {
				if !foundNavArg && analyzesettings.IsNavArg(analyzesettings.ArgTypeReqUrlQuery, name, annotatedTrace.ApiName) {
					annotatedTrace.ApiEndpointName = annotatedTrace.ApiEndpointName + "URL:" + name + "=" + instance
					foundNavArg = true
				}

				ruleType := analyzesettings.GetAuthType(analyzesettings.ArgTypeReqUrlQuery, name, annotatedTrace.ApiName)
				if ruleType == analyzesettings.TokenAuthType {
					annotateJwt(traceEventToProcess, annotatedTrace, instance)
				} else if ruleType == analyzesettings.UserIDAuthType {
					annotatedTrace.ExplicitUserID = instance
				} else if ruleType == analyzesettings.ScopesAuthType {
					valueSplited := strings.Split(instance, " ")
					annotatedTrace.ExplicitScopes = append(annotatedTrace.ExplicitScopes, valueSplited...)
				}
			}
		}
		annotatedTrace.Profile = profiledb.GetApiEndpointProfile(annotatedTrace.ApiName, annotatedTrace.ApiEndpointName)
		annotations = append(annotations, annotatedTrace)
	default:
		return annotations
	}
	return annotations
}

//nolint:gochecknoinits
func init() {
	analyticscore.RegisterTraceAnalyticsModule(newModule, "tracepreprocess")
}
