package analytics_core

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"

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
	customTopics        []string
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
			annotations := make([]interface{}, 0, 100)
			for _, proccFunction := range topicProccFunctions {
				dataFrames := &ProcFuncDataFrames{
					dataFrames: nil,
				}
				proccFunction.HandlerFunc(dataFrames, message, annotations)
			}

		}
	}

}

func newModule(ctx context.Context, accessor core.BackendAccessor) (_ core.Module, err error) {
	p := &analyticsCore{
		httpHandler:         nil,
		msgBroker:           nil,
		accessor:            accessor,
		info:                &core.ModuleInfo{Name: "analytics_core", Description: "analytics_core"},
		numWorkers:          1,
		proccFuncRegistered: map[string][]AnalyticsModuleProccFunc{},
		customTopics:        make([]string, 0, 100),
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
		for _, customTopic := range p.customTopics {
			custom_topic_channel := p.msgBroker.AddSubscriptionShard(customTopic, i)
			go p.handlerFunction(customTopic, i, custom_topic_channel)
		}

	}
	return p, nil
}

type ProcFuncDataFrames struct {
	dataFrames map[int]*interface{}
}

type AnalyticsModuleProccFunc interface {
	GetPriority() int
	HandlerFunc(dataFrames *ProcFuncDataFrames, message interface{}, annotations []interface{})
}

func orderHandlerFuncsByPriority(proccFunctions []AnalyticsModuleProccFunc) []AnalyticsModuleProccFunc {
	sort.Slice(proccFunctions, func(i, j int) bool {
		return proccFunctions[i].GetPriority() < proccFunctions[j].GetPriority()
	})
	return proccFunctions
}

func (p *analyticsCore) RegisterAnalyticsModuleHandler(topic string, proccFunc AnalyticsModuleProccFunc) {
	_, okTopic := p.proccFuncRegistered[topic]
	if !okTopic {
		p.proccFuncRegistered[topic] = make([]AnalyticsModuleProccFunc, 0, 100)
	}

	p.proccFuncRegistered[topic] = append(p.proccFuncRegistered[topic], proccFunc)
	p.proccFuncRegistered[topic] = orderHandlerFuncsByPriority(p.proccFuncRegistered[topic])
}

type httpHandler struct {
	accessor core.BackendAccessor
}

func (p *analyticsCore) InitCustomTopic(customTopic string) {
	for i := 0; i < p.numWorkers; i++ {
		custom_topic_channel := p.msgBroker.AddSubscriptionShard(customTopic, i)
		go p.handlerFunction(customTopic, i, custom_topic_channel)
	}
}

// It supports only increase from 1 that is default
// returns number of workers
func (p *analyticsCore) AddWorkers(numNewWorkers int) int {
	if numNewWorkers <= 0 {
		return 0
	}
	for i := p.numWorkers; i < (p.numWorkers + numNewWorkers); i++ {
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
		for _, customTopic := range p.customTopics {
			custom_topic_channel := p.msgBroker.AddSubscriptionShard(customTopic, i)
			go p.handlerFunction(customTopic, i, custom_topic_channel)
		}
	}
	p.numWorkers += numNewWorkers
	return numNewWorkers
}

func (p *analyticsCore) PublishMessage(topicName string, partitionKey int64, message interface{}) (err bool) {
	return p.msgBroker.PublishByPartitionKey(topicName, partitionKey, message)
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
