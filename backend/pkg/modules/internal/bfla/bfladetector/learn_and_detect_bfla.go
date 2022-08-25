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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/go-openapi/spec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/api3/notifications"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/k8straceannotator"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/recovery"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/restapi"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
)

const (
	ModuleName               = "bfla"
	ModuleDescription        = "Reconstructs an authorization model for an API and detects violations of such authorization model"
	K8sSrcAnnotationName     = "bfla_k8s_src"
	K8sDstAnnotationName     = "bfla_k8s_dst"
	DetectedIDAnnotationName = "bfla_detected_id"

	AuthzModelAnnotationName   = "authz_model"
	BFLAStateAnnotationName    = "bfla_state"
	BFLAFindingsAnnotationName = "bfla_findings"

	automaticLearningAndDetectionEnv     = "BFLA_AUTOMATIC_LEARNING_AND_DETECTION"
	automaticLearningAndDetectionDefault = false

	learningNrTracesEnv     = "BFLA_LEARNING_NR_TRACES"
	learningNrTracesDefault = 100
)

type BFLAStateEnum uint

const (
	BFLAStart BFLAStateEnum = iota
	BFLALearning
	BFLALearnt
	BFLADetecting
)

type BflaConfig struct {
	AutomaticLearningAndDetection bool
	LearningNrTraces              uint
}

func loadConfig() BflaConfig {
	viper.SetDefault(automaticLearningAndDetectionEnv, automaticLearningAndDetectionDefault)
	viper.SetDefault(learningNrTracesEnv, learningNrTracesDefault)

	return BflaConfig{
		AutomaticLearningAndDetection: viper.GetBool(automaticLearningAndDetectionEnv),
		LearningNrTraces:              viper.GetUint(learningNrTracesEnv),
	}
}

func (s BFLAStateEnum) String() string {
	switch s {
	case BFLAStart:
		return "START"
	case BFLALearning:
		return "LEARNING"
	case BFLALearnt:
		return "LEARNT"
	case BFLADetecting:
		return "DETECTING"
	}
	return "UNKNOWN"
}

type BFLAState struct {
	State        BFLAStateEnum `json:"state"`
	TraceCounter int           `json:"trace_counter"`
}

var ErrUnsupportedAuthScheme = errors.New("unsupported auth scheme")

func NewBFLADetector(ctx context.Context, modName string, bflaBackendAccessor bflaBackendAccessor, eventAlerter EventAlerter, bflaNotifier BFLANotifier, sp recovery.StatePersister, notifierResyncInterval time.Duration) BFLADetector {
	l := &learnAndDetectBFLA{
		tracesCh:               make(chan *CompositeTrace),
		commandsCh:             make(chan Command),
		errCh:                  make(chan error),
		bflaBackendAccessor:    bflaBackendAccessor,
		authzModelsMap:         recovery.NewPersistedMap(sp, AuthzModelAnnotationName, reflect.TypeOf(AuthorizationModel{})),
		bflaStateMap:           recovery.NewPersistedMap(sp, BFLAStateAnnotationName, reflect.TypeOf(BFLAState{})),
		statePersister:         sp,
		eventAlerter:           eventAlerter,
		bflaNotifier:           bflaNotifier,
		notifierResyncInterval: notifierResyncInterval,
		findingsRegistry:       NewFindingsRegistry(sp),
		mu:                     &sync.RWMutex{},
		modName:                modName,
		config:                 loadConfig(),
	}
	go func() {
		for {
			select {
			case err := <-l.errCh:
				log.Errorf("BFLA error: %s", err)
			case <-ctx.Done():
				log.Error("BFLA done; ", ctx.Err())
				return
			}
		}
	}()
	go l.run(ctx)
	return l
}

type BFLADetector interface {
	SendTrace(trace *CompositeTrace)

	IsLearning(apiID uint) bool
	GetState(apiID uint) (BFLAStateEnum, error)
	FindSourceObj(path, method, clientUID string, apiID uint) (*SourceObject, error)

	ApproveTrace(path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, user *DetectedUser)
	DenyTrace(path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, user *DetectedUser)

	ResetModel(apiID uint) error
	StartLearning(apiID uint, numberOfTraces int) error
	StopLearning(apiID uint) error

	StartDetection(apiID uint) error
	StopDetection(apiID uint) error

	ProvideAuthzModel(apiID uint, am AuthorizationModel)
}

type bflaBackendAccessor interface {
	GetAPIInfo(ctx context.Context, apiID uint) (*database.APIInfo, error)
	EnableTraces(ctx context.Context, modName string, apiID uint) error
	DisableTraces(ctx context.Context, modName string, apiID uint) error
}

