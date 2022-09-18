package analytics_core

import (
	"fmt"
	"github.com/openclarity/apiclarity/backend/pkg/pubsub"
	"strconv"
	"testing"
	"time"
)

var counterProc int = 0

func handlerFunc(t *testing.T, topicName string, shardId int, inChannel chan interface{}, outChannel chan string) {
	message := <-inChannel
	switch m := message.(type) {
	case string:
		outChannel <- m + "_" + topicName + "_" + strconv.Itoa(shardId)
	default:
		t.Errorf("Wrong message type")
	}
}

type traceAnalyzerTest struct {
	t *testing.T
}

type messageForBrokerTest struct {
}

func (p messageForBrokerTest) GetPartitionKey() int64 {
	return int64(1)
}

func (p traceAnalyzerTest) GetPriority() int {
	return 10
}
func (p traceAnalyzerTest) ProccFunc(topicName TopicType, dataFrames *ProcFuncDataFrames, partitionId int, message pubsub.MessageForBroker, annotations []interface{}, handler *AnalyticsCore) (new_annotations []interface{}) {
	err := handler.PublishMessage(EntityTopicName, message)
	if err {
		p.t.Errorf("Failed to publish by entity")
	}
	if topicName != TraceTopicName {
		p.t.Errorf("Wromg topic " + string(topicName) + " instead of " + string(TraceTopicName))
	}
	if partitionId != 1 {
		p.t.Errorf("Trace procc is sent to a wrong worker " + fmt.Sprint(partitionId) + " " + fmt.Sprint(message.GetPartitionKey()) + " " + fmt.Sprint(handler.msgBroker.GetNumPartitionts(TraceTopicName)))
	}

	counterProc++
	return annotations
}

type entityAnalyzerTest struct {
	priorityValue int
	t             *testing.T
}

func (p entityAnalyzerTest) GetPriority() int {
	return p.priorityValue
}
func (p entityAnalyzerTest) ProccFunc(topicName TopicType, dataFrames *ProcFuncDataFrames, partitionId int, message pubsub.MessageForBroker, annotations []interface{}, handler *AnalyticsCore) (new_annotations []interface{}) {
	if len(annotations) != p.priorityValue {
		p.t.Errorf("Improper order of proccFunction calls " + fmt.Sprint(len(annotations)))
	}
	if topicName != EntityTopicName {
		p.t.Errorf("Wromg topic " + string(topicName) + " instead of " + string(EntityTopicName))
	}

	if partitionId != 1 {
		p.t.Errorf("Entity procc is sent to a wrong worker " + fmt.Sprint(partitionId))
	}

	counterProc++
	return append(annotations, 0)
}

func TestAnalyticsCore(t *testing.T) {
	module, _ := newModule(nil, nil)
	var module_analytics *AnalyticsCore = nil
	switch m := module.(type) {
	case *AnalyticsCore:
		module_analytics = m
	default:
		t.Errorf("Failed to initialize analytics core")
	}

	module_analytics.AddWorkers(2)
	traceAnalyzer := traceAnalyzerTest{
		t: t,
	}
	module_analytics.RegisterAnalyticsModuleHandler(TraceTopicName, traceAnalyzer)

	entityAnalyzer4 := entityAnalyzerTest{
		priorityValue: 3,
		t:             t,
	}
	module_analytics.RegisterAnalyticsModuleHandler(EntityTopicName, entityAnalyzer4)

	entityAnalyzer3 := entityAnalyzerTest{
		priorityValue: 2,
		t:             t,
	}
	module_analytics.RegisterAnalyticsModuleHandler(EntityTopicName, entityAnalyzer3)
	entityAnalyzer1 := entityAnalyzerTest{
		priorityValue: 0,
		t:             t,
	}
	module_analytics.RegisterAnalyticsModuleHandler(EntityTopicName, entityAnalyzer1)
	entityAnalyzer2 := entityAnalyzerTest{
		priorityValue: 1,
		t:             t,
	}
	module_analytics.RegisterAnalyticsModuleHandler(EntityTopicName, entityAnalyzer2)
	msg := messageForBrokerTest{}
	module_analytics.PublishMessage(TraceTopicName, msg)
	time.Sleep(3 * time.Second)

	if counterProc != 5 {
		t.Error("Didn't pass all the procc functions")
	}

}
