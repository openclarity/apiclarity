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

package bfladetector

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"

	"github.com/openclarity/apiclarity/backend/pkg/database"
)

func GetOpenAPI(invInfo *database.APIInfo, apiID uint) (spec *spec.Swagger, err error) {
	if invInfo.HasProvidedSpec {
		spec, err = GetServiceOpenapiSpec([]byte(invInfo.ProvidedSpec))
	} else if invInfo.HasReconstructedSpec {
		spec, err = GetServiceOpenapiSpec([]byte(invInfo.ReconstructedSpec))
	} else {
		return nil, fmt.Errorf("unable to find OpenAPI spec for service: %d", apiID)
	}
	return spec, err
}

func GetServiceOpenapiSpec(specBytes []byte) (*spec.Swagger, error) {
	jsonSpecBytes, err := yaml.YAMLToJSON(specBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert yaml spec to json: %v", err)
	}

	// Creates a new analyzed spec document for the provided spec
	analyzed, err := loads.Analyzed(jsonSpecBytes, "")
	if err != nil {
		return nil, fmt.Errorf("failed to analyze spec. Spec: %s. %v", jsonSpecBytes, err)
	}
	return analyzed.Spec(), nil
}
