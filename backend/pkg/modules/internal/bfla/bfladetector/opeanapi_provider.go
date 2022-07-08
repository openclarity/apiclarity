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
