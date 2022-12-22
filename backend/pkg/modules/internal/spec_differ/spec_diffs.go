// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
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

// nolint: revive,stylecheck
package spec_differ

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/getkin/kin-openapi/openapi2conv"
	v3spec "github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/api3/global"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/spec_differ/config"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/spec_differ/restapi"
	speculatorutils "github.com/openclarity/apiclarity/backend/pkg/utils/speculator"
	_spec "github.com/openclarity/speculator/pkg/spec"
	_speculator "github.com/openclarity/speculator/pkg/speculator"
)

type diffHash [32]byte

const (
	moduleName = "spec_differ"
	moduleInfo = "Calculate API events spec diffs based on provided and reconstructed specs, and send diffs as notifications"
)

type specDiffer struct {
	httpHandler http.Handler

	apiIDToDiffs     map[uint]map[diffHash]global.Diff
	totalUniqueDiffs int
	config           *config.Config

	accessor core.BackendAccessor
	info     core.ModuleInfo
	sync.RWMutex
}

//nolint:gochecknoinits // was needed for the module implementation of ApiClarity
func init() {
	core.RegisterModule(newSpecDiffer)
}

//nolint:ireturn,nolintlint // was needed for the module implementation of ApiClarity
func newSpecDiffer(ctx context.Context, accessor core.BackendAccessor) (core.Module, error) {
	// Use default values
	d := &specDiffer{
		accessor:         accessor,
		config:           config.GetConfig(),
		apiIDToDiffs:     make(map[uint]map[diffHash]global.Diff),
		totalUniqueDiffs: 0,
		info: core.ModuleInfo{
			Name:        moduleName,
			Description: moduleInfo,
		},
	}
	h := restapi.HandlerWithOptions(&httpHandler{differ: d}, restapi.ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + moduleName})
	d.httpHandler = h

	go d.StartDiffsSender(ctx)

	return d, nil
}

func (s *specDiffer) Name() string {
	return moduleName
}

func (s *specDiffer) Info() core.ModuleInfo {
	return s.info
}

func (s *specDiffer) EventNotify(ctx context.Context, event *core.Event) {
	var reconstructedDiff, providedDiff *_spec.APIDiff
	var reconstructedSpecVersion, providedSpecVersion _spec.OASVersion
	var err error

	log.Infof("Got new event notification. event=%+v", event)

	apiEvent := event.APIEvent
	specKey := _speculator.GetSpecKey(apiEvent.HostSpecName, strconv.Itoa(int(apiEvent.DestinationPort)))
	speculatorAccessor := s.accessor.GetSpeculatorAccessor()

	if !speculatorAccessor.HasProvidedSpec(event.APIInfo.TraceSourceID, specKey) && !speculatorAccessor.HasApprovedSpec(event.APIInfo.TraceSourceID, specKey) {
		log.Debugf("No specs to calculate diffs")
		return
	}

	speculatorTelemetry := speculatorutils.ConvertModelsToSpeculatorTelemetry(event.Telemetry)

	reconstructedDiffType := models.DiffTypeNODIFF
	providedDiffType := models.DiffTypeNODIFF
	if speculatorAccessor.HasProvidedSpec(event.APIInfo.TraceSourceID, specKey) {
		// calculate diffs base on the event
		providedDiff, err = speculatorAccessor.DiffTelemetry(event.APIInfo.TraceSourceID, speculatorTelemetry, _spec.SpecSourceProvided)
		if err != nil {
			log.Errorf("Failed to diff telemetry against provided spec: %v", err)
			return
		}
		providedSpecVersion = s.accessor.GetSpeculatorAccessor().GetProvidedSpecVersion(event.APIInfo.TraceSourceID, specKey)
		if err := setAPIEventProvidedDiff(apiEvent, providedDiff, providedSpecVersion); err != nil {
			log.Errorf("Failed to set api event provided diff: %v", err)
			return
		}
		providedDiffType = convertToModelsDiffType(providedDiff.Type)
	}
	if speculatorAccessor.HasApprovedSpec(event.APIInfo.TraceSourceID, specKey) {
		// calculate diffs base on the event
		reconstructedDiff, err = speculatorAccessor.DiffTelemetry(event.APIInfo.TraceSourceID, speculatorTelemetry, _spec.SpecSourceReconstructed)
		if err != nil {
			log.Errorf("Failed to diff telemetry against approved spec: %v", err)
			return
		}
		reconstructedSpecVersion = s.accessor.GetSpeculatorAccessor().GetApprovedSpecVersion(event.APIInfo.TraceSourceID, specKey)
		if err := setAPIEventReconstructedDiff(apiEvent, reconstructedDiff, reconstructedSpecVersion); err != nil {
			log.Errorf("Failed to set api event reconstructed diff: %v", err)
			return
		}
		reconstructedDiffType = convertToModelsDiffType(reconstructedDiff.Type)
	}

	apiEvent.SpecDiffType = getHighestPrioritySpecDiffType(providedDiffType, reconstructedDiffType)

	// save api event with diffs in db
	if err := s.accessor.UpdateAPIEvent(ctx, apiEvent); err != nil {
		log.Errorf("Failed to update api event: %v", err)
		return
	}

	if apiEvent.HasProvidedSpecDiff {
		s.addDiffToSend(providedDiff, providedDiff.ModifiedPathItem, providedDiff.OriginalPathItem, providedDiffType, common.PROVIDED, apiEvent, providedSpecVersion)
	}
	if apiEvent.HasReconstructedSpecDiff {
		s.addDiffToSend(reconstructedDiff, reconstructedDiff.ModifiedPathItem, reconstructedDiff.OriginalPathItem, reconstructedDiffType, common.RECONSTRUCTED, apiEvent, reconstructedSpecVersion)
	}
}