type Command interface{ isCommand() }

type CommandWithError interface {
	Command

	Close()
	SendError(err error)
	RcvError() error
}

type ErrorChan chan error

func NewErrorChan() ErrorChan           { return make(chan error, 1) }
func (e ErrorChan) SendError(err error) { e <- err }
func (e ErrorChan) Close()              { close(e) }
func (e ErrorChan) RcvError() error     { return <-e }

type StopLearningCommand struct {
	apiID uint
	ErrorChan
}

type StartLearningCommand struct {
	apiID          uint
	numberOfTraces int
	ErrorChan
}

type ResetModelCommand struct {
	apiID uint
	ErrorChan
}

type MarkLegitimateCommand struct {
	path         string
	method       string
	detectedUser *DetectedUser
	clientRef    *k8straceannotator.K8sObjectRef
	apiID        uint
}

type MarkIllegitimateCommand struct {
	path         string
	method       string
	detectedUser *DetectedUser
	clientRef    *k8straceannotator.K8sObjectRef
	apiID        uint
}

type ProvideAuthzModelCommand struct {
	apiID      uint
	authzModel AuthorizationModel
}

type StartDetectionCommand struct {
	apiID uint
	ErrorChan
}

type StopDetectionCommand struct {
	apiID uint
	ErrorChan
}

func (a *StopLearningCommand) isCommand()      {}
func (a *StartLearningCommand) isCommand()     {}
func (a *ResetModelCommand) isCommand()        {}
func (a *MarkLegitimateCommand) isCommand()    {}
func (a *MarkIllegitimateCommand) isCommand()  {}
func (a *ProvideAuthzModelCommand) isCommand() {}
func (a *StartDetectionCommand) isCommand()    {}
func (a *StopDetectionCommand) isCommand()     {}

type EventOperation struct {
	Path        string
	Method      string
	Source      string
	Destination string
}

type EventAlerter interface {
	SetEventAlert(ctx context.Context, modName string, eventID uint, severity core.AlertSeverity) error
}

type learnAndDetectBFLA struct {
	tracesCh            chan *CompositeTrace
	commandsCh          CommandsChan
	errCh               chan error
	bflaBackendAccessor bflaBackendAccessor

	authzModelsMap recovery.PersistedMap
	bflaStateMap   recovery.PersistedMap

	statePersister recovery.StatePersister

	eventAlerter           EventAlerter
	bflaNotifier           BFLANotifier
	notifierResyncInterval time.Duration
	findingsRegistry       FindingsRegistry
	mu                     *sync.RWMutex
	modName                string
	config                 BflaConfig
}

type CommandsChan chan Command

func (c CommandsChan) Send(cmd Command) {
	c <- cmd
}

func (c CommandsChan) SendAndReplyErr(cmd CommandWithError) error {
	defer cmd.Close()
	c <- cmd
	return cmd.RcvError() //nolint:wrapcheck
}

type CompositeTrace struct {
	*core.Event

	K8SSource, K8SDestination *k8straceannotator.K8sObjectRef
	DetectedUser              *DetectedUser
}

func (l *learnAndDetectBFLA) logError(err error) {
	if err != nil {
		log.Error(err)
	}
}

func (l *learnAndDetectBFLA) run(ctx context.Context) {
	defer log.Info("ending learnFromTracesAndDetectBFLA")

	for {
		select {
		case command, ok := <-l.commandsCh:
			if ok {
				err := l.commandsRunner(ctx, command)
				if cmdErr, ok := command.(CommandWithError); ok {
					cmdErr.SendError(err)
				}
				if err != nil {
					l.errCh <- err
				}
				continue
			}
		case trace, ok := <-l.tracesCh:
			if ok {
				if err := l.traceRunner(ctx, trace); err != nil {
					l.errCh <- err
				}
				continue
			}
		case <-ctx.Done():
			log.Error(ctx.Err())
		}
		return
	}
}

func runtimeRecover() {
	if err := recover(); err != nil {
		log.Error(err)
		debug.PrintStack()
	}
}

func (l *learnAndDetectBFLA) checkBFLAState(apiID uint, allowedStates ...BFLAStateEnum) (BFLAState, recovery.PersistedValue, error) {
	stateValue, err := l.bflaStateMap.Get(apiID)
	if err != nil {
		return BFLAState{}, nil, fmt.Errorf("unable to get state traces counter: %w", err)
	}
	state := stateValue.Get().(BFLAState) //nolint:forcetypeassert
	log.Debugf("current state for api %d is %v", apiID, state.State)
	for _, s := range allowedStates {
		if state.State == s {
			return state, stateValue, nil
		}
	}
	return state, stateValue, fmt.Errorf("state %v does not allow for the requested operation", state.State)
}

