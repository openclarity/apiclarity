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

package apiclarity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	oapicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/bfladetector"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/k8straceannotator"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/recovery"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/restapi"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	pluginsmodels "github.com/openclarity/apiclarity/plugins/api/server/models"
)

const (
	moduleVersion            = "0.0.0"
	persistenceInterval      = 5 * time.Second
	controllerResyncInterval = 5 * time.Second
)

type bfla struct {
	httpHandler  http.Handler
	bflaDetector bfladetector.BFLADetector
	k8s          k8straceannotator.K8sClient

	accessor core.BackendAccessor
}

func (p *bfla) Name() string              { return bfladetector.ModuleName }
func (p *bfla) HTTPHandler() http.Handler { return p.httpHandler }

func newModule(ctx context.Context, accessor core.BackendAccessor) (_ core.Module, err error) {
	p := &bfla{
		accessor: accessor,
	}
	if accessor.K8SClient() == nil {
		return nil, fmt.Errorf("ignoring bfla module due to missing kubernetes client")
	}

	p.k8s, err = k8straceannotator.NewK8sClient(accessor.K8SClient())
	if err != nil {
		return nil, fmt.Errorf("failed to init bfla module: %w", err)
	}

	sp := recovery.NewStatePersister(ctx, accessor, bfladetector.ModuleName, persistenceInterval)
	ctrlNotifier := bfladetector.NewControllerNotifier(accessor)
	p.bflaDetector = bfladetector.NewBFLADetector(ctx, accessor, eventAlerter{accessor}, ctrlNotifier, sp, controllerResyncInterval)

	handler := &httpHandler{
		bflaDetector:    p.bflaDetector,
		state:           sp,
		accessor:        accessor,
		openAPIProvider: bfladetector.NewBFLAOpenAPIProvider(accessor),
	}
	p.httpHandler = restapi.HandlerWithOptions(handler, restapi.ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + bfladetector.ModuleName})
	return p, nil
}

type eventAlerter struct {
	accessor core.BackendAccessor
}

func (e eventAlerter) SetEventAlert(ctx context.Context, modName string, eventID uint, severity core.AlertSeverity) (err error) {
	switch severity {
	case core.AlertInfo:
		err = e.accessor.CreateAPIEventAnnotations(ctx, modName, eventID, core.AlertInfoAnn)
	case core.AlertWarn:
		err = e.accessor.CreateAPIEventAnnotations(ctx, modName, eventID, core.AlertWarnAnn)
	case core.AlertCritical:
		err = e.accessor.CreateAPIEventAnnotations(ctx, modName, eventID, core.AlertCriticalAnn)
	default:
		return fmt.Errorf("unexpected severity")
	}
	if err != nil {
		return fmt.Errorf("error creating an alert: %w", err)
	}
	return nil
}

func (p *bfla) EventNotify(ctx context.Context, event *core.Event) {
	if err := p.eventNotify(ctx, event); err != nil {
		log.Errorf("[BFLA] EventNotify: %s", err)
	}
}

func (p *bfla) eventNotify(ctx context.Context, event *core.Event) (err error) {
	log.Debugf("[BFLA] received a new event for API(%v) Event(%v) ", event.APIEvent.APIInfoID, event.APIEvent.ID)
	cmpTrace := &bfladetector.CompositeTrace{Event: event}
	if cmpTrace.K8SDestination, cmpTrace.K8SSource, cmpTrace.DetectedUser, err = getBFLAAnnotations(ctx, p.accessor, event.APIEvent.ID); err != nil {
		log.Errorf("unable to get bfla annotations: %s", err)
	}
	if err := p.addK8sDestination(ctx, cmpTrace, event.Telemetry, event.APIEvent.ID); err != nil {
		log.Errorf("unable to add k8s destination: %s", err)
	}
	if err := p.addK8sSource(ctx, cmpTrace, event.Telemetry, event.APIEvent.ID); err != nil {
		log.Errorf("unable to add k8s source: %s", err)
	}
	if err := p.addDetectedUser(ctx, cmpTrace, event.Telemetry, event.APIEvent.ID); err != nil {
		log.Errorf("unable to add detected user: %s", err)
	}

	// using a go routine to send traces in order not to block other modules.
	p.bflaDetector.SendTrace(cmpTrace)
	return err
}

