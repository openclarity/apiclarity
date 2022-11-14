package analyticscore

import "github.com/openclarity/apiclarity/backend/pkg/profiledb"

type ArgInstances struct {
	IsArray   bool
	Instances map[string]string
}

type AnnotatedTrace struct {
	ApiName              string
	ApiEndpointName      string
	EnvironmentID        string
	SamplingSourceID     string
	ExplicitUserID       string
	ExplicitScopes       []string
	AuthToken            string
	ReqArgValuesHeader   map[string]ArgInstances
	ReqArgValuesCookie   map[string]ArgInstances
	ReqArgValuesURLPath  map[string]ArgInstances
	ReqArgValuesURLQuery map[string]ArgInstances
	ReqArgValuesBody     map[string]ArgInstances
	RespArgValuesHeader  map[string]ArgInstances
	RespArgValuesCookie  map[string]ArgInstances
	RespArgValuesBody    map[string]ArgInstances
	RespCode             int
	SourceAddress        string
	DestinationAddress   string
	RequestID            string
	Host                 string
	Method               string
	Path                 string
	TimeRequest          int64
	TimeResponse         int64
	Profile              *profiledb.ApiEndpointProfile
}