//nolint:gocyclo
func (l *learnAndDetectBFLA) commandsRunner(ctx context.Context, command Command) (err error) {
	defer runtimeRecover()
	switch cmd := command.(type) {
	case *MarkLegitimateCommand:
		apiInfo, err := l.bflaBackendAccessor.GetAPIInfo(ctx, cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get api info: %w", err)
		}
		tags, err := ParseSpecInfo(apiInfo)
		if err != nil {
			return fmt.Errorf("unable to parse spec info: %w", err)
		}
		_, _, err = l.checkBFLAState(cmd.apiID, BFLALearnt, BFLADetecting)
		if err != nil {
			return fmt.Errorf("unable to perform command 'Mark Legitimate': %w", err)
		}
		err = l.updateAuthorizationModel(tags, cmd.path, cmd.method, cmd.clientRef, cmd.apiID, cmd.detectedUser, true, true)
		if err != nil {
			return fmt.Errorf("unable to perform command 'Mark Legitimate': %w", err)
		}
		clientUID := ""
		if cmd.clientRef != nil {
			clientUID = cmd.clientRef.Uid
		}
		aud, setAud, err := l.findSourceObj(cmd.path, cmd.method, clientUID, cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to find source obj: %w", err)
		}
		aud.WarningStatus = restapi.LEGITIMATE
		setAud(aud)
		l.logError(l.notifyAuthzModel(ctx, cmd.apiID))

	case *MarkIllegitimateCommand:
		apiInfo, err := l.bflaBackendAccessor.GetAPIInfo(ctx, cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get api info: %w", err)
		}
		tags, err := ParseSpecInfo(apiInfo)
		if err != nil {
			return fmt.Errorf("unable to parse spec info: %w", err)
		}
		_, _, err = l.checkBFLAState(cmd.apiID, BFLALearnt, BFLADetecting)
		if err != nil {
			return fmt.Errorf("unable to perform command 'Mark Illegitimate': %w", err)
		}
		err = l.updateAuthorizationModel(tags, cmd.path, cmd.method, cmd.clientRef, cmd.apiID, cmd.detectedUser, false, true)
		if err != nil {
			return fmt.Errorf("unable to perform command 'Mark Illegitimate': %w", err)
		}
		clientUID := ""
		if cmd.clientRef != nil {
			clientUID = cmd.clientRef.Uid
		}
		aud, setAud, err := l.findSourceObj(cmd.path, cmd.method, clientUID, cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to find source obj: %w", err)
		}
		aud.WarningStatus = ResolveBFLAStatusInt(int(aud.StatusCode))
		setAud(aud)
	case *StopLearningCommand:
		state, stateValue, err := l.checkBFLAState(cmd.apiID, BFLALearning)
		if err != nil {
			return fmt.Errorf("unable to perform command 'Stop Learning': %w", err)
		}
		state.TraceCounter = 0
		if l.config.AutomaticLearningAndDetection {
			state.State = BFLADetecting
		} else {
			err = l.bflaBackendAccessor.DisableTraces(ctx, l.modName, cmd.apiID)
			if err != nil {
				return fmt.Errorf("cannot disable traces: %w", err)
			}
			state.State = BFLALearnt
		}
		stateValue.Set(state)
		l.logError(l.notifyAuthzModel(ctx, cmd.apiID))

	case *StartLearningCommand:
		state, stateValue, err := l.checkBFLAState(cmd.apiID, BFLAStart, BFLALearnt, BFLADetecting)
		if err != nil {
			return fmt.Errorf("unable to perform command 'Start Learning': %w", err)
		}
		if state.State == BFLAStart || state.State == BFLALearnt {
			err = l.bflaBackendAccessor.EnableTraces(ctx, l.modName, cmd.apiID)
			if err != nil {
				return fmt.Errorf("cannot enable traces: %w", err)
			}
		}
		state.State = BFLALearning
		state.TraceCounter = cmd.numberOfTraces
		stateValue.Set(state)
		err = l.initAuthorizationModel(cmd.apiID)
		if err != nil {
			return fmt.Errorf("cannot initialize authorization model: %w", err)
		}
		// TODO: Check if state is "start" and the (reconstructed or provided) spec is available

	case *ResetModelCommand:
		state, stateValue, err := l.checkBFLAState(cmd.apiID, BFLAStart, BFLALearning, BFLALearnt, BFLADetecting)
		if err != nil {
			return fmt.Errorf("unable to perform command 'Reset Model': %w", err)
		}
		if state.State == BFLAStart {
			break
		}
		if state.State == BFLADetecting || state.State == BFLALearning {
			err = l.bflaBackendAccessor.DisableTraces(ctx, l.modName, cmd.apiID)
			if err != nil {
				return fmt.Errorf("cannot disable traces: %w", err)
			}
		}

		// Set existing auth model to empty
		authzModel, err := l.authzModelsMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get authz model state: %w", err)
		}
		authzModel.Set(AuthorizationModel{})
		state.State = BFLAStart
		stateValue.Set(state)
		if err := l.findingsRegistry.Clear(cmd.apiID); err != nil {
			return fmt.Errorf("unable to get authz model state: %w", err)
		}
		l.logError(l.notifyAuthzModel(ctx, cmd.apiID))

	case *StartDetectionCommand:
		state, stateValue, err := l.checkBFLAState(cmd.apiID, BFLALearning, BFLALearnt)
		if err != nil {
			return fmt.Errorf("unable to perform command 'Start Detection': %w", err)
		}
		if state.State == BFLALearnt {
			err = l.bflaBackendAccessor.EnableTraces(ctx, l.modName, cmd.apiID)
			if err != nil {
				return fmt.Errorf("cannot enable traces: %w", err)
			}
		}
		state.State = BFLADetecting
		stateValue.Set(state)

	case *StopDetectionCommand:
		state, stateValue, err := l.checkBFLAState(cmd.apiID, BFLADetecting)
		if err != nil {
			return fmt.Errorf("unable to perform command 'Stop Detection': %w", err)
		}
		err = l.bflaBackendAccessor.DisableTraces(ctx, l.modName, cmd.apiID)
		if err != nil {
			return fmt.Errorf("cannot disable traces: %w", err)
		}
		state.State = BFLALearnt
		stateValue.Set(state)
		l.logError(l.notifyAuthzModel(ctx, cmd.apiID))

	case *ProvideAuthzModelCommand:
		_, _, err = l.checkBFLAState(cmd.apiID, BFLALearnt, BFLADetecting)
		if err != nil {
			return fmt.Errorf("unable to perform command 'Provide Authz Model': %w", err)
		}
		pv, err := l.authzModelsMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get authz model state: %w", err)
		}
		authzModel, err := l.validateAuthzModel(ctx, cmd.authzModel, cmd.apiID)
		if err != nil {
			return fmt.Errorf("invalid authorization model provided: %w", err)
		}
		pv.Set(authzModel)
	}
	if err != nil {
		return fmt.Errorf("error when trying to update the authz model: %w", err)
	}

	if err = l.statePersister.Persist(ctx); err != nil {
		return fmt.Errorf("unable to persist the new state: %w", err)
	}

	log.Info("bfla synced for authz model")
	return nil
}

