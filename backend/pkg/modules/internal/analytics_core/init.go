package analytics_core

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/analytics_core/restapi"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/pubsub"

	log "github.com/sirupsen/logrus"
)

const (
	moduleVersion = "0.0.0"
)

type analyticsCore struct {
	httpHandler         http.Handler
	msgBroker           *pubsub.Handler
	accessor            core.BackendAccessor
	info                *core.ModuleInfo
	numWorkers          int
	proccFuncRegistered map[string][]AnalyticsModuleProccFunc
}

func (p *analyticsCore) Info() core.ModuleInfo {
	return *p.info
}

func (p *analyticsCore) Name() string              { return "analytics_core" }
func (p *analyticsCore) HTTPHandler() http.Handler { return p.httpHandler }

func (p *analyticsCore) handlerFunction(topic string, paritionId int, msgChannel chan interface{}) {
	for {
		message := <-msgChannel
		topicProccFunctions, okTopic := p.proccFuncRegistered[topic]
		if okTopic {
			for _, proccFunction := range topicProccFunctions {

				proccFunction.HandlerFunc(nil, message)
			}
		}
	}

}

func newModule(ctx context.Context, accessor core.BackendAccessor) (_ core.Module, err error) {
	p := &analyticsCore{
		httpHandler: nil,
		msgBroker:   nil,
		accessor:    accessor,
		info:        &core.ModuleInfo{Name: "analytics_core", Description: "analytics_core"},
		numWorkers:  1,
	}
	handler := &httpHandler{
		accessor: accessor,
	}
	p.httpHandler = restapi.HandlerWithOptions(handler, restapi.ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + "analytics_core"})
	p.msgBroker = pubsub.NewHandler()

	for i := 0; i < p.numWorkers; i++ {
		trace_channel := p.msgBroker.AddSubscriptionShard("trace", i)
		go p.handlerFunction("trace", i, trace_channel)
		api_channel := p.msgBroker.AddSubscriptionShard("api", i)
		go p.handlerFunction("trace", i, api_channel)
		api_endpoint_channel := p.msgBroker.AddSubscriptionShard("api_endpoint", i)
		go p.handlerFunction("trace", i, api_endpoint_channel)
		object_channel := p.msgBroker.AddSubscriptionShard("object", i)
		go p.handlerFunction("trace", i, object_channel)
		entity_channel := p.msgBroker.AddSubscriptionShard("entity", i)
		go p.handlerFunction("trace", i, entity_channel)
	}

	return p, nil
}

type ProcFuncDataFrames struct {
	dataFrames map[int]*interface{}
}

type AnalyticsModuleProccFunc interface {
	HandlerFunc(dataFrames *ProcFuncDataFrames, message interface{})
}

func (p *analyticsCore) RegisterAnalyticsModuleHandler(topic string, proccFunc AnalyticsModuleProccFunc) {
	topicProccFunctions, okTopic := p.proccFuncRegistered[topic]
	if !okTopic {
		p.proccFuncRegistered[topic] = make([]AnalyticsModuleProccFunc, 0, 100)
		topicProccFunctions = p.proccFuncRegistered[topic]
	}
	topicProccFunctions = append(topicProccFunctions, proccFunc)
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
