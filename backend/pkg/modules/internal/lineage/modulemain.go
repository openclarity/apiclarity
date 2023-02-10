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
	"fmt"
	"net/http"
	"time"

	apicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/lineage/restapi"
	apilabels "github.com/openclarity/apiclarity/plugins/api/labels"
	log "github.com/sirupsen/logrus"
)

const (
	ModuleName        = "lineage"
	ModuleDescription = "Attempts to infer data lineage of APIs and add extensions to the OpenAPI spec"
	ModuleVersion     = "0.0.1"

	DefaultBasePath = core.BaseHTTPPath + "/" + ModuleName
)

var (
	relationshipMapSize int64         = 8192
	relationshipMapTTL  time.Duration = time.Second * 60
)

func init() {
	core.RegisterModule(newModule)
}

type controller struct {
	accessor    core.BackendAccessor
	info        *core.ModuleInfo
	httpHandler http.Handler
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

	parentMap, err := NewParentMap(relationshipMapSize, relationshipMapTTL)
	if err != nil {
		return nil, fmt.Errorf("cannot allocate relationship map of %d bytes: %v", relationshipMapSize, err)
	}

	sh := restapi.NewStrictHandler(lineageHTTPHandler{accessor: accessor, parents: parentMap}, nil)
	s.httpHandler = restapi.HandlerWithOptions(sh, restapi.ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + ModuleName})
	log.Debugf("[%s] newModule():: <--", ModuleName)
	return s, nil
}

func (p *controller) Info() core.ModuleInfo {
	return *p.info
}

func (p *controller) Name() string {
	return ModuleName
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

type lineageHTTPHandler struct {
	accessor core.BackendAccessor
	parents  *ParentMap
}

func (h lineageHTTPHandler) GetVersion(ctx context.Context, request restapi.GetVersionRequestObject) (restapi.GetVersionResponseObject, error) {
	return restapi.GetVersion200JSONResponse(apicommon.ModuleVersion{Version: ModuleVersion}), nil
}

func convertLabelsToAPIOperations(labels []database.Label) []restapi.APIOperation {
	operations := make([]restapi.APIOperation, len(labels))
	for idx, lbl := range labels {
		operations[idx].Path = &labels[idx].Path
		operations[idx].Id = int64(lbl.APIInfoID)
		operations[idx].Operation = &labels[idx].Method
	}
	return operations
}

func (h lineageHTTPHandler) GetLineage(ctx context.Context, request restapi.GetLineageRequestObject) (restapi.GetLineageResponseObject, error) {
	lineageResponse := restapi.APILineage{
		Id: restapi.APIOperation{
			Id:        request.ApiID,
			Operation: request.Params.Operation,
			Path:      request.Params.Path,
		},
		Children: nil,
		Parents:  nil,
	}

	foundChildren, err := h.accessor.GetLabelsTable(ctx).GetLabelsLineageChildren(ctx, uint(request.ApiID), request.Params.Operation, request.Params.Path)
	if err != nil {
		log.Errorf("[%s] error in children labels lookup for apiID %d: %v", ModuleName, request.ApiID, err)
	} else if len(foundChildren) == 0 {
		log.Debugf("[%s] no children labels found, skipping apiID %d", ModuleName, request.ApiID)
	} else {
		children := convertLabelsToAPIOperations(foundChildren)
		lineageResponse.Children = &children
	}

	//Find parent
	foundParents, err := h.accessor.GetLabelsTable(ctx).GetLabelsLineageParents(ctx, uint(request.ApiID), request.Params.Operation, request.Params.Path)
	if err != nil {
		log.Errorf("[%s] error in parent labels lookup for apiID %d: %v", ModuleName, request.ApiID, err)
	} else if len(foundParents) == 0 {
		log.Debugf("[%s] no parent labels found, skipping apiID %d", ModuleName, request.ApiID)
	} else {
		parents := make([]restapi.APIOperation, len(foundParents))
		for idx, lbl := range foundParents {
			if lbl.Key != apilabels.DataLineageIDKey {
				log.Errorf("[%s] ignoring parent label with unexpected key for apiID %d: %s", ModuleName, request.ApiID, lbl.Key)
				continue
			}
			parents[idx].Path = &foundParents[idx].Path
			parents[idx].Operation = &foundParents[idx].Method
			parents[idx].Id = int64(lbl.APIInfoID)
			// Is this the right ID?
			newPID, found := h.parents.GetParent(lbl.Value)
			if found {
				updateRelation := false
				for found {
					if ppID, found := h.parents.GetParent(newPID); found {
						newPID = ppID
						updateRelation = true
					}
				}
				// Update searches for the future
				if updateRelation {
					h.parents.PutParent(lbl.Value, newPID)
				}
				// We could probably do this only once if we tracked it in this call since it's likely to be the same parent.
				h.accessor.GetLabelsTable(ctx).ReplaceLabelMatching(ctx, apilabels.DataLineageParentKey, lbl.Value, newPID)
			}
		}
		lineageResponse.Parents = &parents
	}

	//Encode
	return restapi.GetLineage200JSONResponse(lineageResponse), nil
}

func (h lineageHTTPHandler) PostLineageExtended(ctx context.Context, request restapi.PostLineageExtendedRequestObject) (restapi.PostLineageExtendedResponseObject, error) {
	if request.Body == nil {
		return restapi.PostLineageExtended200Response{}, nil
	}

	for _, rlt := range *request.Body {
		if rlt.Parent != nil {
			log.Infof("[%s] post lineage parent labels found, id: %s, parent: %s", ModuleName, rlt.Id, *rlt.Parent)
			h.parents.PutParent(rlt.Id, *rlt.Parent)
		}
		if rlt.Child != nil {
			h.parents.PutParent(*rlt.Child, rlt.Id)
		}
	}

	return restapi.PostLineageExtended200Response{}, nil
}