func (p *bfla) addDetectedUser(ctx context.Context, cmpTrace *bfladetector.CompositeTrace, trace *pluginsmodels.Telemetry, eventID uint) (err error) {
	if cmpTrace.DetectedUser != nil {
		return nil
	}
	if trace == nil {
		return nil
	}

	cmpTrace.DetectedUser, err = bfladetector.GetUserID(convertHeadersToMap(trace.Request.Common.Headers))
	if err != nil {
		log.Error(err)
	}
	if cmpTrace.DetectedUser == nil {
		return nil
	}
	cmpTrace.DetectedUser.IPAddress = trace.SourceAddress
	annDest := core.Annotation{Name: bfladetector.DetectedIDAnnotationName}
	if annDest.Annotation, err = json.Marshal(cmpTrace.DetectedUser); err != nil {
		return fmt.Errorf("unable to marshal user: %w", err)
	}
	if err := p.accessor.CreateAPIEventAnnotations(ctx, bfladetector.ModuleName, eventID, annDest); err != nil {
		return fmt.Errorf("failed to create event annotation: %w", err)
	}
	return nil
}

func convertHeadersToMap(headers []*pluginsmodels.Header) http.Header {
	httpheaders := http.Header{}
	for _, h := range headers {
		httpheaders.Add(h.Key, h.Value)
	}
	return httpheaders
}

func (p *bfla) addK8sSource(ctx context.Context, cmpTrace *bfladetector.CompositeTrace, trace *pluginsmodels.Telemetry, eventID uint) error {
	if cmpTrace.K8SSource != nil {
		return nil
	}
	if trace == nil {
		return nil
	}
	srcObj, err := k8straceannotator.DetectSourceObject(ctx, p.k8s, trace)
	if err != nil {
		return fmt.Errorf("unable to detect k8s src: %w", err)
	}
	cmpTrace.K8SSource = k8straceannotator.NewRef(srcObj)
	annSrc := core.Annotation{Name: bfladetector.K8sSrcAnnotationName}
	if annSrc.Annotation, err = json.Marshal(cmpTrace.K8SSource); err != nil {
		return fmt.Errorf("unable to marshal src: %w", err)
	}
	if err := p.accessor.CreateAPIEventAnnotations(ctx, bfladetector.ModuleName, eventID, annSrc); err != nil {
		return fmt.Errorf("failure creating src event annotation: %w", err)
	}
	return nil
}

func (p *bfla) addK8sDestination(ctx context.Context, cmpTrace *bfladetector.CompositeTrace, trace *pluginsmodels.Telemetry, eventID uint) error {
	if cmpTrace.K8SDestination != nil {
		return nil
	}
	if trace == nil {
		return nil
	}
	destObj, err := k8straceannotator.DetectDestinationObject(ctx, p.k8s, trace)
	if err != nil {
		return fmt.Errorf("unable to detect k8s: %w", err)
	}
	cmpTrace.K8SDestination = k8straceannotator.NewRef(destObj)
	annDest := core.Annotation{Name: bfladetector.K8sDstAnnotationName}
	if annDest.Annotation, err = json.Marshal(cmpTrace.K8SDestination); err != nil {
		return fmt.Errorf("unable to marshal k8s dest: %w", err)
	}
	if err := p.accessor.CreateAPIEventAnnotations(ctx, bfladetector.ModuleName, eventID, annDest); err != nil {
		return fmt.Errorf("failure creating dest event annotation: %w", err)
	}
	return nil
}

