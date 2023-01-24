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

package lineage

import (
	"context"
	"net/http"

	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	log "github.com/sirupsen/logrus"
)

const (
	ModuleName        = "lineage"
	ModuleDescription = "Attempts to infer data lineage of APIs and add extensions to the OpenAPI spec"
	ModuleVersion     = "0.0.1"
)

func init() {
	core.RegisterModule(newModule)
}

type controller struct {
	accessor    core.BackendAccessor
	info        *core.ModuleInfo
	httpHandler http.Handler
	config      *Config
}

func newModule(ctx context.Context, accessor core.BackendAccessor) (core.Module, error) {
	log.Debugf("[%s] newModule():: -->", ModuleName)
	s := &controller{
		accessor: accessor,
		info: &core.ModuleInfo{
			Name:        ModuleName,
			Description: ModuleDescription,
		},
	}
	h := &httpHandler{
		accessor: accessor,
	}
	s.httpHandler = HandlerWithOptions(h, ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + ModuleName})
	log.Debugf("[%s] newModule():: <--", ModuleName)
	return s, nil
}

func (p *controller) Info() core.ModuleInfo {
	return *p.info
}

func (p *controller) HTTPHandler() http.Handler {
	return p.httpHandler
}

func (p *controller) EventNotify(ctx context.Context, event *core.Event) {
	apiEvent := event.APIEvent

	log.Infof("[%s] APIEvent.ID=%d APIInfoID=%d Path=%s Method=%s", ModuleName, apiEvent.ID, apiEvent.APIInfoID, event.Telemetry.Request.Path, event.Telemetry.Request.Method)
	labelMap, err := p.accessor.GetLabelsTable(ctx).GetLabelsByEventID(ctx, apiEvent.ID)
	if err != nil {
		log.Errorf("[%s] error in labels lookup for event %d: %v", ModuleName, apiEvent.ID, err)
		return
	} else if len(labelMap) == 0 {
		log.Debugf("[%s] no labels found, skipping event: %d", ModuleName, apiEvent.ID)
		return
	}
	for label, value := range labelMap {
		log.Infof("[%s] label found, event: %d, apiID: %d, label: %s, value: %s", ModuleName, apiEvent.ID, apiEvent.APIInfoID, label, value)
	}
}

type httpHandler struct {
	accessor core.BackendAccessor
}

func (h httpHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	common.HTTPResponse(w, http.StatusOK, Version{Version: ModuleVersion})
}

func convertLabelsToAPIOperations(labels []database.Label) []APIOperation {
	ids := make([]int64, len(labels))
	operations := make([]APIOperation, len(labels))
	for idx, _ := range labels {
		ids[idx] = int64(labels[idx].APIInfoID)
		operations[idx].Path = &labels[idx].Path
		operations[idx].Id = &ids[idx]
		operations[idx].Operation = &labels[idx].Method
	}
	return operations
}

func (h httpHandler) GetLineage(w http.ResponseWriter, r *http.Request, apiID int64, params GetLineageParams) {
	lineageResponse := APILineage{
		Id: &APIOperation{
			Id:        &apiID,
			Operation: params.Operation,
			Path:      params.Path,
		},
		Children: nil,
		Parents:  nil,
	}

	foundChildren, err := h.accessor.GetLabelsTable(r.Context()).GetLabelsLineageChildren(r.Context(), uint(apiID), params.Operation, params.Path)
	if err != nil {
		log.Errorf("[%s] error in children labels lookup for apiID %d: %v", ModuleName, apiID, err)
	} else if len(foundChildren) == 0 {
		log.Debugf("[%s] no children labels found, skipping apiID %d", ModuleName, apiID)
	} else {
		children := convertLabelsToAPIOperations(foundChildren)
		lineageResponse.Children = &children
	}

	//Find parent
	foundParents, err := h.accessor.GetLabelsTable(r.Context()).GetLabelsLineageParents(r.Context(), uint(apiID), params.Operation, params.Path)
	if err != nil {
		log.Errorf("[%s] error in parent labels lookup for apiID %d: %v", ModuleName, apiID, err)
	} else if len(foundParents) == 0 {
		log.Debugf("[%s] no parent labels found, skipping apiID %d", ModuleName, apiID)
	} else {
		parents := convertLabelsToAPIOperations(foundParents)
		lineageResponse.Parents = &parents
	}

	//Encode
	common.HTTPResponse(w, http.StatusOK, lineageResponse)
}
