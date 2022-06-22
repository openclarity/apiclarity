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
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/go-openapi/spec"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/k8straceannotator"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/recovery"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/restapi"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
)

const (
	ModuleName               = "bfla"
	K8sSrcAnnotationName     = "bfla_k8s_src"
	K8sDstAnnotationName     = "bfla_k8s_dst"
	DetectedIDAnnotationName = "bfla_detected_id"

	AuthzModelAnnotationName           = "authz_model"
	AuthzProcessedTracesAnnotationName = "authz_processed_traces"
	AuthzTracesToLearnAnnotationName   = "authz_traces_to_learn"
)

var ErrUnsupportedAuthScheme = errors.New("unsupported auth scheme")

func NewBFLADetector(ctx context.Context, apiInfoProvider apiInfoProvider, eventAlerter EventAlerter, sp recovery.StatePersister) BFLADetector {
	l := &learnAndDetectBFLA{
		tracesCh:         make(chan *CompositeTrace),
		commandsCh:       make(chan Command),
		errCh:            make(chan error),
		apiInfoProvider:  apiInfoProvider,
		authzModelsMap:   recovery.NewPersistedMap(sp, AuthzModelAnnotationName, reflect.TypeOf(AuthorizationModel{})),
		tracesCounterMap: recovery.NewPersistedMap(sp, AuthzProcessedTracesAnnotationName, reflect.TypeOf(1)),
		statePersister:   sp,
		eventAlerter:     eventAlerter,
		mu:               &sync.RWMutex{},
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
	FindSourceObj(path, method, clientUid string, apiID uint) (*SourceObject, error)

	ApproveTrace(path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, user *DetectedUser)
	DenyTrace(path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, user *DetectedUser)

	ResetLearning(apiID uint, numberOfTraces int)
	StartLearning(apiID uint, numberOfTraces int)
	StopLearning(apiID uint)

	ProvideAuthzModel(apiID uint, am AuthorizationModel)
}

type apiInfoProvider interface {
	GetAPIInfo(ctx context.Context, apiID uint) (*database.APIInfo, error)
}

type Command interface{ isCommand() }

type StopLearningCommand struct {
	apiID uint
}

type StartLearningCommand struct {
	apiID          uint
	numberOfTraces int
}

type ResetLearningCommand struct {
	apiID          uint
	numberOfTraces int
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

func (a *StopLearningCommand) isCommand()      {}
func (a *StartLearningCommand) isCommand()     {}
func (a *ResetLearningCommand) isCommand()     {}
func (a *MarkLegitimateCommand) isCommand()    {}
func (a *MarkIllegitimateCommand) isCommand()  {}
func (a *ProvideAuthzModelCommand) isCommand() {}

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
	tracesCh        chan *CompositeTrace
	commandsCh      chan Command
	errCh           chan error
	apiInfoProvider apiInfoProvider

	authzModelsMap   recovery.PersistedMap
	tracesCounterMap recovery.PersistedMap

	statePersister recovery.StatePersister

	eventAlerter EventAlerter
	mu           *sync.RWMutex
}

type CompositeTrace struct {
	*core.Event

	K8SSource, K8SDestination *k8straceannotator.K8sObjectRef
	DetectedUser              *DetectedUser
}

func (l *learnAndDetectBFLA) run(ctx context.Context) {
	defer log.Info("ending learnFromTracesAndDetectBFLA")
	for {
		select {
		case feedback, ok := <-l.commandsCh:
			if ok {
				if err := l.commandsRunner(ctx, feedback); err != nil {
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

func (l *learnAndDetectBFLA) commandsRunner(ctx context.Context, command Command) (err error) {
	defer runtimeRecover()
	switch cmd := command.(type) {
	case *MarkLegitimateCommand:
		err = l.updateAuthorizationModel(cmd.path, cmd.method, cmd.clientRef, cmd.apiID, cmd.detectedUser, true, true)
	case *MarkIllegitimateCommand:
		err = l.updateAuthorizationModel(cmd.path, cmd.method, cmd.clientRef, cmd.apiID, cmd.detectedUser, false, true)
	case *StopLearningCommand:
		counter, err := l.tracesCounterMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get state traces counter: %w", err)
		}

		counter.Set(0)
	case *StartLearningCommand:
		tracesToProcess, err := l.tracesCounterMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get state traces counter: %w", err)
		}
		if _, ok := l.mustLearn(cmd.apiID); ok {
			log.Warn("won't start learning, because the learning has already started")
			return nil
		}

		tracesToProcess.Set(cmd.numberOfTraces)
	case *ResetLearningCommand:
		counter, err := l.tracesCounterMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get state traces counter: %w", err)
		}
		counter.Set(cmd.numberOfTraces)

		// Set existing auth model to empty
		authzModel, err := l.authzModelsMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get authz model state: %w", err)
		}
		authzModel.Set(AuthorizationModel{})

	case *ProvideAuthzModelCommand:
		pv, err := l.authzModelsMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get state traces to learn: %w", err)
		}
		pv.Set(cmd.authzModel)

		// stop learning
		counter, err := l.tracesCounterMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get state traces counter: %w", err)
		}

		toLearn, err := l.tracesToLearnMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get state traces to learn: %w", err)
		}
		toLearn.Set(counter.Get())
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
		//op = spc.Paths.Paths[resolvedPath].Connect TODO
	case models.HTTPMethodOPTIONS:
		return spc.Paths.Paths[resolvedPath].Options
	case models.HTTPMethodTRACE:
		//op = spc.Paths.Paths[resolvedPath].Trace TODO
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

func (l *learnAndDetectBFLA) traceRunner(ctx context.Context, trace *CompositeTrace) (err error) {
	defer runtimeRecover()
	defer l.statePersister.AckSubmit(trace.APIEvent.ID)
	apiID := trace.APIEvent.APIInfoID
	log.Infof("bfla received event: %d", apiID)
	// load the model from store in case the it is not already present in memory or don't do anything if the model with id does not exist
	apiInfo, err := l.apiInfoProvider.GetAPIInfo(ctx, apiID)
	if err != nil {
		return fmt.Errorf("unable to get api info: %w", err)
	}
	resolvedPath := ResolvePath(apiInfo, trace.APIEvent)

	if SpecTypeFromAPIInfo(apiInfo) == SpecTypeNone {
		return fmt.Errorf("spec not present cannot learn BFLA")
	}
	var tracesProcessed int
	tracesProcessedEntry, err := l.tracesCounterMap.Get(apiID)
	if err != nil {
		log.Warnf("Could not load processed traces number: %s", err)
	} else {
		tracesProcessed, _ = tracesProcessedEntry.Get().(int)
	}

	if decrement, ok := l.mustLearn(apiID); ok {
		log.Debugf("api %d; processed: %d", trace.APIEvent.APIInfoID, tracesProcessed)
		// to still learn
		err := l.updateAuthorizationModel(resolvedPath, string(trace.APIEvent.Method),
			trace.K8SSource, trace.APIEvent.APIInfoID, trace.DetectedUser, true, false)
		if err != nil {
			return err
		}

		decrement()
		return nil
	}
	if err := l.updateAuthorizationModel(resolvedPath, string(trace.APIEvent.Method), trace.K8SSource, trace.APIEvent.APIInfoID, trace.DetectedUser, false, false); err != nil {
		return err
	}
	var srcUid string
	if trace.K8SSource != nil {
		srcUid = trace.K8SSource.Uid
	}
	aud, setAud, err := l.findSourceObj(resolvedPath, string(trace.APIEvent.Method), srcUid, trace.APIEvent.APIInfoID)
	if err != nil {
		return err
	}
	aud.WarningStatus = restapi.LEGITIMATE
	if !aud.Authorized {
		// updates the auth model but this time as unauthorized
		severity := core.AlertWarn
		code := trace.APIEvent.StatusCode
		if 200 > code || code > 299 {
			severity = core.AlertInfo
		}

		if err := l.eventAlerter.SetEventAlert(ctx, ModuleName, trace.APIEvent.ID, severity); err != nil {
			return fmt.Errorf("unable to set alert annotation: %w", err)
		}
		aud.WarningStatus = ResolveBFLAStatusInt(int(trace.APIEvent.StatusCode))
	}
	aud.StatusCode = trace.APIEvent.StatusCode
	aud.LastTime = time.Time(trace.APIEvent.Time)
	setAud(aud)
	return nil
}

func (l *learnAndDetectBFLA) mustLearn(apiID uint) (decrementFn func(), ok bool) {
	tracesToLearn, err := l.tracesCounterMap.Get(apiID)
	if err != nil {
		log.Error("load traces to learn error: ", err)
		return nil, false
	}

	tracesInt, _ := tracesToLearn.Get().(int)
	if !tracesToLearn.Exists() {
		return nil, false
	}
	return func() {
		tracesInt--
		tracesToLearn.Set(tracesInt)
	}, tracesInt > 0 || tracesInt == -1
}

func (l *learnAndDetectBFLA) updateAuthorizationModel(path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, user *DetectedUser, authorize, updateAuthorized bool) error {
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
				Audience: []*SourceObject{{External: external, K8sObject: clientRef, Authorized: authorize}},
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
			Audience: []*SourceObject{{External: external, K8sObject: clientRef, Authorized: authorize}},
		}
		if user != nil {
			op.Audience[0].EndUsers = append(op.Audience[0].EndUsers, user)
		}
		authzModel.Operations = append(authzModel.Operations, op)
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
		sa := &SourceObject{External: external, K8sObject: clientRef, Authorized: authorize}
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
	_, ok := l.mustLearn(apiID)
	return ok
}

