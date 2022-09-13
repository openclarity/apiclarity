package analytics_core

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/analytics_core/restapi"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"

	log "github.com/sirupsen/logrus"
)

const (
	moduleVersion = "0.0.0"
)

type analyticsCore struct {
	httpHandler http.Handler

	accessor core.BackendAccessor
	info     *core.ModuleInfo
}

func (p *analyticsCore) Info() core.ModuleInfo {
	return *p.info
}

func (p *analyticsCore) Name() string              { return "analytics_core" }
func (p *analyticsCore) HTTPHandler() http.Handler { return p.httpHandler }

func newModule(ctx context.Context, accessor core.BackendAccessor) (_ core.Module, err error) {
	p := &analyticsCore{
		httpHandler: nil,
		accessor:    accessor,
		info:        &core.ModuleInfo{Name: "analytics_core", Description: "analytics_core"},
	}
	handler := &httpHandler{
		accessor: accessor,
	}

	p.httpHandler = restapi.HandlerWithOptions(handler, restapi.ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + "analytics_core"})
	//sp := recovery.NewStatePersister(ctx, accessor, bfladetector.ModuleName, persistenceInterval)
	return p, nil
}

type httpHandler struct {
	accessor core.BackendAccessor
}

func (p *analyticsCore) EventNotify(ctx context.Context, event *core.Event) {
	if err := p.eventNotify(ctx, event); err != nil {
		log.Errorf("[Analytics-Core] EventNotify: %s", err)
	}
}

func (p *analyticsCore) eventNotify(ctx context.Context, event *core.Event) (err error) {
	log.Debugf("[Analytics-Core] received a new event for API(%v) Event(%v) ", event.APIEvent.APIInfoID, event.APIEvent.ID)
	log.Errorf("[Analytics-Core] received a new event for API(%v) Event(%v) ", event.APIEvent.APIInfoID, event.APIEvent.ID)

	return nil
}

func (h httpHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	httpResponse(w, http.StatusOK, &restapi.Version{Version: moduleVersion})
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
	core.RegisterModule(newModule)
}
