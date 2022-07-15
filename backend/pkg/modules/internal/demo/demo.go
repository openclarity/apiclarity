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

package demo

import (
	"context"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
)

//nolint:gochecknoinits
func init() {
	core.RegisterModule(newModule)
}

const (
	ModuleName        = "demo"
	ModuleDescription = "This is a demo module doing nothing"
)

func newModule(ctx context.Context, accessor core.BackendAccessor) (core.Module, error) {
	d := demo{
		info: &core.ModuleInfo{
			Name:        ModuleName,
			Description: ModuleDescription,
		},
	}
	d.handler = HandlerWithOptions(&controller{accessor: accessor, demo: &d}, ChiServerOptions{BaseURL: "/api/modules/demo"})
	return &d, nil
}

type controller struct {
	accessor core.BackendAccessor
	demo     *demo
}

func httpError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
}

func httpResponse(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		httpError(w, err)
	}
}

func (c *controller) PostAlertEventID(w http.ResponseWriter, r *http.Request, eventID int, params PostAlertEventIDParams) {
	var ann core.Annotation
	switch params.Type {
	case "INFO":
		ann = core.AlertInfoAnn
	case "WARN":
		ann = core.AlertWarnAnn
	}
	err := c.accessor.CreateAPIEventAnnotations(r.Context(), c.demo.info.Name, uint(eventID), ann)
	if err != nil {
		httpError(w, err)
		return
	}
	httpResponse(w, "success")
}

//nolint:stylecheck,revive
func (c *controller) GetApiApiID(w http.ResponseWriter, r *http.Request, apiID int) {
	res, err := c.accessor.GetAPIInfo(r.Context(), uint(apiID))
	if err != nil {
		httpError(w, err)
		return
	}
	httpResponse(w, res)
}

//nolint:stylecheck,revive
func (c *controller) GetApiApiIDAnnotationAnnotation(w http.ResponseWriter, r *http.Request, apiID int, annotation string) {
	api, err := c.accessor.GetAPIInfoAnnotation(r.Context(), c.demo.info.Name, uint(apiID), annotation)
	if err != nil {
		httpError(w, err)
		return
	}
	httpResponse(w, api)
}

//nolint:stylecheck,revive
func (c *controller) DeleteApiApiIDAnnotationAnnotation(w http.ResponseWriter, r *http.Request, apiID int, annotation string) {
	err := c.accessor.DeleteAPIInfoAnnotations(r.Context(), c.demo.info.Name, uint(apiID), annotation)
	if err != nil {
		httpError(w, err)
		return
	}
	httpResponse(w, "success")
}

//nolint:stylecheck,revive
func (c *controller) PostApiApiIDAnnotationAnnotation(w http.ResponseWriter, r *http.Request, apiID int, annotation string) {
	type Data struct {
		Data string `json:"data"`
	}
	d := &Data{}
	_ = json.NewDecoder(r.Body).Decode(d)
	err := c.accessor.StoreAPIInfoAnnotations(r.Context(), c.demo.info.Name, uint(apiID), core.Annotation{
		Name:       annotation,
		Annotation: []byte(d.Data),
	})
	if err != nil {
		httpError(w, err)
		return
	}
	httpResponse(w, "success")
}

func (c *controller) GetEventEventIDAnnotationAnnotation(w http.ResponseWriter, r *http.Request, eventID int, annotation string) {
	api, err := c.accessor.GetAPIEventAnnotation(r.Context(), c.demo.info.Name, uint(eventID), annotation)
	if err != nil {
		httpError(w, err)
		return
	}
	httpResponse(w, api)
}

func (c *controller) PostEventEventIDAnnotationAnnotation(w http.ResponseWriter, r *http.Request, eventID int, annotation string) {
	type Data struct {
		Data string `json:"data"`
	}
	d := &Data{}
	_ = json.NewDecoder(r.Body).Decode(d)
	err := c.accessor.CreateAPIEventAnnotations(r.Context(), c.demo.info.Name, uint(eventID), core.Annotation{
		Name:       annotation,
		Annotation: []byte(d.Data),
	})
	if err != nil {
		httpError(w, err)
		return
	}
	httpResponse(w, "success")
}

func (c *controller) PostEvents(w http.ResponseWriter, r *http.Request) {
	filter := database.GetAPIEventsQuery{}
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		httpError(w, err)
		return
	}
	evs, err := c.accessor.GetAPIEvents(r.Context(), filter)
	if err != nil {
		httpError(w, err)
		return
	}
	httpResponse(w, evs)
}

func (c *controller) GetVersion(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(Version{Version: "0.0.0"})
}

type demo struct {
	handler http.Handler
	info    *core.ModuleInfo
}

func (d *demo) Info() core.ModuleInfo {
	return *d.info
}

func (d *demo) EventNotify(ctx context.Context, event *core.Event) {
	log.Infof("[demo] APIEvent.ID=%d Path=%s Method=%s", event.APIEvent.ID, event.Telemetry.Request.Path, event.Telemetry.Request.Method)
}

func (d *demo) HTTPHandler() http.Handler {
	return d.handler
}
