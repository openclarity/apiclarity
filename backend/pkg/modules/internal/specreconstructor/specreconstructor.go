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

package specreconstructor

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	oapicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/logging"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/specreconstructor/restapi"
)

//nolint:gochecknoinits
func init() {
	core.RegisterModule(newModule)
}

const (
	ModuleName        = "specreconstructor"
	ModuleDescription = "Reconstruct an openapi specification from traces"
	ModuleVersion     = "0.0.0"
	EmptyJSON         = "{}"
)

type specReconstructorPlugin struct {
	httpHandler http.Handler
	info        *core.ModuleInfo
	accessor    core.BackendAccessor
}

func newModule(ctx context.Context, accessor core.BackendAccessor) (core.Module, error) {
	plugin := specReconstructorPlugin{
		httpHandler: nil,
		accessor:    accessor,
		info: &core.ModuleInfo{
			Name:        ModuleName,
			Description: ModuleDescription,
		},
	}
	plugin.httpHandler = restapi.HandlerWithOptions(&specReconstructorPluginHTTPHandler{plugin: &plugin}, restapi.ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + ModuleName})
	return &plugin, nil
}

func (d *specReconstructorPlugin) Info() core.ModuleInfo {
	return *d.info
}

func (d *specReconstructorPlugin) EventNotify(ctx context.Context, event *core.Event) {
	// This music doesn't use this
}

func (d *specReconstructorPlugin) HTTPHandler() http.Handler {
	return d.httpHandler
}

type specReconstructorPluginHTTPHandler struct {
	plugin *specReconstructorPlugin
}

func httpError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
}

func httpResponse(writer http.ResponseWriter, statusCode int, data interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	if err := json.NewEncoder(writer).Encode(data); err != nil {
		httpError(writer, err)
	}
}

func (c *specReconstructorPluginHTTPHandler) PostApiIdStart(w http.ResponseWriter, r *http.Request, apiId int64) {
	log.Debugf("PostApiIdStart(%v): --> <--", apiId)

	apiID := uint32(apiId)
	component := "*"

	if err := c.plugin.accessor.GetTraceSamplingAccessor().AddHostToTrace(component, apiID); err != nil {
		log.Errorf("Failed to add API %v in APIs to trace: %v", apiID, err)
		httpResponse(w, http.StatusInternalServerError, EmptyJSON)
		return
	}
	log.Infof("Tracing successfully started for api=%d", apiID)

	// Success...
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (c *specReconstructorPluginHTTPHandler) PostApiIdStop(w http.ResponseWriter, r *http.Request, apiId int64) {
	log.Debugf("PostApiIdStop(%v): --> <--", apiId)

	apiID := uint32(apiId)
	component := "*"

	if err := c.plugin.accessor.GetTraceSamplingAccessor().RemoveHostToTrace(component, apiID); err != nil {
		log.Errorf("Failed to remove API %v from APIs to trace: %v", apiID, err)
		httpResponse(w, http.StatusInternalServerError, EmptyJSON)
		return
	}
	log.Infof("Tracing successfully stoped for api=%d", apiID)

	// Success...
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (c *specReconstructorPluginHTTPHandler) PostEnable(w http.ResponseWriter, r *http.Request) {
	log.Debugf("PostEnable(): --> <--")

	// Decode the restapi.TestInput requesBody
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Can't read body content, error=(%v)", err)
		httpResponse(w, http.StatusBadRequest, EmptyJSON)
		return
	}
	log.Debugf(string(body))
	var enable restapi.FeatureEnable
	err = json.Unmarshal(body, &enable)
	if err != nil {
		logging.Errorf("Failed to decode the request body, error=(%v)", err)
		httpResponse(w, http.StatusInternalServerError, EmptyJSON)
		return
	}

	component := "*"
	flag := enable.Enable
	if !*flag {
		if err := c.plugin.accessor.GetTraceSamplingAccessor().ResetForComponent(component); err != nil {
			log.Errorf("Failed to reset trace sampling for module %v: %v", component, err)
			httpResponse(w, http.StatusInternalServerError, EmptyJSON)
		}
	}

	// Success...
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (c *specReconstructorPluginHTTPHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(oapicommon.ModuleVersion{Version: ModuleVersion}); err != nil {
		httpError(w, err)
	}
}