func (l *learnAndDetectBFLA) validateAuthzModel(ctx context.Context, m AuthorizationModel, apiID uint) (AuthorizationModel, error) {
	jmodel, err := json.Marshal(m)
	if err != nil {
		log.Errorf("unable to marshal auth model %v", err)
	}
	log.Debugf("updated auth model:%s\n", jmodel)
	apiInfo, err := l.bflaBackendAccessor.GetAPIInfo(ctx, apiID)
	if err != nil {
		return m, fmt.Errorf("unable to get api info: %w", err)
	}
	tags, err := ParseSpecInfo(apiInfo)
	if err != nil {
		return m, fmt.Errorf("unable to parse spec info: %w", err)
	}
	specType := SpecTypeFromAPIInfo(apiInfo)
	if specType == SpecTypeNone {
		return m, fmt.Errorf("spec not present cannot process authorization model; apiID=%d", apiID)
	}
	for _, op := range m.Operations {
		if op.Path == "" || op.Method == "" {
			return m, fmt.Errorf("invalid auth model operation: %v. apiID=%d", op, apiID)
		}
		op.Tags = resolveTagsForPathAndMethod(tags, op.Path, op.Method)

		for audIdx, aud := range op.Audience {
			if !aud.External {
				if aud.K8sObject == nil ||
					aud.K8sObject.Name == "" ||
					aud.K8sObject.Uid == "" ||
					aud.K8sObject.ApiVersion == "" ||
					aud.K8sObject.Kind == "" {
					return m, fmt.Errorf("invalid auth model audience for operation %v: [%d] %v . apiID = %d", op, audIdx, aud, apiID)
				}
			}
			if aud.Authorized {
				aud.WarningStatus = restapi.LEGITIMATE
			} else if aud.StatusCode < 200 || aud.StatusCode > 299 {
				aud.WarningStatus = restapi.SUSPICIOUSHIGH
			} else {
				aud.WarningStatus = restapi.SUSPICIOUSMEDIUM
			}
		}
	}
	return m, nil
}

