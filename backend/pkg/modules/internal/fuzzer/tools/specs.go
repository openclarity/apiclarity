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
	"context"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/ghodss/yaml"
	logging "github.com/sirupsen/logrus"
	"k8s.io/utils/strings/slices"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
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
		if err := yaml.Unmarshal([]byte(apiInfo.ProvidedSpecInfo), &specInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provided spec info. info=%+v: %v", apiInfo.ProvidedSpecInfo, err)
		}
		fuzzerSpecsInfo.ProvidedSpecInfo = &specInfo
	}

	if apiInfo.ProvidedSpec != "" {
		fuzzerSpecsInfo.ProvidedSpec = apiInfo.ProvidedSpec
	}

	if apiInfo.ReconstructedSpecInfo != "" {
		specInfo := models.SpecInfo{}
		if err := yaml.Unmarshal([]byte(apiInfo.ReconstructedSpecInfo), &specInfo); err != nil {
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
		logging.Debugf("[Fuzzer] getDocFromSpec(): spec V2 identified")
		loader := openapi3.NewLoader()
		var docV2 openapi2.T
		if err := yaml.Unmarshal(spec, &docV2); err != nil {
			return nil, fmt.Errorf("invalid V2 spec")
		}
		doc, err := openapi2conv.ToV3(&docV2)
		if err != nil {
			logging.Errorf("can't convert V2 spec to V3, err=(%v)", err)
			return nil, fmt.Errorf("can't convert V2 spec to V3")
		}
		if err := doc.Validate(loader.Context); err != nil {
			logging.Errorf("invalid V2 to V3 conversion spec result, err=(%v)", err)
			return nil, fmt.Errorf("invalid V2 to V3 conversion spec result")
		}
		return doc, nil
	} else if IsV3Specification(spec) {
		logging.Debugf("[Fuzzer] getDocFromSpec(): spec V3 identified")
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData(spec)
		if err != nil {
			logging.Errorf("can't load V3 openapi spec, err=(%v)", err)
			return nil, fmt.Errorf("can't load V3 openapi spec")
		}
		if err := doc.Validate(loader.Context); err != nil {
			logging.Errorf("invalid V3 spec, err=(%v)", err)
			return nil, fmt.Errorf("invalid V3 spec")
		}
		return doc, nil
	}
	return nil, fmt.Errorf("invalid openapi spec")
}

func GetBasePathsFromServers(servers *openapi3.Servers) []string {
	result := []string{}
	for _, server := range *servers {
		basePath := GetBasePathFromURL(server.URL)
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
		return nil, fmt.Errorf("can't convert (%v %v) to http request, err=(%v)", verb, uri, err)
	}
	req.Header.Set("Content-Type", "application/json") // Report this path in shortreport
	route, _, err := (*router).FindRoute(req)
	if err != nil {
		return nil, fmt.Errorf("can't find route for (%v %v), err=(%v)", verb, uri, err)
	}
	return route, nil
}
