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
	"fmt"
	"github.com/openclarity/apiclarity/backend/pkg/pubsub"
	"testing"
	"time"
)

var counterProc int = 0

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
func (p traceAnalyzerTest) ProccFunc(topicName TopicType, dataFrames *ProcFuncDataFrames, partitionID int, message pubsub.MessageForBroker, annotations []interface{}, handler *AnalyticsCore) (new_annotations []interface{}) {
	err := handler.PublishMessage(EntityTopicName, message)
	if err != nil {
		p.t.Errorf("Failed to publish by entity")
	}
	if topicName != TraceTopicName {
		p.t.Errorf("Wrong topic " + string(topicName) + " instead of " + string(TraceTopicName))
	}
	if partitionID != 1 {
		p.t.Errorf("Trace procc is sent to a wrong worker " + fmt.Sprint(partitionID) + " " + fmt.Sprint(message.GetPartitionKey()) + " " + fmt.Sprint(handler.msgBroker.GetNumPartitions(TraceTopicName)))
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
		p.t.Errorf("Wrong topic " + string(topicName) + " instead of " + string(EntityTopicName))
	}

	if partitionId != 1 {
		p.t.Errorf("Entity procc is sent to a wrong worker " + fmt.Sprint(partitionId))
	}

	counterProc++
	return append(annotations, 0)
}

func TestAnalyticsCore(t *testing.T) {
	module, _ := newModuleRaw()
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
	err := module_analytics.PublishMessage(TraceTopicName, msg)
	if err != nil {
		t.Error("Failed to publish message")
	}
	time.Sleep(3 * time.Second)

	if counterProc != 5 {
		t.Error("Didn't pass all the procc functions")
	}

}