func getBFLAAnnotations(ctx context.Context, accessor core.BackendAccessor, eventID uint) (dest, src *k8straceannotator.K8sObjectRef, user *bfladetector.DetectedUser, err error) {
	anns, err := accessor.ListAPIEventAnnotations(ctx, bfladetector.ModuleName, eventID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to get annotations for event=%d; %v", eventID, err)
	}
	for _, ann := range anns {
		switch ann.Name {
		case bfladetector.K8sDstAnnotationName:
			dest = &k8straceannotator.K8sObjectRef{}
			if err := json.Unmarshal(ann.Annotation, dest); err != nil {
				log.Errorf("unable to unmarshal k8s dest annotation for event=%d; %v", eventID, err)
			}
		case bfladetector.K8sSrcAnnotationName:
			src = &k8straceannotator.K8sObjectRef{}
			if err := json.Unmarshal(ann.Annotation, src); err != nil {
				log.Errorf("unable to unmarshal k8s dest annotation for event=%d; %v", eventID, err)
			}
		case bfladetector.DetectedIDAnnotationName:
			user = &bfladetector.DetectedUser{}
			if err := json.Unmarshal(ann.Annotation, user); err != nil {
				log.Errorf("unable to unmarshal k8s dest annotation for event=%d; %v", eventID, err)
			}
		}
	}

	return
}

func (p *bfla) EventAnnotationNotify(modName string, eventID uint, ann core.Annotation) error {
	log.Debugf("[BFLA] EventAnnotationNotify %s %d %s", modName, eventID, ann.Name)
	return nil
}

func (p *bfla) APIAnnotationNotify(modName string, apiID uint, ann *core.Annotation) error {
	log.Debugf("[BFLA] APIAnnotationNotify %s %d %s", modName, apiID, ann.Name)
	return nil
}

type httpHandler struct {
	state           recovery.StatePersister
	bflaDetector    bfladetector.BFLADetector
	openAPIProvider bfladetector.OpenAPIProvider
	accessor        core.BackendAccessor
}

func (h httpHandler) GetEvent(w http.ResponseWriter, r *http.Request, eventID int) {
	uEventID := uint32(eventID)
	events, err := h.accessor.GetAPIEvents(r.Context(), database.GetAPIEventsQuery{EventID: &uEventID})
	if err != nil {
		httpResponse(w, http.StatusBadRequest, &oapicommon.ApiResponse{Message: err.Error()})
		return
	}
	if len(events) == 0 {
		httpResponse(w, http.StatusNotFound, &oapicommon.ApiResponse{Message: fmt.Sprintf("not found event with id: %d", eventID)})
		return
	}
	event := events[0]

	dest, src, user, err := getBFLAAnnotations(r.Context(), h.accessor, uint(eventID))
	if err != nil {
		httpResponse(w, http.StatusBadRequest, &oapicommon.ApiResponse{Message: err.Error()})
		return
	}
	e := restapi.APIEventAnnotations{
		BflaStatus:           restapi.LEGITIMATE,
		DestinationK8sObject: (*restapi.K8sObjectRef)(dest),
		SourceK8sObject:      (*restapi.K8sObjectRef)(src),
		External:             src == nil,
	}
	apiinfo, err := h.accessor.GetAPIInfo(r.Context(), event.APIInfoID)
	if err != nil {
		httpResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: err.Error()})
		return
	}
	spc, err := h.openAPIProvider.GetOpenAPI(apiinfo, event.APIInfoID)
	if err != nil {
		log.Error("unable to get the spec")
	}
	if user != nil {
		e.DetectedUser = &restapi.DetectedUser{
			Id:        user.ID,
			Source:    restapi.DetectedUserSource(user.Source.String()),
			IpAddress: user.IPAddress,
		}
		e.MismatchedScopes = user.IsMismatchedScopes(bfladetector.GetSpecOperation(spc, event.Method, event.Path))
	}
	specType := bfladetector.SpecTypeFromAPIInfo(apiinfo)
	if specType == bfladetector.SpecTypeNone {
		e.BflaStatus = restapi.NOSPEC
		httpResponse(w, http.StatusOK, e)
		return
	}
	if h.bflaDetector.IsLearning(event.APIInfoID) {
		e.BflaStatus = restapi.LEARNING
		httpResponse(w, http.StatusOK, e)
		return
	}

	resolvedPath := bfladetector.ResolvePath(apiinfo, event)
	if obj, err := h.bflaDetector.FindSourceObj(resolvedPath, string(event.Method), src.Uid, event.APIInfoID); err != nil {
		log.Error(err)
	} else if !obj.Authorized {
		e.BflaStatus = bfladetector.ResolveBFLAStatusInt(int(event.StatusCode))
	}
	httpResponse(w, http.StatusOK, e)
}