func GetSpecOperation(spc *spec.Swagger, method models.HTTPMethod, resolvedPath string) *spec.Operation {
	switch method {
	case models.HTTPMethodGET:
		return spc.Paths.Paths[resolvedPath].Get
	case models.HTTPMethodHEAD:
		return spc.Paths.Paths[resolvedPath].Head
	case models.HTTPMethodPOST:
		return spc.Paths.Paths[resolvedPath].Post
	case models.HTTPMethodPUT:
		return spc.Paths.Paths[resolvedPath].Put
	case models.HTTPMethodDELETE:
		return spc.Paths.Paths[resolvedPath].Delete
	case models.HTTPMethodCONNECT:
		// op = spc.Paths.Paths[resolvedPath].Connect TODO
	case models.HTTPMethodOPTIONS:
		return spc.Paths.Paths[resolvedPath].Options
	case models.HTTPMethodTRACE:
		// op = spc.Paths.Paths[resolvedPath].Trace TODO
	case models.HTTPMethodPATCH:
		return spc.Paths.Paths[resolvedPath].Patch
	}
	return nil
}

func ContainsAll(items []string, vals []string) bool {
	for _, item := range items {
		if !Contains(vals, item) {
			return false
		}
	}
	return true
}

func Contains(items []string, val string) bool {
	for _, item := range items {
		if val == item {
			return true
		}
	}
	return false
}

//nolint: gocyclo
func (l *learnAndDetectBFLA) traceRunner(ctx context.Context, trace *CompositeTrace) (err error) {
	defer runtimeRecover()
	defer l.statePersister.AckSubmit(trace.APIEvent.ID)
	apiID := trace.APIEvent.APIInfoID
	log.Infof("bfla received event: %d", apiID)
	// load the model from store in the case it's not already present in memory or don't do anything if the model with id does not exist
	apiInfo, err := l.bflaBackendAccessor.GetAPIInfo(ctx, apiID)
	if err != nil {
		return fmt.Errorf("unable to get api info: %w", err)
	}
	tags, err := ParseSpecInfo(apiInfo)
	if err != nil {
		return fmt.Errorf("unable to parse spec info: %w", err)
	}
	resolvedPath, err := ResolvePath(tags, trace.APIEvent)
	if err != nil {
		return fmt.Errorf("unable to process trace: %w", err)
	}

	specType := SpecTypeFromAPIInfo(apiInfo)
	if specType == SpecTypeNone {
		return fmt.Errorf("spec not present cannot learn BFLA; apiID=%d", trace.APIEvent.APIInfoID)
	}

	var (
		state      BFLAState
		stateValue recovery.PersistedValue
	)

	if l.config.AutomaticLearningAndDetection {
		state, stateValue, err = l.checkBFLAState(apiID, BFLAStart, BFLALearning, BFLADetecting)
	} else {
		state, stateValue, err = l.checkBFLAState(apiID, BFLALearning, BFLADetecting)
	}
	if err != nil {
		return fmt.Errorf("unable to handle traces in the current state: %w", err)
	}
	var srcUID string
	if trace.K8SSource != nil {
		srcUID = trace.K8SSource.Uid
	}

	switch state.State {
	case BFLAStart:
		/* we are in the automatic learning and detection: start learning */
		/* In this case we start learning automatically. This at the moment only works when selective tracing is disabled, otherwise we never receive the first trace */
		/* TODO: Enable trace for BFLA as soon as the spec is added */
		state.TraceCounter = int(l.config.LearningNrTraces)
		state.State = BFLALearning
		err = l.bflaBackendAccessor.EnableTraces(ctx, l.modName, trace.APIEvent.APIInfoID)
		if err != nil {
			return fmt.Errorf("cannot enable traces: %w", err)
		}
		stateValue.Set(state)
		fallthrough
	case BFLALearning:
		/* We are in the learning state */
		log.Debugf("api %d; To process: %d", trace.APIEvent.APIInfoID, state.TraceCounter)
		err := l.updateAuthorizationModel(tags, resolvedPath, string(trace.APIEvent.Method),
			trace.K8SSource, trace.APIEvent.APIInfoID, trace.DetectedUser, true, false)
		if err != nil {
			return err
		}
		if state.TraceCounter != -1 {
			state.TraceCounter--

			if state.TraceCounter == 0 {
				if l.config.AutomaticLearningAndDetection {
					/* switch directly to detecting */
					state.State = BFLADetecting
				} else {
					state.State = BFLALearnt
					err = l.bflaBackendAccessor.DisableTraces(ctx, l.modName, apiID)
					if err != nil {
						log.Errorf("cannot disable traces: %v", err)
					}
				}
			}
			stateValue.Set(state)
		}

		aud, setAud, err := l.findSourceObj(resolvedPath, string(trace.APIEvent.Method), srcUID, trace.APIEvent.APIInfoID)
		if err != nil {
			return fmt.Errorf("unable to find source obj: %w", err)
		}
		aud.StatusCode = trace.APIEvent.StatusCode
		aud.LastTime = time.Time(trace.APIEvent.Time)
		setAud(aud)
		return err
	case BFLADetecting:
		/* We are in detecting state */
		if err := l.updateAuthorizationModel(tags, resolvedPath, string(trace.APIEvent.Method),
			trace.K8SSource, trace.APIEvent.APIInfoID, trace.DetectedUser, false, false); err != nil {
			return err
		}
		aud, setAud, err := l.findSourceObj(resolvedPath, string(trace.APIEvent.Method), srcUID, trace.APIEvent.APIInfoID)
		if err != nil {
			return fmt.Errorf("unable to find source obj: %w", err)
		}
		findingsUpdated := false
		if !aud.Authorized {
			// updates the auth model but this time as unauthorized
			var severity core.AlertSeverity
			var finding common.APIFinding
			code := trace.APIEvent.StatusCode
			if 200 > code || code > 299 {
				severity = core.AlertInfo
				finding = APIFindingBFLASuspiciousCallMedium(specType, resolvedPath, trace.APIEvent.Method)
			} else {
				severity = core.AlertWarn
				finding = APIFindingBFLASuspiciousCallHigh(specType, resolvedPath, trace.APIEvent.Method)
			}

			findingsUpdated, err = l.findingsRegistry.Add(apiID, finding)
			if err != nil {
				log.Warnf("unable to add findings: %s", err)
			}
			if err := l.eventAlerter.SetEventAlert(ctx, l.modName, trace.APIEvent.ID, severity); err != nil {
				log.Warnf("unable to set alert annotation: %s", err)
			}

			// l.logError(l.notifyAuthzModel(ctx, trace.APIEvent.APIInfoID))
			aud.WarningStatus = ResolveBFLAStatusInt(int(trace.APIEvent.StatusCode))
		}

		spc, err := GetOpenAPI(apiInfo, apiID)
		if err != nil {
			log.Warnf("unable to get openapi spec")
		}
		op := GetSpecOperation(spc, trace.APIEvent.Method, resolvedPath)
		for _, user := range aud.EndUsers {
			if user.IsMismatchedScopes(op) {
				updated, err := l.findingsRegistry.Add(trace.APIEvent.APIInfoID, APIFindingBFLAScopesMismatch(specType, resolvedPath, trace.APIEvent.Method))
				if err != nil {
					log.Warnf("unable to add scope mismatch findings: %s", err)
				}
				findingsUpdated = findingsUpdated || updated
			}
		}
		if findingsUpdated {
			err = l.notifyFindings(ctx, apiID)
			if err != nil {
				log.Errorf("unable to send finding notification: %v", err)
			}
		}
		aud.StatusCode = trace.APIEvent.StatusCode
		aud.LastTime = time.Time(trace.APIEvent.Time)
		setAud(aud)

		return nil
	case BFLALearnt:
		return fmt.Errorf("illegal state %s", state.State)
	}
	return nil
}

