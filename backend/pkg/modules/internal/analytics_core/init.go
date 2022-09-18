package analytics_core

import (
	"context"
	"net/http"
	"sort"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/pubsub"

	log "github.com/sirupsen/logrus"
)

const (
	moduleVersion        = "0.0.0"
	TraceTopicName       = "trace"
	ApiTopicName         = "api"
	ApiEndpointTopicName = "api_endpoint"
	ObjectTopicName      = "object"
	EntityTopicName      = "entity"
)

type TopicType string

type AnalyticsCore struct {
	httpHandler         http.Handler
	msgBroker           *pubsub.Handler
	accessor            core.BackendAccessor
	info                *core.ModuleInfo
	numWorkers          int
	proccFuncRegistered map[TopicType][]AnalyticsModuleProccFunc
	topics              []TopicType
}

type ProcFuncDataFrames struct {
	dataFrames map[int]*interface{}
}

type AnalyticsModuleProccFunc interface {
	GetPriority() int
	ProccFunc(topicName TopicType, dataFrames *ProcFuncDataFrames, partitionId int, message pubsub.MessageForBroker, annotations []interface{}, handler *AnalyticsCore) (new_annotations []interface{})
}

func (p *AnalyticsCore) Info() core.ModuleInfo {
	return *p.info
}

func (p *AnalyticsCore) Name() string              { return "analytics_core" }
func (p *AnalyticsCore) HTTPHandler() http.Handler { return p.httpHandler }

func (p *AnalyticsCore) handlerFunction(topic TopicType, paritionId int, msgChannel chan pubsub.MessageForBroker) {
	for {
		message := <-msgChannel
		topicProccFunctions, okTopic := p.proccFuncRegistered[topic]
		if okTopic {
			annotations := make([]interface{}, 0, 100)
			for _, proccFunction := range topicProccFunctions {
				dataFrames := &ProcFuncDataFrames{
					dataFrames: nil,
				}
				annotations = proccFunction.ProccFunc(topic, dataFrames, paritionId, message, annotations, p)
			}

		}
	}

}

func newModule(ctx context.Context, accessor core.BackendAccessor) (_ core.Module, err error) {
	p := &AnalyticsCore{
		httpHandler:         nil,
		msgBroker:           nil,
		accessor:            accessor,
		info:                &core.ModuleInfo{Name: "analytics_core", Description: "analytics_core"},
		numWorkers:          1,
		proccFuncRegistered: map[TopicType][]AnalyticsModuleProccFunc{},
		topics:              make([]TopicType, 0, 100),
	}
	/* We do not need to expose API at this point
	handler := &httpHandler{
		accessor: accessor,
	}
	p.httpHandler = restapi.HandlerWithOptions(handler, restapi.ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + "analytics_core"})
	*/
	p.msgBroker = pubsub.NewHandler()

	p.InitTopic(TraceTopicName)
	p.InitTopic(ApiTopicName)
	p.InitTopic(ApiEndpointTopicName)
	p.InitTopic(ObjectTopicName)
	p.InitTopic(EntityTopicName)

	return p, nil
}

func orderHandlerFuncsByPriority(proccFunctions []AnalyticsModuleProccFunc) []AnalyticsModuleProccFunc {
	sort.Slice(proccFunctions, func(i, j int) bool {
		return proccFunctions[i].GetPriority() < proccFunctions[j].GetPriority()
	})
	return proccFunctions
}

func (p *AnalyticsCore) RegisterAnalyticsModuleHandler(topic TopicType, proccFunc AnalyticsModuleProccFunc) {
	_, okTopic := p.proccFuncRegistered[topic]
	if !okTopic {
		p.proccFuncRegistered[topic] = make([]AnalyticsModuleProccFunc, 0, 100)
	}

	p.proccFuncRegistered[topic] = append(p.proccFuncRegistered[topic], proccFunc)
	p.proccFuncRegistered[topic] = orderHandlerFuncsByPriority(p.proccFuncRegistered[topic])
}

/*
type httpHandler struct {
	accessor core.BackendAccessor
}
*/

func (p *AnalyticsCore) InitTopic(topicName TopicType) {
	for i := 0; i < p.numWorkers; i++ {
		topic_channel := p.msgBroker.AddSubscriptionShard(string(topicName), i)
		go p.handlerFunction(topicName, i, topic_channel)
	}
	p.topics = append(p.topics, topicName)
}

// It supports only increase from 1 that is default
// returns number of workers
func (p *AnalyticsCore) AddWorkers(numNewWorkers int) int {
	if numNewWorkers <= 0 {
		return 0
	}
	for i := p.numWorkers; i < (p.numWorkers + numNewWorkers); i++ {
		for _, topicName := range p.topics {
			topic_channel := p.msgBroker.AddSubscriptionShard(string(topicName), i)
			go p.handlerFunction(topicName, i, topic_channel)
		}
	}
	p.numWorkers += numNewWorkers
	return numNewWorkers
}

func (p *AnalyticsCore) PublishMessage(topicName TopicType, message pubsub.MessageForBroker) (err bool) {
	return p.msgBroker.PublishByPartitionKey(string(topicName), message)
}

func (p *AnalyticsCore) EventNotify(ctx context.Context, event *core.Event) {
	if err := p.eventNotify(ctx, event); err != nil {
		log.Errorf("[Analytics-Core] EventNotify: %s", err)
	}
}

func (p *AnalyticsCore) eventNotify(ctx context.Context, event *core.Event) (err error) {
	log.Debugf("[Analytics-Core] received a new event for API(%v) Event(%v) ", event.APIEvent.APIInfoID, event.APIEvent.ID)
	log.Errorf("[Analytics-Core] received a new event for API(%v) Event(%v) ", event.APIEvent.APIInfoID, event.APIEvent.ID)

	return nil
}

/*
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
*/
//nolint:gochecknoinits
func init() {
	core.RegisterModule(newModule)
}