// nolint:stylecheck,revive
func (h httpHandler) PostAuthorizationModelApiID(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID) {
	defer r.Body.Close()

	ctx := r.Context()
	select {
	case <-ctx.Done():
		httpResponse(w, http.StatusCreated, &oapicommon.ApiResponse{Message: fmt.Sprintf("the request took too long: %s", ctx.Err())})
	default:
		specType := bfladetector.SpecTypeNone
		if apiinfo, err := h.accessor.GetAPIInfo(r.Context(), uint(apiID)); err != nil {
			log.Errorf("error getting api info; id=%d", apiID)
		} else {
			specType = bfladetector.SpecTypeFromAPIInfo(apiinfo)
		}
		if specType == bfladetector.SpecTypeNone {
			httpResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: "Spec not found, please either provide or reconstruct an api spec"})
			return
		}
		authModelReq := &restapi.AuthorizationModel{}
		if err := json.NewDecoder(r.Body).Decode(authModelReq); err != nil {
			httpResponse(w, http.StatusBadRequest, &oapicommon.ApiResponse{Message: fmt.Sprintf("error decoding body; id=%d err: %s", apiID, err)})
			return
		}

		h.bflaDetector.ProvideAuthzModel(uint(apiID), FromRestapiAuthorizationModel(authModelReq))

		httpResponse(w, http.StatusCreated, &oapicommon.ApiResponse{Message: "Success"})
	}
}

// nolint:stylecheck,revive
func (h httpHandler) GetAuthorizationModelApiID(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID) {
	apiinfo, err := h.accessor.GetAPIInfo(r.Context(), uint(apiID))
	if err != nil {
		log.Errorf("error getting api info; id=%d", apiID)
	}
	specType := bfladetector.SpecTypeFromAPIInfo(apiinfo)
	if specType == bfladetector.SpecTypeNone {
		httpResponse(w, http.StatusOK, &restapi.AuthorizationModel{SpecType: bfladetector.ToRestapiSpecType(specType)})
		return
	}
	if h.bflaDetector.IsLearning(uint(apiID)) {
		httpResponse(w, http.StatusOK, &restapi.AuthorizationModel{Learning: true})
		return
	}
	authModel := &bfladetector.AuthorizationModel{}
	_, found, err := h.state.UseState(uint(apiID), bfladetector.AuthzModelAnnotationName, authModel)
	if err != nil {
		httpResponse(w, http.StatusBadRequest, &oapicommon.ApiResponse{Message: err.Error()})
		return
	}
	if !found {
		httpResponse(w, http.StatusNotFound, &oapicommon.ApiResponse{Message: fmt.Sprintf("auth model with id=%d not found", apiID)})
		return
	}
	res := ToRestapiAuthorizationModel(authModel)
	res.SpecType = bfladetector.ToRestapiSpecType(specType)
	httpResponse(w, http.StatusOK, res)
}

func FromRestapiAuthorizationModel(am *restapi.AuthorizationModel) bfladetector.AuthorizationModel {
	res := bfladetector.AuthorizationModel{}
	for _, o := range am.Operations {
		resOp := &bfladetector.Operation{Method: o.Method, Path: o.Path}
		for _, aud := range o.Audience {
			resAud := &bfladetector.SourceObject{
				Authorized: aud.Authorized,
				External:   aud.External,
				K8sObject:  (*k8straceannotator.K8sObjectRef)(aud.K8sObject),
			}
			for _, user := range aud.EndUsers {
				resAud.EndUsers = append(resAud.EndUsers, &bfladetector.DetectedUser{
					ID:        user.Id,
					IPAddress: user.IpAddress,
					Source:    bfladetector.DetectedUserSourceFromString(string(user.Source)),
				})
			}
			resOp.Audience = append(resOp.Audience, resAud)
		}
		res.Operations = append(res.Operations, resOp)
	}
	return res
}