func (l *learnAndDetectBFLA) notifyFindings(ctx context.Context, apiID uint) error {
	findings, err := l.findingsRegistry.GetAll(apiID)
	if err != nil {
		log.Errorf("unable to retrieve findings for API %v: %v", apiID, err)
	}
	ntf := notifications.ApiFindingsNotification{}
	ntf.Items = &findings

	return l.bflaNotifier.NotifyFindings(ctx, apiID, ntf) //nolint:wrapcheck
}

func (l *learnAndDetectBFLA) notifyAuthzModel(ctx context.Context, apiID uint) error {
	ntf := AuthzModelNotification{}

	if l.isLearning(apiID) {
		ntf.Learning = true
	} else {
		v, err := l.authzModelsMap.Get(apiID)
		if err != nil {
			return fmt.Errorf("unable to geet authz model %w", err)
		}
		if !v.Exists() {
			return fmt.Errorf("authorization model not found")
		}

		apiInfo, err := l.bflaBackendAccessor.GetAPIInfo(ctx, apiID)
		if err != nil {
			return fmt.Errorf("unable to get api info: %w", err)
		}
		specType := SpecTypeFromAPIInfo(apiInfo)
		ntf.SpecType = specType
		if specType != SpecTypeNone {
			ntf.AuthzModel, _ = v.Get().(AuthorizationModel)
		}
	}
	return l.bflaNotifier.NotifyAuthzModel(ctx, apiID, ntf) //nolint:wrapcheck
}

