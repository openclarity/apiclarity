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

package analyticscore

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/backend/pkg/dataframe"
	// cache "github.com/openclarity/apiclarity/backend/pkg/dataframe/ristretto"
	// cache "github.com/openclarity/apiclarity/backend/pkg/dataframe/gcache"
	cache "github.com/openclarity/apiclarity/backend/pkg/dataframe/gocache"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/pubsub"
)

const (
	TraceTopicName          = "trace"
	APITopicName            = "api"
	APIEndpointTopicName    = "api_endpoint"
	ObjectTopicName         = "object"
	EntityTopicName         = "entity"
	annotationArrayCapacity = 100
	maxNumTopics            = 100
	maxNumProccFuncPerTopic = 100
	defaultDataFrameTTL     = 10 * time.Minute
)

type (
	TopicType   string
	DataFrameID string
	DataFrame   map[int]dataframe.DataFrame
)

type AnalyticsCore struct {
	httpHandler         http.Handler
	msgBroker           *pubsub.Handler
	accessor            core.BackendAccessor
	info                *core.ModuleInfo
	numWorkers          int
	proccFuncRegistered map[TopicType][]AnalyticsModuleProccFunc
	dataFramesRegistry  DataFramesRegistry
	topics              []TopicType
}

type DataFramesRegistry struct {
	dataFrames map[DataFrameID]DataFrame
}

func NewDataFramesRegistry(numWorkers int) DataFramesRegistry {
	return DataFramesRegistry{
		dataFrames: map[DataFrameID]DataFrame{},
	}
}

func (dfr DataFramesRegistry) NewDataFrame(name DataFrameID, shards int) error {
	dfr.dataFrames[name] = make(map[int]dataframe.DataFrame)
	for i := 0; i < shards; i++ {
		df := cache.DataFrame{}
		err := df.Init(defaultDataFrameTTL)
		if err != nil {
			return fmt.Errorf("unable to create shard %d of dataframe %s", i, name)
		}
		dfr.dataFrames[name][i] = &df
	}

	return nil
}

func (dfr DataFramesRegistry) Get(name DataFrameID) (DataFrame, bool) {
	df, found := dfr.dataFrames[name]
	return df, found
}

type AnalyticsModuleProccFunc interface {
	GetPriority() int
	ProccFunc(topicName TopicType, dataFramesRegistry DataFramesRegistry, partitionID int, message pubsub.MessageForBroker, annotations []interface{}, handler *AnalyticsCore) (newAnnotations []interface{})
}

func (p *AnalyticsCore) Info() core.ModuleInfo {
	return *p.info
}

func (p *AnalyticsCore) Name() string              { return "analytics_core" }
func (p *AnalyticsCore) HTTPHandler() http.Handler { return p.httpHandler }

func (p *AnalyticsCore) handlerFunction(topic TopicType, partitionID int, msgChannel chan pubsub.MessageForBroker) {
	for {
		message := <-msgChannel
		topicProccFunctions, okTopic := p.proccFuncRegistered[topic]
		if okTopic && len(topicProccFunctions) > 0 {
			annotations := make([]interface{}, 0, annotationArrayCapacity)
			for _, proccFunction := range topicProccFunctions {
				annotations = proccFunction.ProccFunc(topic, p.dataFramesRegistry, partitionID, message, annotations, p)
			}
		}
	}
}

//nolint:unparam
func newModule(ctx context.Context, accessor core.BackendAccessor) (_ core.Module, err error) {
	p := &AnalyticsCore{
		httpHandler:         nil,
		msgBroker:           nil,
		accessor:            accessor,
		info:                &core.ModuleInfo{Name: "analytics_core", Description: "analytics_core"},
		numWorkers:          1,
		proccFuncRegistered: map[TopicType][]AnalyticsModuleProccFunc{},
		dataFramesRegistry:  NewDataFramesRegistry(1),
		topics:              make([]TopicType, 0, maxNumTopics),
	}
	/* We do not need to expose API at this point
	handler := &httpHandler{
		accessor: accessor,
	}
	p.httpHandler = restapi.HandlerWithOptions(handler, restapi.ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + "analytics_core"})
	*/
	p.msgBroker = pubsub.NewHandler()

	p.InitTopic(TraceTopicName)
	p.InitTopic(APITopicName)
	p.InitTopic(APIEndpointTopicName)
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
		p.proccFuncRegistered[topic] = make([]AnalyticsModuleProccFunc, 0, maxNumProccFuncPerTopic)
	}

	p.proccFuncRegistered[topic] = append(p.proccFuncRegistered[topic], proccFunc)
	p.proccFuncRegistered[topic] = orderHandlerFuncsByPriority(p.proccFuncRegistered[topic])
}

/*
type httpHandler struct {
	accessor core.BackendAccessor
}
*/
// WARNING: This function shall be executed only during initialization step. Running this function.
// when the message broker is in use may generate synchronization problem due to lack of lock on topics map.
// this is done on purpose to avoid unnecessary lock overhead.
func (p *AnalyticsCore) InitTopic(topicName TopicType) {
	for i := 0; i < p.numWorkers; i++ {
		topicChannel, _ := p.msgBroker.AddSubscriptionShard(string(topicName))
		go p.handlerFunction(topicName, i, topicChannel)
	}
	p.topics = append(p.topics, topicName)
}

// It supports only increase from 1 that is default
// returns number of workers
// WARNING: This function shall be executed only during initialization step. Running this function
// when the message broker is in use may generate synchronization problem due to lack of lock on topics map
// this is done on purpose to avoid unnecessary lock overhead.
func (p *AnalyticsCore) AddWorkers(numNewWorkers int) int {
	if numNewWorkers <= 0 {
		return 0
	}
	for i := p.numWorkers; i < (p.numWorkers + numNewWorkers); i++ {
		for _, topicName := range p.topics {
			topicChannel, _ := p.msgBroker.AddSubscriptionShard(string(topicName))
			go p.handlerFunction(topicName, i, topicChannel)
		}
	}
	p.numWorkers += numNewWorkers
	return numNewWorkers
}

func (p *AnalyticsCore) PublishMessage(topicName TopicType, message pubsub.MessageForBroker) (_ error) {
	err := p.msgBroker.PublishByPartitionKey(string(topicName), message)
	if err != nil {
		return fmt.Errorf("[Analytics-Core] failed to publish message %s", err)
	}
	return nil
}

func (p *AnalyticsCore) EventNotify(ctx context.Context, event *core.Event) {
	if err := p.eventNotify(ctx, event); err != nil {
		log.Errorf("[Analytics-Core] EventNotify: %s", err)
	}
}

//nolint:unparam
func (p *AnalyticsCore) eventNotify(ctx context.Context, event *core.Event) (err error) {
	log.Debugf("[Analytics-Core] received a new event for API(%v) Event(%v) ", event.APIEvent.APIInfoID, event.APIEvent.ID)
	log.Errorf("[Analytics-Core] received a new event for API(%v) Event(%v) ", event.APIEvent.APIInfoID, event.APIEvent.ID)
	if ctx == nil {
		log.Errorf("[Analytics-Core] No context")
	}
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
