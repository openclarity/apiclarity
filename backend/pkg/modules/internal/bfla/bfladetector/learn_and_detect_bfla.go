package bfladetector

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/bfla/k8straceannotator"
	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/bfla/recovery"
	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/bfla/restapi"
	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/core"
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

func NewBFLADetector(ctx context.Context, learnTracesNr int, pathResolver PathResolver, eventAlerter EventAlerter, sp recovery.StatePersister) BFLADetector {
	l := &learnAndDetectBFLA{
		tracesCh:             make(chan *CompositeTrace),
		commandsCh:           make(chan Command),
		errCh:                make(chan error),
		pathResolver:         pathResolver,
		defaultTracesToLearn: learnTracesNr,
		authzModelsMap:       recovery.NewPersistedMap(sp, AuthzModelAnnotationName, reflect.TypeOf(AuthorizationModel{})),
		tracesCounterMap:     recovery.NewPersistedMap(sp, AuthzProcessedTracesAnnotationName, reflect.TypeOf(1)),
		tracesToLearnMap:     recovery.NewPersistedMap(sp, AuthzTracesToLearnAnnotationName, reflect.TypeOf(1)),
		statePersister:       sp,
		eventAlerter:         eventAlerter,
		mu:                   &sync.RWMutex{},
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
	IsUnauthorized(path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, user *DetectedUser) bool

	ApproveTrace(path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, user *DetectedUser)
	DenyTrace(path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, user *DetectedUser)
	ResetLearning(apiID uint, numberOfTraces int)
	StartLearning(apiID uint, numberOfTraces int)
	StopLearning(apiID uint)
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

func (a *StopLearningCommand) isCommand()     {}
func (a *StartLearningCommand) isCommand()    {}
func (a *ResetLearningCommand) isCommand()    {}
func (a *MarkLegitimateCommand) isCommand()   {}
func (a *MarkIllegitimateCommand) isCommand() {}

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
	tracesCh             chan *CompositeTrace
	commandsCh           chan Command
	errCh                chan error
	pathResolver         PathResolver
	defaultTracesToLearn int

	authzModelsMap   recovery.PersistedMap
	tracesCounterMap recovery.PersistedMap
	tracesToLearnMap recovery.PersistedMap

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
		resolvedPath, specType := l.pathResolver.RezolvePath(ctx, cmd.apiID, cmd.path)
		if specType == SpecTypeNone {
			log.Warnf("Spec not present, cannot learn BFLA")
			return nil
		}
		err = l.updateAuthorizationModel(resolvedPath, cmd.method, cmd.clientRef, cmd.apiID, cmd.detectedUser, true, true)
	case *MarkIllegitimateCommand:
		resolvedPath, specType := l.pathResolver.RezolvePath(ctx, cmd.apiID, cmd.path)
		if specType == SpecTypeNone {
			log.Warnf("Spec not present, cannot learn BFLA")
			return err
		}
		err = l.updateAuthorizationModel(resolvedPath, cmd.method, cmd.clientRef, cmd.apiID, cmd.detectedUser, false, true)
	case *StopLearningCommand:

		counter, err := l.tracesCounterMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get state traces counter: %w", err)
		}

		toLearn, err := l.tracesToLearnMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get state traces to learn: %w", err)
		}
		toLearn.Set(counter.Get())
	case *StartLearningCommand:
		tracesProcessed, err := l.tracesCounterMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get state traces counter: %w", err)
		}
		if tracesProcessed.Get().(int) < l.tracesToLearn(cmd.apiID) {
			log.Info("won't start learning, because the learning has already started")
			return nil
		}
		toLearn, err := l.tracesToLearnMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get state traces to learn: %w", err)
		}
		toLearn.Set(l.tracesToLearn(cmd.apiID) + cmd.numberOfTraces)
	case *ResetLearningCommand:
		toLearn, err := l.tracesToLearnMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get state traces to learn: %w", err)
		}
		toLearn.Set(cmd.numberOfTraces)

		counter, err := l.tracesCounterMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get state traces counter: %w", err)
		}
		counter.Set(0)

		// Set existing auth model to empty
		authzModel, err := l.authzModelsMap.Get(cmd.apiID)
		if err != nil {
			return fmt.Errorf("unable to get authz model state: %w", err)
		}
		authzModel.Set(AuthorizationModel{})
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