func (l *learnAndDetectBFLA) initAuthorizationModel(apiID uint) error {
	authzModelEntry, err := l.authzModelsMap.Get(apiID)
	if err != nil {
		return fmt.Errorf("unable to get authz model state: %w", err)
	}
	if authzModelEntry.Exists() {
		return nil
	}
	authzModel := AuthorizationModel{}
	authzModelEntry.Set(authzModel)
	return nil
}

func (l *learnAndDetectBFLA) updateAuthorizationModel(tags []*models.SpecTag, path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, user *DetectedUser, authorize, updateAuthorized bool) error {
	log.Debugf("Update auth model: tags = %v, path = %s, method=%s, apidId=%d", tags, path, method, apiID)

	if path == "" || method == "" {
		return fmt.Errorf("unable to update Authorization Model. Invalid method %s or path %s", method, path)
	}
	external := clientRef == nil
	authzModelEntry, err := l.authzModelsMap.Get(apiID)
	if err != nil {
		return fmt.Errorf("unable to get authz model state: %w", err)
	}
	authzModel, _ := authzModelEntry.Get().(AuthorizationModel)
	if !authzModelEntry.Exists() {
		authzModel = AuthorizationModel{
			Operations: []*Operation{{
				Method:   method,
				Path:     path,
				Tags:     resolveTagsForPathAndMethod(tags, path, method),
				Audience: []*SourceObject{{External: external, K8sObject: clientRef, Authorized: authorize, WarningStatus: restapi.LEGITIMATE}},
			}},
		}
		if user != nil {
			authzModel.Operations[0].Audience[0].EndUsers = append(authzModel.Operations[0].Audience[0].EndUsers, user)
		}
		authzModelEntry.Set(authzModel)
		return nil
	}

	opIndex, op := authzModel.Operations.Find(func(op *Operation) bool {
		return op.Method == method && op.Path == path
	})
	if op == nil {
		op = &Operation{
			Method:   method,
			Path:     path,
			Tags:     resolveTagsForPathAndMethod(tags, path, method),
			Audience: []*SourceObject{{External: external, K8sObject: clientRef, Authorized: authorize, WarningStatus: restapi.LEGITIMATE}},
		}
		if user != nil {
			op.Audience[0].EndUsers = append(op.Audience[0].EndUsers, user)
		}
		authzModel.Operations = append(authzModel.Operations, op)
		log.Debugf("Setting authModel %v", authzModel)
		authzModelEntry.Set(authzModel)
		return nil
	}

	audienceIndex, audience := op.Audience.Find(func(sa *SourceObject) bool {
		if external {
			return sa.External
		}
		if sa.External {
			return external
		}
		return sa.K8sObject.Uid == clientRef.Uid
	})
	if audience == nil {
		sa := &SourceObject{External: external, K8sObject: clientRef, Authorized: authorize, WarningStatus: restapi.LEGITIMATE}
		if user != nil {
			sa.EndUsers = append(sa.EndUsers, user)
		}
		op.Audience = append(op.Audience, sa)
		authzModelEntry.Set(authzModel)
		return nil
	}

	if user != nil {
		if _, endUser := audience.EndUsers.Find(func(u *DetectedUser) bool {
			return u.ID == user.ID
		}); endUser == nil {
			audience.EndUsers = append(audience.EndUsers, user)
			authzModelEntry.Set(authzModel)
		}
	}

	// TODO think of a prettier way to be able to update only on certain cases
	if updateAuthorized {
		oldAuthorized := audience.Authorized
		authzModel.Operations[opIndex].Audience[audienceIndex].Authorized = authorize
		if oldAuthorized != authorize {
			authzModelEntry.Set(authzModel)
			return nil
		}
	}

	authzModelEntry.Set(authzModel)
	return nil
}

func (l *learnAndDetectBFLA) IsLearning(apiID uint) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.isLearning(apiID)
}

func (l *learnAndDetectBFLA) GetState(apiID uint) (BFLAStateEnum, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	state, err := l.getState(apiID)
	if err != nil {
		return BFLAStart, err
	}
	return state.State, err
}

func (l *learnAndDetectBFLA) isLearning(apiID uint) bool {
	_, _, err := l.checkBFLAState(apiID, BFLALearning)
	return err == nil
}

func (l *learnAndDetectBFLA) getState(apiID uint) (BFLAState, error) {
	stateValue, err := l.bflaStateMap.Get(apiID)
	if err != nil {
		return BFLAState{}, fmt.Errorf("unable to get state traces counter: %w", err)
	}
	state := stateValue.Get().(BFLAState) //nolint:forcetypeassert
	log.Debugf("current state for api %d is %v", apiID, state.State)
	return state, nil
}

