package profiledb

type ApiEndpointProfile struct {
	RevisionID      int
	ApiName         string
	ApiEndpointName string
	EnvironmentID   int
	CreatedAt       int
	ChangedAt       int
	Spec            interface{}
}

type ApiProfileData map[string]*ApiEndpointProfile

type ProfileDB struct {
	ApiProfile map[string]ApiProfileData
}

var profileDB = ProfileDB{
	ApiProfile: make(map[string]ApiProfileData),
}

func GetApiEndpointProfile(apiName string, apiEndpointName string) *ApiEndpointProfile {
	_, found := profileDB.ApiProfile[apiName]
	if !found {
		return nil
	}
	apiEndpointProfile, foundApiEndpoint := profileDB.ApiProfile[apiName][apiEndpointName]

	if !foundApiEndpoint {
		return nil
	}
	return apiEndpointProfile
}

func ApiEndpointProfileAddModify(apiName string, apiEndpointName string, profile *ApiEndpointProfile) {
	_, found := profileDB.ApiProfile[apiName]
	if !found {
		profileDB.ApiProfile[apiName] = make(map[string]*ApiEndpointProfile)
	}
	profileDB.ApiProfile[apiName][apiEndpointName] = profile
}

func ApiEndpointProfileDelete(apiName string, apiEndpointName string) {
	apiProfile, found := profileDB.ApiProfile[apiName]
	if found {
		delete(apiProfile, apiEndpointName)
	}
}

func ApiProfileDelete(apiName string) {
	delete(profileDB.ApiProfile, apiName)
}