func (s *specDiffer) addDiffToSend(diff *_spec.APIDiff, modifiedPathItem, originalPathItem *v3spec.PathItem, diffType models.DiffType, specType common.SpecType, event *database.APIEvent, version _spec.OASVersion) {
	if diffType == models.DiffTypeNODIFF {
		return
	}
	diffsSendThreshold := s.config.DiffsSendThreshold()
	if s.getTotalUniqueDiffs() > diffsSendThreshold {
		log.Warnf("Diff events threshold reached (%v), ignoring event", diffsSendThreshold)
		return
	}

	newSpecB, err := yaml.Marshal(getPathItemForVersionOrOriginal(modifiedPathItem, version))
	if err != nil {
		log.Errorf("Failed to marshal modified path item: %v", err)
		return
	}
	oldSpecB, err := yaml.Marshal(getPathItemForVersionOrOriginal(originalPathItem, version))
	if err != nil {
		log.Errorf("Failed to marshal original path item: %v", err)
		return
	}
	newSpec := string(newSpecB)
	oldSpec := string(oldSpecB)

	hash := sha256.Sum256([]byte(newSpec + oldSpec + string(specType)))

	apiInfo, err := s.accessor.GetAPIInfo(context.TODO(), event.APIInfoID)
	if err != nil {
		log.Errorf("Failed to get api info with apiID=%v: %v", event.APIInfoID, err)
		return
	}

	var specTimestamp time.Time
	switch specType {
	case common.PROVIDED:
		specTimestamp = time.Time(apiInfo.ProvidedSpecCreatedAt)
	case common.RECONSTRUCTED:
		specTimestamp = time.Time(apiInfo.ReconstructedSpecCreatedAt)
	case common.NONE:
		log.Warnf("spec type NONE, Using provided")
		specTimestamp = time.Time(apiInfo.ProvidedSpecCreatedAt)
	default:
		log.Warnf("Unknown spec type %v, Using provided", specType)
		specTimestamp = time.Time(apiInfo.ProvidedSpecCreatedAt)
	}

	s.Lock()
	defer s.Unlock()
	if len(s.apiIDToDiffs[event.APIInfoID]) == 0 {
		s.apiIDToDiffs[event.APIInfoID] = make(map[diffHash]global.Diff)
	}
	if _, ok := s.apiIDToDiffs[event.APIInfoID][hash]; !ok {
		s.totalUniqueDiffs++
	}
	s.apiIDToDiffs[event.APIInfoID][hash] = global.Diff{
		DiffType:      convertFromModelsDiffType(diffType),
		LastSeen:      time.Time(event.Time),
		Method:        convertFromModelsMethod(event.Method),
		NewSpec:       newSpec,
		OldSpec:       oldSpec,
		Path:          diff.Path,
		SpecTimestamp: specTimestamp,
		SpecType:      specType,
	}
}

func setAPIEventReconstructedDiff(apiEvent *database.APIEvent, reconstructedDiff *_spec.APIDiff, version _spec.OASVersion) error {
	if reconstructedDiff.Type != _spec.DiffTypeNoDiff {
		original, modified, err := convertSpecDiffToEventDiff(reconstructedDiff, version)
		if err != nil {
			return fmt.Errorf("failed to convert spec diff to event diff: %v", err)
		}
		apiEvent.HasReconstructedSpecDiff = true
		apiEvent.HasSpecDiff = true
		apiEvent.OldReconstructedSpec = string(original)
		apiEvent.NewReconstructedSpec = string(modified)
	}
	return nil
}

