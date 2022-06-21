package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"k8s.io/utils/strings/slices"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/logging"
	"gopkg.in/yaml.v2"
)

// FuzzerSpecInfo An object containing info about a spec.
type FuzzerSpecsInfo struct {
	ProvidedSpec          string
	ReconstructedSpec     string
	ProvidedSpecInfo      *models.SpecInfo
	ReconstructedSpecInfo *models.SpecInfo
}

func GetAPISpecsInfo(ctx context.Context, accessor core.BackendAccessor, apiID uint) (*FuzzerSpecsInfo, error) {
	fuzzerSpecsInfo := FuzzerSpecsInfo{}

	apiInfo, err := accessor.GetAPIInfo(ctx, apiID)
	if err != nil {
		return nil, fmt.Errorf("can't access to apiID (%v), error=%v", apiID, err)
	}

	if apiInfo.ProvidedSpecInfo != "" {
		specInfo := models.SpecInfo{}
		if err := json.Unmarshal([]byte(apiInfo.ProvidedSpecInfo), &specInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provided spec info. info=%+v: %v", apiInfo.ProvidedSpecInfo, err)
		}
		fuzzerSpecsInfo.ProvidedSpecInfo = &specInfo
	}

	if apiInfo.ProvidedSpec != "" {
		fuzzerSpecsInfo.ProvidedSpec = apiInfo.ProvidedSpec
	}

	if apiInfo.ReconstructedSpecInfo != "" {
		specInfo := models.SpecInfo{}
		if err := json.Unmarshal([]byte(apiInfo.ReconstructedSpecInfo), &specInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal reconstructed spec info. info=%+v: %v", apiInfo.ReconstructedSpecInfo, err)
		}
		fuzzerSpecsInfo.ReconstructedSpecInfo = &specInfo
	}

	if apiInfo.ReconstructedSpec != "" {
		fuzzerSpecsInfo.ReconstructedSpec = apiInfo.ReconstructedSpec
	}

	return &fuzzerSpecsInfo, nil
}

func IsV2Specification(data []byte) bool {
	specV2OrV3 := struct {
		Openapi *string `json:"openapi"`
		Swagger *string `json:"swagger"`
	}{}
	if err := yaml.Unmarshal(data, &specV2OrV3); err != nil {
		return false
	}
	return specV2OrV3.Swagger != nil
}

func IsV3Specification(data []byte) bool {
	specV2OrV3 := struct {
		Openapi *string `json:"openapi"`
		Swagger *string `json:"swagger"`
	}{}
	if err := yaml.Unmarshal(data, &specV2OrV3); err != nil {
		return false
	}
	return specV2OrV3.Openapi != nil
}

func LoadSpec(spec []byte) (*openapi3.T, error) {
	if IsV2Specification(spec) {
		logging.Logf("[Fuzzer] getDocFromSpec(): spec V2 identified")
		loader := openapi3.NewLoader()
		var docV2 openapi2.T
		if err := json.Unmarshal(spec, &docV2); err != nil {
			panic(err)
		}
		doc, err := openapi2conv.ToV3(&docV2)
		if err != nil {
			panic(err)
		}
		if err := doc.Validate(loader.Context); err != nil {
			panic(err)
		}
		return doc, nil
	} else if IsV3Specification(spec) {
		logging.Logf("[Fuzzer] getDocFromSpec(): spec V3 identified")
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData(spec)
		if err != nil {
			panic(err)
		}
		if err := doc.Validate(loader.Context); err != nil {
			panic(err)
		}
		return doc, nil
	}
	return nil, fmt.Errorf("invalid spec")
}

func GetBasePathsFromServers(servers *openapi3.Servers) []string {
	result := []string{}
	for _, server := range *servers {
		// convert the full URL in http request to extract later the base path
		// olint:noctx	// No need of context, the http.NewRequest is used only for formatting
		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			panic(err)
		}
		basePath := req.URL.Path
		if !slices.Contains(result, basePath) {
			result = append(result, basePath)
		}
	}
	return result
}

func FindRoute(router *routers.Router, verb string, uri string) (*routers.Route, error) {
	logging.Debugf("[Fuzzer] findRoute(): process path (%v %v)", verb, uri)
	//nolint:noctx // No need of context, the http.NewRequest is used only for formatting
	req, err := http.NewRequest(verb, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("can't convert (%v %v) to http request, err=(%v)", verb, uri, err.Error())
	}
	req.Header.Set("Content-Type", "application/json") // Report this path in shortreport
	logging.Debugf("[Fuzzer] findRoute(): ... req to find (%v %v)", verb, req)
	route, _, err := (*router).FindRoute(req)
	if err != nil {
		return nil, fmt.Errorf("can't find route for (%v %v), err=(%v)", verb, uri, err.Error())
	}
	return route, nil
}
