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
	"reflect"
	"strings"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/recovery"
	"github.com/openclarity/apiclarity/backend/pkg/modules/utils"
)

func APIFindingBFLAScopesMismatch(specType SpecType, path string, method models.HTTPMethod) common.APIFinding {
	f := common.APIFinding{
		Name:        "Scopes mismatch",
		Description: "The scopes detected in the token do not match the scopes defined in the openapi specification",
		Severity:    common.HIGH,
		Source:      ModuleName,
		Type:        "BFLA_SCOPES_MISMATCH",
	}
	switch specType {
	case SpecTypeReconstructed:
		f.ReconstructedSpecLocation = getLocation(path, method)
	case SpecTypeProvided:
		f.ProvidedSpecLocation = getLocation(path, method)
	case SpecTypeNone:
		// Nothing
	}
	return f
}

func APIFindingBFLASuspiciousCallMedium(specType SpecType, path string, method models.HTTPMethod) common.APIFinding {
	f := common.APIFinding{
		Name:        "Suspicious Source Denied",
		Description: "This call looks suspicious, as it would represent a violation of the current authorization model. The API server correctly rejected the call.",
		Severity:    common.MEDIUM,
		Source:      ModuleName,
		Type:        "BFLA_SUSPICIOUS_CALL_MEDIUM",
	}
	switch specType {
	case SpecTypeReconstructed:
		f.ReconstructedSpecLocation = getLocation(path, method)
	case SpecTypeProvided:
		f.ProvidedSpecLocation = getLocation(path, method)
	case SpecTypeNone:
		// Nothing
	}
	return f
}

func APIFindingBFLASuspiciousCallHigh(specType SpecType, path string, method models.HTTPMethod) common.APIFinding {
	f := common.APIFinding{
		Name:        "Suspicious Source Allowed",
		Description: "This call looks suspicious, as it represents a violation of the current authorization model. Moreover, the API server accepted the call, which implies a possible Broken Function Level Authorisation. Please verify authorisation implementation in the API server.",
		Severity:    common.HIGH,
		Source:      ModuleName,
		Type:        "BFLA_SUSPICIOUS_CALL_HIGH",
	}
	switch specType {
	case SpecTypeReconstructed:
		f.ReconstructedSpecLocation = getLocation(path, method)
	case SpecTypeProvided:
		f.ProvidedSpecLocation = getLocation(path, method)
	case SpecTypeNone:
		// Nothing
	}
	return f
}

func getLocation(path string, method models.HTTPMethod) *string {
	s := utils.JSONPointer("paths", path, strings.ToLower(string(method)))
	return &s
}

type FindingsRegistry interface {
	Add(apiID uint, finding common.APIFinding) (bool, error)
	GetAll(apiID uint) ([]common.APIFinding, error)
	Clear(apiID uint) error
}

func NewFindingsRegistry(sp recovery.StatePersister) FindingsRegistry {
	return findingsRegistry{
		findingsMap:          recovery.NewPersistedMap(sp, BFLAFindingsAnnotationName, reflect.TypeOf(common.APIFindings{})),
		findingsByTypeAndLoc: map[typeAndLoc]struct{}{},
	}
}

type findingsRegistry struct {
	findingsMap          recovery.PersistedMap
	findingsByTypeAndLoc map[typeAndLoc]struct{}
}

type typeAndLoc struct {
	typ string
	loc string
}

func (f findingsRegistry) Add(apiID uint, ff common.APIFinding) (updated bool, err error) {
	pv, err := f.findingsMap.Get(apiID)
	if err != nil {
		return updated, fmt.Errorf("error getting findings annotation")
	}
	var findings common.APIFindings
	if pv.Exists() {
		findings, _ = pv.Get().(common.APIFindings)
	}
	if findings.Items == nil {
		findings.Items = &[]common.APIFinding{}
	}

	for _, finding := range *findings.Items {
		f.findingsByTypeAndLoc[getTypeAndLoc(finding)] = struct{}{}
	}
	_, ok := f.findingsByTypeAndLoc[getTypeAndLoc(ff)]
	if !ok {
		*findings.Items = append(*findings.Items, ff)
		updated = true
		pv.Set(findings)
	}
	return updated, nil
}

func getTypeAndLoc(f common.APIFinding) typeAndLoc {
	key := typeAndLoc{typ: f.Type}
	if f.ProvidedSpecLocation != nil {
		key.loc = *f.ProvidedSpecLocation
		return key
	}
	if f.ReconstructedSpecLocation != nil {
		key.loc = *f.ReconstructedSpecLocation
		return key
	}
	return key
}

func (f findingsRegistry) GetAll(apiID uint) ([]common.APIFinding, error) {
	pv, err := f.findingsMap.Get(apiID)
	if err != nil {
		return nil, fmt.Errorf("error getting findings annotation")
	}
	if !pv.Exists() {
		return nil, nil
	}

	findings, _ := pv.Get().(common.APIFindings)
	if *findings.Items != nil {
		return *findings.Items, nil
	}
	return nil, nil
}

func (f findingsRegistry) Clear(apiID uint) error {
	pv, err := f.findingsMap.Get(apiID)
	if err != nil {
		return fmt.Errorf("error getting findings annotation")
	}
	if !pv.Exists() {
		return nil
	}

	pv.Set(common.APIFindings{})
	return nil
}