func (l *learnAndDetectBFLA) traceRunner(ctx context.Context, trace *CompositeTrace) (err error) {
	defer runtimeRecover()
	defer l.statePersister.AckSubmit(trace.APIEvent.ID)
	apiID := trace.APIEvent.APIInfoID
	log.Infof("bfla received event: %d", apiID)
	// load the model from store in case the it is not already present in memory or don't do anything if the model with id does not exist
	resolvedPath, specType := l.pathResolver.RezolvePath(ctx, apiID, trace.APIEvent.Path)
	if specType == SpecTypeNone {
		return fmt.Errorf("spec not present cannot learn BFLA")
	}
	var tracesProcessed int
	tracesProcessedEntry, err := l.tracesCounterMap.Get(apiID)
	if err != nil {
		log.Warnf("Could not load processed traces number: %s", err)
	} else {
		tracesProcessed, _ = tracesProcessedEntry.Get().(int)
	}

	if tracesProcessed < l.tracesToLearn(apiID) {
		log.Infof("api %d; processed: %d; traces to learn: %d", trace.APIEvent.APIInfoID, tracesProcessed, l.defaultTracesToLearn)
		// to still learn
		err := l.updateAuthorizationModel(resolvedPath, string(trace.APIEvent.Method),
			trace.K8SSource, trace.APIEvent.APIInfoID, trace.DetectedUser, true, false)
		if err != nil {
			return err
		}

		tracesProcessed++
		tracesProcessedEntry.Set(tracesProcessed)
		return nil
	}
	if err := l.updateAuthorizationModel(resolvedPath, string(trace.APIEvent.Method), trace.K8SSource, trace.APIEvent.APIInfoID, trace.DetectedUser, false, false); err != nil {
		return err
	}
	if l.isUnauthorized(resolvedPath, string(trace.APIEvent.Method),
		trace.K8SSource, trace.APIEvent.APIInfoID, trace.DetectedUser) {
		// updates the auth model but this time as unauthorized
		severity := core.AlertWarn
		code := trace.APIEvent.StatusCode
		if 200 > code || code > 299 {
			severity = core.AlertInfo
		}

		if err := l.eventAlerter.SetEventAlert(ctx, ModuleName, trace.APIEvent.ID, severity); err != nil {
			return fmt.Errorf("unable to set alert annotation: %w", err)
		}
	}
	return nil
}

func (l *learnAndDetectBFLA) tracesToLearn(apiID uint) int {
	tracesToLearn, err := l.tracesToLearnMap.Get(apiID)
	if err != nil {
		log.Error(err)
		return l.defaultTracesToLearn
	}
	if tracesToLearn.Exists() {
		return tracesToLearn.Get().(int)
	}
	return l.defaultTracesToLearn
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
		if sa.External == external {
			return true
		}
		if sa.External && !external {
			return false
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
	tracesProcessed, err := l.tracesCounterMap.Get(apiID)
	if err != nil {
		log.Error("load traces error: ", err)
		return true
	}
	if tracesProcessed.Get().(int) < l.tracesToLearn(apiID) {
		return true
	}
	return false
}

func (l *learnAndDetectBFLA) IsUnauthorized(path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, user *DetectedUser) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.isUnauthorized(path, method, clientRef, apiID, user)
}

func (l *learnAndDetectBFLA) isUnauthorized(path, method string, clientRef *k8straceannotator.K8sObjectRef, apiID uint, _ *DetectedUser) bool {
	var err error
	external := clientRef == nil
	authzModelEntry, err := l.authzModelsMap.Get(apiID)
	if err != nil {
		log.Error("authz model load error: ", err)
		return true
	}
	authzModel, _ := authzModelEntry.Get().(AuthorizationModel)
	_, op := authzModel.Operations.Find(func(op *Operation) bool {
		return op.Path == path &&
			op.Method == method
	})
	if op == nil {
		log.Error("operation not found", err)
		return true
	}
	_, aud := op.Audience.Find(func(sa *SourceObject) bool {
		if sa.External == external {
			return true
		}
		if sa.External && !external {
			return false
		}
		return sa.K8sObject.Uid == clientRef.Uid
	})
	if aud == nil {
		log.Error("audience not found", err)
		return true
	}

	return !aud.Authorized
}

type GenericOpenapiSpec struct {
	Paths map[string]*Path `yaml:"paths"`
}

type Path struct {
	Ref         string                    `yaml:"$ref,omitempty"`
	Summary     string                    `yaml:"summary,omitempty"`
	Description string                    `yaml:"description,omitempty"`
	Servers     interface{}               `yaml:"servers,omitempty"`
	Operations  map[string]*HasParameters `yaml:",inline"`
	Parameters  []*Parameter              `yaml:"parameters,omitempty"`
}

type HasParameters struct {
	Parameters []*Parameter `yaml:"parameters,omitempty"`
}

type Parameter struct {
	Name string `yaml:"name,omitempty"`
	In   string `yaml:"in,omitempty"`
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
	l.commandsCh <- &ResetLearningCommand{
		apiID:          apiID,
		numberOfTraces: numberOfTraces,
	}
}

func (l *learnAndDetectBFLA) StopLearning(apiID uint) {
	l.commandsCh <- &StopLearningCommand{
		apiID: apiID,
	}
}

func (l *learnAndDetectBFLA) StartLearning(apiID uint, numberOfTraces int) {
	l.commandsCh <- &StartLearningCommand{
		apiID:          apiID,
		numberOfTraces: numberOfTraces,
	}
}

func ResolveBFLAStatus(statusCode string) restapi.BFLAStatus {
	code, err := strconv.Atoi(statusCode)
	if err == nil {
		return ResolveBFLAStatusInt(code)
	}

	return restapi.BFLAStatusSUSPICIOUSHIGH
}

func ResolveBFLAStatusInt(code int) restapi.BFLAStatus {
	if 200 > code || code > 299 {
		return restapi.BFLAStatusSUSPICIOUSMEDIUM
	}

	return restapi.BFLAStatusSUSPICIOUSHIGH
}