func ToRestapiAuthorizationModel(am *bfladetector.AuthorizationModel) *restapi.AuthorizationModel {
	res := &restapi.AuthorizationModel{}
	for _, o := range am.Operations {
		resOp := restapi.AuthorizationModelOperation{Method: o.Method, Path: o.Path}
		for _, aud := range o.Audience {
			resAud := restapi.AuthorizationModelAudience{
				Authorized:    aud.Authorized,
				External:      aud.External,
				K8sObject:     (*restapi.K8sObjectRef)(aud.K8sObject),
				StatusCode:    int(aud.StatusCode),
				LastTime:      &aud.LastTime,
				WarningStatus: aud.WarningStatus,
			}
			for _, user := range aud.EndUsers {
				resAud.EndUsers = append(resAud.EndUsers, restapi.DetectedUser{
					Id:        user.ID,
					Source:    restapi.DetectedUserSource(user.Source.String()),
					IpAddress: user.IPAddress,
				})
			}
			resOp.Audience = append(resOp.Audience, resAud)
		}
		res.Operations = append(res.Operations, resOp)
	}
	return res
}

// nolint:stylecheck,revive
func (h httpHandler) PutAuthorizationModelApiIDApprove(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID, params restapi.PutAuthorizationModelApiIDApproveParams) {
	done := make(chan struct{})
	ctx := r.Context()
	clientRef := &k8straceannotator.K8sObjectRef{Uid: params.K8sClientUid} // TODO this looks wrong.
	go func() {
		log.Infof("approve operation on api=%d path=%s method=%s ", apiID, params.Path, params.Method)
		h.bflaDetector.ApproveTrace(params.Path, strings.ToUpper(params.Method), clientRef, uint(apiID), nil)
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		err := ctx.Err()
		log.Error(err)
		httpResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: err.Error()})
	case <-done:
		log.Infof("approve applied successfully on api=%d path=%s method=%s ", apiID, params.Path, params.Method)
		httpResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: "Requested approve operation on api event"})
	}
}

// nolint:stylecheck,revive
func (h httpHandler) PutAuthorizationModelApiIDDeny(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID, params restapi.PutAuthorizationModelApiIDDenyParams) {
	done := make(chan struct{})
	ctx := r.Context()
	clientRef := &k8straceannotator.K8sObjectRef{Uid: params.K8sClientUid} // TODO this looks wrong.
	go func() {
		log.Infof("deny operation on api=%d path=%s method=%s ", apiID, params.Path, params.Method)
		h.bflaDetector.DenyTrace(params.Path, strings.ToUpper(params.Method), clientRef, uint(apiID), nil)
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		err := ctx.Err()
		log.Error(err)
		httpResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: err.Error()})
	case <-done:
		log.Infof("deny applied successfully on api=%d path=%s method=%s ", apiID, params.Path, params.Method)
		httpResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: "Reqested deny operation on api event"})
	}
}

// nolint:stylecheck,revive
func (h httpHandler) PutAuthorizationModelApiIDLearningReset(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID, params restapi.PutAuthorizationModelApiIDLearningResetParams) {
	done := make(chan struct{})
	ctx := r.Context()
	go func() {
		log.Infof("reset learning api=%d", apiID)
		if params.NrTraces == nil {
			h.bflaDetector.ResetLearning(uint(apiID), -1)
		} else {
			h.bflaDetector.ResetLearning(uint(apiID), *params.NrTraces)
		}
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		err := ctx.Err()
		log.Error(err)
		httpResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: err.Error()})
	case <-done:
		log.Infof("reset learning applied successfully on api=%d", apiID)
		httpResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: "Reqested reset learning operation on api event"})
	}
}

// nolint:stylecheck,revive
func (h httpHandler) PutAuthorizationModelApiIDLearningStart(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID, params restapi.PutAuthorizationModelApiIDLearningStartParams) {
	done := make(chan struct{})
	ctx := r.Context()
	go func() {
		log.Infof("start learning api=%d", apiID)
		if params.NrTraces == nil {
			h.bflaDetector.StartLearning(uint(apiID), -1)
		} else {
			h.bflaDetector.StartLearning(uint(apiID), *params.NrTraces)
		}
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		err := ctx.Err()
		log.Error(err)
		httpResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: err.Error()})
	case <-done:
		log.Infof("start learning applied successfully on api=%d", apiID)
		httpResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: "Reqested start learning operation on api event"})
	}
}

