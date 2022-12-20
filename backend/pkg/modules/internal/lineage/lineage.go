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
	"encoding/json"
	"net/http"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	apilabels "github.com/openclarity/apiclarity/plugins/api/labels"
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
	s.httpHandler = HandlerWithOptions(s, ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + ModuleName})
	log.Debugf("[%s] newModule():: <--", ModuleName)
	return s, nil
}

func (c *controller) GetVersion(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(Version{Version: ModuleVersion})
}

func (p *controller) Info() core.ModuleInfo {
	return *p.info
}

func (p *controller) HTTPHandler() http.Handler {
	return p.httpHandler
}

func (p *controller) EventNotify(ctx context.Context, event *core.Event) {
	apiEvent := event.APIEvent

	log.Infof("[%s] APIEvent.ID=%d Path=%s Method=%s", ModuleName, apiEvent.ID, event.Telemetry.Request.Path, event.Telemetry.Request.Method)
	labelMap, err := p.accessor.GetLabelsTable(ctx).GetLabels(ctx, apiEvent.ID)
	if err != nil {
		log.Errorf("[%s] error in labels lookup for event: %d", ModuleName, apiEvent.ID)
		return
	} else if len(labelMap) == 0 {
		log.Debugf("[%s] no labels found, skipping event: %d", ModuleName, apiEvent.ID)
		return
	}

	//Convert labels that we care about to API annotations
	transferLabels := []string{
		apilabels.DataLineageUpstreamKey,
	}
	for _, label := range transferLabels {
		if value, found := labelMap[label]; found {
			log.Debugf("[%s] found label: %s -> %s", ModuleName, label, value)
			err = p.accessor.StoreAPIInfoAnnotations(ctx, ModuleName, apiEvent.APIInfoID, core.Annotation{Name: label, Annotation: []byte(value)})
			if err != nil {
				log.Errorf("[%s] error in APIInfoAnnotation lookup for event: %d, APIInfoID: %d", ModuleName, apiEvent.ID, apiEvent.APIInfoID)
			}
		}
	}
}