func (l *learnAndDetectBFLA) FindSourceObj(path, method, clientUid string, apiID uint) (*SourceObject, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	aud, _, err := l.findSourceObj(path, method, clientUid, apiID)
	return aud, err
}

func (l *learnAndDetectBFLA) findSourceObj(path, method, clientUid string, apiID uint) (obj *SourceObject, setFn func(v *SourceObject), err error) {
	external := clientUid == ""
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
		if sa.External == external {
			return true
		}
		if sa.External && !external {
			return false
		}
		return sa.K8sObject.Uid == clientUid
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
	l.commandsCh <- &MarkLegitimateCommand{
		detectedUser: user,
		path:         path,
		method:       method,
		clientRef:    clientRef,
		apiID:        apiID,
	}
}

func (l *learnAndDetectBFLA) DenyTrace(path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, user *DetectedUser) {
	l.commandsCh <- &MarkIllegitimateCommand{
		detectedUser: user,
		path:         path,
		method:       method,
		clientRef:    clientRef,
		apiID:        apiID,
	}
}

func (l *learnAndDetectBFLA) ResetLearning(apiID uint, numberOfTraces int) {
	if numberOfTraces < -1 {
		log.Errorf("value %v not allowed", numberOfTraces)
		return
	}
	l.commandsCh <- &ResetLearningCommand{
		apiID:           apiID,
		numberOfTraces: numberOfTraces,
	}
}

func (l *learnAndDetectBFLA) StopLearning(apiID uint) {
	l.commandsCh <- &StopLearningCommand{
		apiID: apiID,
	}
}

func (l *learnAndDetectBFLA) StartLearning(apiID uint, numberOfTraces int) {
	if numberOfTraces < -1 {
		log.Errorf("value %v not allowed", numberOfTraces)
		return
	}
	l.commandsCh <- &StartLearningCommand{
		apiID:          apiID,
		numberOfTraces: numberOfTraces,
	}
}

func (l *learnAndDetectBFLA) ProvideAuthzModel(apiID uint, am AuthorizationModel) {
	l.commandsCh <- &ProvideAuthzModelCommand{
		apiID:      apiID,
		authzModel: am,
	}
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