// nolint:stylecheck,revive
func (h httpHandler) PutAuthorizationModelApiIDLearningStop(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID) {
	done := make(chan struct{})
	ctx := r.Context()
	go func() {
		log.Infof("stop learning api=%d", apiID)
		h.bflaDetector.StopLearning(uint(apiID))
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		err := ctx.Err()
		log.Error(err)
		httpResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: err.Error()})
	case <-done:
		log.Infof("stop learning applied successfully on api=%d", apiID)
		httpResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: "Reqested stop learning operation on api event"})
	}
}

// nolint:stylecheck,revive
func (h httpHandler) PutEventIdOperation(w http.ResponseWriter, r *http.Request, eventID int, operation restapi.OperationEnum) {
	uEventID := uint32(eventID)
	events, err := h.accessor.GetAPIEvents(r.Context(), database.GetAPIEventsQuery{EventID: &uEventID})
	if err != nil {
		httpResponse(w, http.StatusBadRequest, &oapicommon.ApiResponse{Message: err.Error()})
		return
	}
	if len(events) == 0 {
		httpResponse(w, http.StatusNotFound, &oapicommon.ApiResponse{Message: fmt.Sprintf("not found event with id: %d", eventID)})
		return
	}
	apiEvent := events[0]

	_, src, user, err := getBFLAAnnotations(r.Context(), h.accessor, uint(eventID))
	if err != nil {
		log.Error(err)
		httpResponse(w, http.StatusBadRequest, &oapicommon.ApiResponse{
			Message: err.Error(),
		})
		return
	}

	apiInfo, err := h.accessor.GetAPIInfo(r.Context(), apiEvent.APIInfoID)
	if err != nil {
		log.Error(err)
		httpResponse(w, http.StatusBadRequest, &oapicommon.ApiResponse{
			Message: err.Error(),
		})
		return
	}
	done := make(chan struct{})
	ctx := r.Context()
	go func() {
		log.Infof("apply %s operation on trace=%d", operation, eventID)
		resolvedPath := bfladetector.ResolvePath(apiInfo, apiEvent)
		switch operation {
		case restapi.Approve:
			h.bflaDetector.ApproveTrace(resolvedPath, string(apiEvent.Method), src, apiEvent.APIInfoID, nil)
		case restapi.Deny:
			h.bflaDetector.DenyTrace(resolvedPath, string(apiEvent.Method), src, apiEvent.APIInfoID, nil)
		case restapi.ApproveUser:
			h.bflaDetector.ApproveTrace(resolvedPath, string(apiEvent.Method), src, apiEvent.APIInfoID, user)
		case restapi.DenyUser:
			h.bflaDetector.DenyTrace(resolvedPath, string(apiEvent.Method), src, apiEvent.APIInfoID, user)
		}
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		err := ctx.Err()
		log.Error(err)
		httpResponse(w, http.StatusInternalServerError, &oapicommon.ApiResponse{Message: err.Error()})
	case <-done:
		log.Infof("%s operation applied successfully on trace=%d", operation, eventID)
		httpResponse(w, http.StatusOK, &oapicommon.ApiResponse{Message: fmt.Sprintf("Reqested %s operation on api event", operation)})
	}
}

func (h httpHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	httpResponse(w, http.StatusOK, &oapicommon.ModuleVersion{Version: moduleVersion})
}

func (h httpHandler) GetApiFindings(w http.ResponseWriter, r *http.Request, apiID oapicommon.ApiID, params restapi.GetApiFindingsParams) {
	return
}

func httpResponse(w http.ResponseWriter, code int, v interface{}) {
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), code)
		return
	}
}

//nolint:gochecknoinits
func init() {
	core.RegisterModule(recovery.ResyncedModule(newModule))
}
