package demo

import (
	"context"
	"encoding/json"
	"github.com/apiclarity/apiclarity/backend/pkg/database"
	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/core"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func init() {
	core.RegisterModule(newModule)
}

const ModuleName = "demo"

func newModule(ctx context.Context, accessor core.BackendAccessor) (core.Module, error) {
	//accessor.CreateAPIEventAnnotations(ctx, "mod", 1, core.AlertInfoAnn)
	return &demo{
		handler: HandlerWithOptions(&controller{accessor: accessor}, ChiServerOptions{BaseURL: "/api/modules/demo"}),
	}, nil
}

type controller struct {
	accessor core.BackendAccessor
}

func httpError(w http.ResponseWriter, err error) {
	w.WriteHeader(400)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
}

func httpResponse(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(400)
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
	err := c.accessor.CreateAPIEventAnnotations(r.Context(), ModuleName, uint(eventID), ann)
	if err != nil {
		httpError(w, err)
		return
	}
	httpResponse(w, "success")
}

func (c *controller) GetApiApiID(w http.ResponseWriter, r *http.Request, apiID int) {
	res, err := c.accessor.GetAPIInfo(r.Context(), uint(apiID))
	if err != nil {
		httpError(w, err)
		return
	}
	httpResponse(w, res)
}

func (c *controller) GetApiApiIDAnnotationAnnotation(w http.ResponseWriter, r *http.Request, apiID int, annotation string) {
	api, err := c.accessor.GetAPIInfoAnnotation(r.Context(), ModuleName, uint(apiID), annotation)
	if err != nil {
		httpError(w, err)
		return
	}
	httpResponse(w, api)
}

func (c *controller) DeleteApiApiIDAnnotationAnnotation(w http.ResponseWriter, r *http.Request, apiID int, annotation string) {
	err := c.accessor.DeleteAPIInfoAnnotations(r.Context(), ModuleName, uint(apiID), annotation)
	if err != nil {
		httpError(w, err)
		return
	}
	httpResponse(w, "success")
}

func (c *controller) PostApiApiIDAnnotationAnnotation(w http.ResponseWriter, r *http.Request, apiID int, annotation string) {
	type Data struct {
		Data string `json:"data"`
	}
	d := &Data{}
	json.NewDecoder(r.Body).Decode(d)
	err := c.accessor.StoreAPIInfoAnnotations(r.Context(), ModuleName, uint(apiID), core.Annotation{
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
	api, err := c.accessor.GetAPIEventAnnotation(r.Context(), ModuleName, uint(eventID), annotation)
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
	json.NewDecoder(r.Body).Decode(d)
	err := c.accessor.CreateAPIEventAnnotations(r.Context(), ModuleName, uint(eventID), core.Annotation{
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
	json.NewEncoder(w).Encode(Version{Version: "0.0.0"})
}

type demo struct {
	handler http.Handler
}

func (d *demo) Name() string {
	return ModuleName
}

func (d *demo) EventNotify(ctx context.Context, event *core.Event) {
	log.Infof("[demo] APIEvent.ID=%d Path=%s Method=%s", event.APIEvent.ID, event.Telemetry.Request.Path, event.Telemetry.Request.Method)
}

func (d *demo) HTTPHandler() http.Handler {
	return d.handler
}
