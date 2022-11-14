package analyzesettings

import (
	"strings"
)

const (
	ArgTypeReqUrlPath  = "req_url_path"
	ArgTypeReqUrlQuery = "req_url_query"
	ArgTypeReqHeader   = "req_header"
	ArgTypeReqCookie   = "req_cookie"
	ArgTypeReqBody     = "req_body"
	ArgTypeRespHeader  = "resp_header"
	ArgTypeRespCookie  = "resp_cookie"
	ArgTypeRespBody    = "resp_body"
	MethodRPC          = "method"
	NoneAuthType       = "none"
	TokenAuthType      = "token"
	UserIDAuthType     = "user_id"
	ScopesAuthType     = "scopes"
	AccessToken        = "access_token"
)

type ArgIdentifier struct {
	ArgType string
	Name    string
	ApiName string
}

func (a ArgIdentifier) Hash() string {

	return a.ApiName + " " + a.ArgType + ":" + a.Name
}

type ApiNamingRule struct {
	ApiName   string
	Host      string
	UrlPrefix string
}

type AuthInfoType string

type AuthDiscoveryRule struct {
	ArgType string
	Name    string
	ApiName string
}

func (a AuthDiscoveryRule) Hash() string {

	return a.ApiName + " " + a.ArgType + ":" + a.Name
}

type AnalyzeSettings struct {
	revisionID     int
	apiNamingRules map[string]ApiNamingRule

	navArgs map[int64]ArgIdentifier

	navArgsMap map[ArgIdentifier]int

	autodiscoveryRules map[int64]AuthDiscoveryRule

	autodiscoveryRulesMap map[AuthDiscoveryRule]AuthInfoType
}

var analyzeSettings = AnalyzeSettings{
	revisionID:         0,
	apiNamingRules:     make(map[string]ApiNamingRule),
	navArgs:            make(map[int64]ArgIdentifier),
	autodiscoveryRules: make(map[int64]AuthDiscoveryRule),
}

func GetAPIName(host string, urlPath string) (apiName string) {
	for key, value := range analyzeSettings.apiNamingRules {
		if host == value.Host && strings.Index(urlPath, value.UrlPrefix) == 0 {
			return key
		}
	}
	return host
}

func IsNavArg(argType string, argName string, apiName string) bool {
	if argType == ArgTypeReqBody && argName == MethodRPC {
		return true
	}

	identifier := ArgIdentifier{
		ApiName: apiName,
		ArgType: argType,
		Name:    argName,
	}
	res, found := analyzeSettings.navArgsMap[identifier]
	if !found {
		return false
	}
	return res > 0
}

func GetAuthType(argType string, argName string, apiName string) AuthInfoType {
	if argType == ArgTypeReqBody && argName == AccessToken {
		return TokenAuthType
	}
	if argType == ArgTypeReqUrlQuery && argName == AccessToken {
		return TokenAuthType
	}

	identifier := AuthDiscoveryRule{
		ApiName: apiName,
		ArgType: argType,
		Name:    argName,
	}
	res, found := analyzeSettings.autodiscoveryRulesMap[identifier]
	if !found {
		return NoneAuthType
	}
	return res
}

/*func GetAPIEndpointProfile(apiName string, apiEndpointName string) (profile interface{}) {
	return nil
}*/

func GetSettingsRevisionID() int {
	return analyzeSettings.revisionID
}
func SetSettingsRevisionID(newID int) {
	analyzeSettings.revisionID = newID
}

func ApiNamingRuleDelete(apiName string) {
	delete(analyzeSettings.apiNamingRules, apiName)
}
func ApiNamingRuleAddModify(apiName string, apiNamingRule ApiNamingRule) {
	analyzeSettings.apiNamingRules[apiName] = apiNamingRule
}

func NavArgDelete(navArgID int64) {
	argIdentifier, found := analyzeSettings.navArgs[navArgID]
	if found {
		_, foundNavArg := analyzeSettings.navArgsMap[argIdentifier]
		if foundNavArg {
			analyzeSettings.navArgsMap[argIdentifier]--
			if analyzeSettings.navArgsMap[argIdentifier] <= 0 {
				delete(analyzeSettings.navArgsMap, argIdentifier)
			}

		}
		delete(analyzeSettings.navArgs, navArgID)
	}
}
func NavArgAddModify(navArgID int64, arg ArgIdentifier) {
	_, found := analyzeSettings.navArgs[navArgID]
	if found {
		NavArgDelete(navArgID)
	}
	analyzeSettings.navArgs[navArgID] = arg
	_, foundNavArgsMap := analyzeSettings.navArgsMap[arg]
	if foundNavArgsMap {
		analyzeSettings.navArgsMap[arg]++
	} else {
		analyzeSettings.navArgsMap[arg] = 1
	}
}

func AuthRuleDelete(ruleID int64) {
	argIdentifier, found := analyzeSettings.autodiscoveryRules[ruleID]
	if found {
		_, foundRule := analyzeSettings.autodiscoveryRulesMap[argIdentifier]
		if foundRule {
			delete(analyzeSettings.autodiscoveryRulesMap, argIdentifier)
		}
		delete(analyzeSettings.autodiscoveryRules, ruleID)
	}
}
func AuthRuleAddModify(ruleID int64, arg AuthDiscoveryRule, typeRule AuthInfoType) bool {
	_, found := analyzeSettings.autodiscoveryRules[ruleID]
	if found {
		AuthRuleDelete(ruleID)
	}
	analyzeSettings.autodiscoveryRules[ruleID] = arg
	_, foundRulesMap := analyzeSettings.autodiscoveryRulesMap[arg]
	if !foundRulesMap {
		analyzeSettings.autodiscoveryRulesMap[arg] = typeRule
		analyzeSettings.autodiscoveryRules[ruleID] = arg
		return true
	}

	return false
}

func init() {
	analyzeSettings.navArgsMap = map[ArgIdentifier]int{}
	analyzeSettings.autodiscoveryRulesMap = map[AuthDiscoveryRule]AuthInfoType{}
}
