package spec_differ

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-openapi/spec"
	_spec "github.com/openclarity/speculator/pkg/spec"
	_speculator "github.com/openclarity/speculator/pkg/speculator"
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
)

type diffHash [32]byte

const (
	moduleName         = "spec_differ"
	moduleInfo         = "Calculate spec diffs base on events and send diffs notifications"
	diffsSendThreshold = 500
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

	// TODO this should under the start method of API handler?
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
	var reconstructedDiff *_spec.APIDiff
	var providedDiff *_spec.APIDiff
	var err error

	log.Infof("Got new event notification. event=%+v", event)

	apiEvent := event.APIEvent
	specKey := _speculator.GetSpecKey(apiEvent.HostSpecName, strconv.Itoa(int(apiEvent.DestinationPort)))
	speculatorAccessor := s.accessor.GetSpeculatorAccessor()

	if !speculatorAccessor.HasProvidedSpec(specKey) && !speculatorAccessor.HasApprovedSpec(specKey) {
		log.Debugf("No diffs to calculate")
		return
	}

	speculatorTelemetry := speculatorutils.ConvertModelsToSpeculatorTelemetry(event.Telemetry)

	reconstructedDiffType := models.DiffTypeNODIFF
	providedDiffType := models.DiffTypeNODIFF
	if speculatorAccessor.HasProvidedSpec(specKey) {
		// calculate diffs base on the event
		providedDiff, err = speculatorAccessor.DiffTelemetry(speculatorTelemetry, _spec.DiffSourceProvided)
		if err != nil {
			log.Errorf("Failed to diff telemetry against provided spec: %v", err)
			return
		}
		if err := setAPIEventProvidedDiff(apiEvent, providedDiff); err != nil {
			log.Errorf("Failed to set api event provided diff: %v", err)
			return
		}
		providedDiffType = convertToModelsDiffType(providedDiff.Type)
	}
	if speculatorAccessor.HasApprovedSpec(specKey) {
		// calculate diffs base on the event
		reconstructedDiff, err = speculatorAccessor.DiffTelemetry(speculatorTelemetry, _spec.DiffSourceReconstructed)
		if err != nil {
			log.Errorf("Failed to diff telemetry against approved spec: %v", err)
			return
		}
		if err := setAPIEventReconstructedDiff(apiEvent, reconstructedDiff); err != nil {
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
		s.addDiffToSend(providedDiff, apiEvent.NewProvidedSpec, apiEvent.OldProvidedSpec, providedDiffType, common.PROVIDED, apiEvent)
	}
	if apiEvent.HasReconstructedSpecDiff {
		s.addDiffToSend(reconstructedDiff, apiEvent.NewReconstructedSpec, apiEvent.OldReconstructedSpec, reconstructedDiffType, common.RECONSTRUCTED, apiEvent)
	}
}

func (s *specDiffer) addDiffToSend(diff *_spec.APIDiff, newSpec, oldSpec string, diffType models.DiffType, specType common.SpecType, event *database.APIEvent, ) {
	if diffType == models.DiffTypeNODIFF {
		return
	}
	if s.getTotalUniqueDiffs() > diffsSendThreshold {
		log.Warnf("Diff events threshold reached (%v), ignoring event", diffsSendThreshold)
		return
	}

	var hash diffHash

	// TODO should we include also specType in the hash?
	hash = sha256.Sum256([]byte(newSpec + oldSpec))

	apiInfo, err := s.accessor.GetAPIInfo(context.TODO(), event.APIInfoID)
	if err != nil {
		log.Errorf("Failed to get api info with apiID=%v: %v", event.APIInfoID, err)
		return
	}

	var specTimestamp time.Time
	if specType == common.PROVIDED {
		specTimestamp = time.Time(apiInfo.ProvidedSpecCreatedAt)
	} else {
		specTimestamp = time.Time(apiInfo.ReconstructedSpecCreatedAt)
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

func setAPIEventReconstructedDiff(apiEvent *database.APIEvent, reconstructedDiff *_spec.APIDiff) error {
	if reconstructedDiff.Type != _spec.DiffTypeNoDiff {
		original, modified, err := convertSpecDiffToEventDiff(reconstructedDiff)
		if err != nil {
			return fmt.Errorf("failed to convert spec diff to event diff: %v", err)
		}
		apiEvent.HasReconstructedSpecDiff = true
		apiEvent.HasSpecDiff = true
		apiEvent.OldReconstructedSpec = string(original)
		apiEvent.NewReconstructedSpec = string(modified)
	}
	apiEvent.ReconstructedPathID = reconstructedDiff.PathID
	return nil
}

func setAPIEventProvidedDiff(apiEvent *database.APIEvent, providedDiff *_spec.APIDiff) error {
	if providedDiff.Type != _spec.DiffTypeNoDiff {
		original, modified, err := convertSpecDiffToEventDiff(providedDiff)
		if err != nil {
			return fmt.Errorf("failed to convert spec diff to event diff: %v", err)
		}
		apiEvent.HasProvidedSpecDiff = true
		apiEvent.HasSpecDiff = true
		apiEvent.OldProvidedSpec = string(original)
		apiEvent.NewProvidedSpec = string(modified)
	}
	apiEvent.ProvidedPathID = providedDiff.PathID
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
	PathItem *spec.PathItem
}

func convertSpecDiffToEventDiff(diff *_spec.APIDiff) (originalRet, modifiedRet []byte, err error) {
	original := eventDiff{
		Path:     diff.Path,
		PathItem: diff.OriginalPathItem,
	}
	modified := eventDiff{
		Path:     diff.Path,
		PathItem: diff.ModifiedPathItem,
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

func (s *specDiffer) HTTPHandler() http.Handler {
	return s.httpHandler
}
