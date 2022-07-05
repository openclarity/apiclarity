package bfladetector

import (
	"fmt"
	"github.com/openclarity/apiclarity/api/server/models"
	"strings"

	"github.com/openclarity/apiclarity/api3/common"
)

func APIFindingBFLAScopesMismatch(isReconstructed bool, path string, method models.HTTPMethod) common.APIFinding {
	f := common.APIFinding{
		Name:        "Scopes mismatch",
		Description: "The scopes detected in the token do not match the scopes defined in the openapi specification",
		Severity:    common.HIGH,
		Source:      ModuleName,
		Type:        "BFLA_SCOPES_MISMATCH",
	}
	if isReconstructed {
		f.ReconstructedSpecLocation = getLocation(path, method)
	} else {
		f.ProvidedSpecLocation = getLocation(path, method)
	}
	return f
}

func APIFindingBFLASuspiciousCallMedium(isReconstructed bool, path string, method models.HTTPMethod) common.APIFinding {
	f := common.APIFinding{
		Name:        "Suspicious Source Denied",
		Description: "This call looks suspicious, as it would represent a violation of the current authorization model. The API server correctly rejected the call.",
		Severity:    common.MEDIUM,
		Source:      ModuleName,
		Type:        "BFLA_SUSPICIOUS_CALL_MEDIUM",
	}
	if isReconstructed {
		f.ReconstructedSpecLocation = getLocation(path, method)
	} else {
		f.ProvidedSpecLocation = getLocation(path, method)
	}
	return f
}

func APIFindingBFLASuspiciousTraceHigh(isReconstructed bool, path string, method models.HTTPMethod) common.APIFinding {
	f := common.APIFinding{
		Description: "Suspicious Source Allowed",
		Name:        "This call looks suspicious, as it represents a violation of the current authorization model. Moreover, the API server accepted the call, which implies a possible Broken Function Level Authorisation. Please verify authorisation implementation in the API server.",
		Severity:    common.HIGH,
		Source:      ModuleName,
		Type:        "BFLA_SUSPICIOUS_CALL_HIGH",
	}
	if isReconstructed {
		f.ReconstructedSpecLocation = getLocation(path, method)
	} else {
		f.ProvidedSpecLocation = getLocation(path, method)
	}
	return f
}

func getLocation(path string, method models.HTTPMethod) *string {
	s := fmt.Sprintf("/paths/%s/%s", path, strings.ToLower(string(method)))
	return &s
}
