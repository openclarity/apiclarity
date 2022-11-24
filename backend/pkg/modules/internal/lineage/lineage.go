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
	"fmt"
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

func newModule(ctx context.Context, accessor core.BackendAccessor) (core.Module, error) {
	log.Debugf("[%s] Start():: -->", ModuleName)
	s := serverHandler{
		info: &core.ModuleInfo{
			Name:        ModuleName,
			Description: ModuleDescription,
		},
	}
	s.httpHandler = HandlerWithOptions(&controller{accessor: accessor, serverHandler: &s}, ChiServerOptions{BaseURL: fmt.Sprintf("/api/modules/%s", ModuleName)})
	log.Debugf("[%s] Start():: <--", ModuleName)
	return &s, nil
}

type controller struct {
	accessor      core.BackendAccessor
	serverHandler *serverHandler
}

func (c *controller) GetVersion(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(Version{Version: ModuleVersion})
}

type serverHandler struct {
	info        *core.ModuleInfo
	httpHandler http.Handler
	config      *Config
}

func (p *serverHandler) Info() core.ModuleInfo {
	return *p.info
}

func (p *serverHandler) EventNotify(ctx context.Context, event *core.Event) {
	log.Infof("[demo] APIEvent.ID=%d Path=%s Method=%s", event.APIEvent.ID, event.Telemetry.Request.Path, event.Telemetry.Request.Method)
}

func (p *serverHandler) HTTPHandler() http.Handler {
	return p.httpHandler
}