func setAPIEventProvidedDiff(apiEvent *database.APIEvent, providedDiff *_spec.APIDiff, version _spec.OASVersion) error {
	if providedDiff.Type != _spec.DiffTypeNoDiff {
		original, modified, err := convertSpecDiffToEventDiff(providedDiff, version)
		if err != nil {
			return fmt.Errorf("failed to convert spec diff to event diff: %v", err)
		}
		apiEvent.HasProvidedSpecDiff = true
		apiEvent.HasSpecDiff = true
		apiEvent.OldProvidedSpec = string(original)
		apiEvent.NewProvidedSpec = string(modified)
	}
	return nil
}

func convertToModelsDiffType(diffType _spec.DiffType) models.DiffType {
	switch diffType {
	case _spec.DiffTypeNoDiff:
		return models.DiffTypeNODIFF
	case _spec.DiffTypeShadowDiff:
		return models.DiffTypeSHADOWDIFF
	case _spec.DiffTypeZombieDiff:
		return models.DiffTypeZOMBIEDIFF
	case _spec.DiffTypeGeneralDiff:
		return models.DiffTypeGENERALDIFF
	default:
		log.Warnf("Unknown diff type: %v", diffType)
	}

	return models.DiffTypeNODIFF
}

func convertFromModelsDiffType(diffType models.DiffType) common.DiffType {
	switch diffType {
	case models.DiffTypeSHADOWDIFF:
		return common.SHADOWDIFF
	case models.DiffTypeNODIFF:
		return common.NODIFF
	case models.DiffTypeZOMBIEDIFF:
		return common.ZOMBIEDIFF
	case models.DiffTypeGENERALDIFF:
		return common.GENERALDIFF
	default:
		log.Warnf("Unknown diff type: %v", diffType)
		return common.NODIFF
	}
}

func convertFromModelsMethod(method models.HTTPMethod) common.HttpMethod {
	switch method {
	case models.HTTPMethodGET:
		return common.GET
	case models.HTTPMethodPOST:
		return common.POST
	case models.HTTPMethodPUT:
		return common.PUT
	case models.HTTPMethodTRACE:
		return common.TRACE
	case models.HTTPMethodDELETE:
		return common.DELETE
	case models.HTTPMethodCONNECT:
		return common.CONNECT
	case models.HTTPMethodOPTIONS:
		return common.OPTIONS
	case models.HTTPMethodHEAD:
		return common.HEAD
	case models.HTTPMethodPATCH:
		return common.PATCH
	default:
		log.Warnf("Unknown method: %v", method)
		return common.GET
	}
}

//nolint:gomnd
var diffTypePriority = map[models.DiffType]int{
	// starting from 1 since unknown type will return 0
	models.DiffTypeNODIFF:      1,
	models.DiffTypeGENERALDIFF: 2,
	models.DiffTypeSHADOWDIFF:  3,
	models.DiffTypeZOMBIEDIFF:  4,
}

// getHighestPrioritySpecDiffType will return the type with the highest priority.
func getHighestPrioritySpecDiffType(providedDiffType, reconstructedDiffType models.DiffType) models.DiffType {
	if diffTypePriority[providedDiffType] > diffTypePriority[reconstructedDiffType] {
		return providedDiffType
	}

	return reconstructedDiffType
}

type eventDiff struct {
	Path     string
	PathItem interface{}
}

func convertSpecDiffToEventDiff(diff *_spec.APIDiff, version _spec.OASVersion) (originalRet, modifiedRet []byte, err error) {
	original := eventDiff{
		Path:     diff.Path,
		PathItem: getPathItemForVersionOrOriginal(diff.OriginalPathItem, version),
	}
	modified := eventDiff{
		Path:     diff.Path,
		PathItem: getPathItemForVersionOrOriginal(diff.ModifiedPathItem, version),
	}
	originalRet, err = yaml.Marshal(original)
	if err != nil {
		return nil, nil, fmt.Errorf("failed marshal original: %v", err)
	}
	modifiedRet, err = yaml.Marshal(modified)
	if err != nil {
		return nil, nil, fmt.Errorf("failed marshal modified: %v", err)
	}

	return originalRet, modifiedRet, nil
}

func getPathItemForVersionOrOriginal(v3PathItem *v3spec.PathItem, version _spec.OASVersion) interface{} {
	if v3PathItem == nil {
		return v3PathItem
	}

	switch version {
	case _spec.OASv2:
		log.Info("Converting to OASv2 path item")
		v2PathItem, err := openapi2conv.FromV3PathItem(&v3spec.T{Components: v3spec.Components{}}, v3PathItem)
		if err != nil {
			log.Errorf("Failed to convert v3 path item to v2, keeping v3: %v", err)
			return v3PathItem
		}

		return v2PathItem
	case _spec.OASv3:
		return v3PathItem
	case _spec.Unknown:
		log.Warnf("Unknown spec version, using v3. version=%v", version)
	default:
		log.Warnf("Unknown spec version, using v3. version=%v", version)
	}

	return v3PathItem
}

func (s *specDiffer) HTTPHandler() http.Handler {
	return s.httpHandler
}
