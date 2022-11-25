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
	log.Infof("[%s] APIEvent.ID=%d Path=%s Method=%s", ModuleName, event.APIEvent.ID, event.Telemetry.Request.Path, event.Telemetry.Request.Method)
	labelMap, err := p.accessor.GetLabelsTable(ctx).GetLabels(ctx, event.APIEvent.ID)
	if err != nil {
		log.Errorf("[%s] error in labels lookup for event: %d", ModuleName, event.APIEvent.ID)
		return
	}
	if len(labelMap) > 0 {
		for k, v := range labelMap {
			log.Infof("[%s] found label: %s -> %s", ModuleName, k, v)
		}
	}
}