func (l *learnAndDetectBFLA) FindSourceObj(path, method, clientUID string, apiID uint) (*SourceObject, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	aud, _, err := l.findSourceObj(path, method, clientUID, apiID)
	return aud, err
}

func (l *learnAndDetectBFLA) findSourceObj(path, method, clientUID string, apiID uint) (obj *SourceObject, setFn func(v *SourceObject), err error) {
	external := clientUID == ""
	authzModelEntry, err := l.authzModelsMap.Get(apiID)
	if err != nil {
		return nil, nil, fmt.Errorf("authz model load error: %w", err)
	}
	authzModel, _ := authzModelEntry.Get().(AuthorizationModel)
	_, op := authzModel.Operations.Find(func(op *Operation) bool {
		return op.Path == path &&
			op.Method == method
	})
	if op == nil {
		return nil, nil, fmt.Errorf("operation not found: %w", err)
	}
	audIndex, obj := op.Audience.Find(func(sa *SourceObject) bool {
		if external {
			return sa.External
		}
		return sa.K8sObject.Uid == clientUID
	})
	if obj == nil {
		return nil, nil, fmt.Errorf("audience not found: %w", err)
	}

	return obj, func(v *SourceObject) {
		op.Audience[audIndex] = v
		authzModelEntry.Set(authzModel)
	}, nil
}

func (l *learnAndDetectBFLA) SendTrace(trace *CompositeTrace) {
	l.tracesCh <- trace
}

func (l *learnAndDetectBFLA) ApproveTrace(path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, user *DetectedUser) {
	l.commandsCh.Send(&MarkLegitimateCommand{
		detectedUser: user,
		path:         path,
		method:       method,
		clientRef:    clientRef,
		apiID:        apiID,
	})
}

func (l *learnAndDetectBFLA) DenyTrace(path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, user *DetectedUser) {
	l.commandsCh.Send(&MarkIllegitimateCommand{
		detectedUser: user,
		path:         path,
		method:       method,
		clientRef:    clientRef,
		apiID:        apiID,
	})
}

func (l *learnAndDetectBFLA) ResetModel(apiID uint) error {
	return l.commandsCh.SendAndReplyErr(&ResetModelCommand{
		apiID:     apiID,
		ErrorChan: NewErrorChan(),
	})
}

func (l *learnAndDetectBFLA) StopLearning(apiID uint) error {
	return l.commandsCh.SendAndReplyErr(&StopLearningCommand{
		apiID:     apiID,
		ErrorChan: NewErrorChan(),
	})
}

func (l *learnAndDetectBFLA) StartLearning(apiID uint, numberOfTraces int) error {
	if numberOfTraces < -1 {
		return fmt.Errorf("value %v not allowed", numberOfTraces)
	}
	return l.commandsCh.SendAndReplyErr(&StartLearningCommand{
		apiID:          apiID,
		numberOfTraces: numberOfTraces,
		ErrorChan:      NewErrorChan(),
	})
}

func (l *learnAndDetectBFLA) StartDetection(apiID uint) error {
	return l.commandsCh.SendAndReplyErr(&StartDetectionCommand{
		apiID:     apiID,
		ErrorChan: NewErrorChan(),
	})
}

func (l *learnAndDetectBFLA) StopDetection(apiID uint) error {
	return l.commandsCh.SendAndReplyErr(&StopDetectionCommand{
		apiID:     apiID,
		ErrorChan: NewErrorChan(),
	})
}

func (l *learnAndDetectBFLA) ProvideAuthzModel(apiID uint, am AuthorizationModel) {
	l.commandsCh.Send(&ProvideAuthzModelCommand{
		apiID:      apiID,
		authzModel: am,
	})
}

func ResolveBFLAStatus(statusCode string) restapi.BFLAStatus {
	code, err := strconv.Atoi(statusCode)
	if err == nil {
		return ResolveBFLAStatusInt(code)
	}

	return restapi.SUSPICIOUSHIGH
}

func ResolveBFLAStatusInt(code int) restapi.BFLAStatus {
	if 200 > code || code > 299 {
		return restapi.SUSPICIOUSMEDIUM
	}

	return restapi.SUSPICIOUSHIGH
}

type SpecType uint

const (
	SpecTypeNone SpecType = iota
	SpecTypeProvided
	SpecTypeReconstructed
)

func SpecTypeFromAPIInfo(apiinfo *database.APIInfo) SpecType {
	if apiinfo.HasProvidedSpec {
		return SpecTypeProvided
	}
	if apiinfo.HasReconstructedSpec {
		return SpecTypeReconstructed
	}
	return SpecTypeNone
}
