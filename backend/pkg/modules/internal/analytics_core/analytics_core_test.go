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
	"testing"
	"time"

	"github.com/openclarity/apiclarity/backend/pkg/pubsub"
)

const fixedPartition = 1

var counterProc int

type traceAnalyzerTest struct {
	t *testing.T
}

type messageForBrokerTest struct {
}

func (p messageForBrokerTest) GetPartitionKey() int64 {
	return int64(fixedPartition)
}

func (p traceAnalyzerTest) GetPriority() int {
	return 10
}
func (p traceAnalyzerTest) ProccFunc(topicName TopicType, dataFrames *ProcFuncDataFrames, partitionID int, message pubsub.MessageForBroker, annotations []interface{}, handler *AnalyticsCore) (newAnnotations []interface{}) {
	counter := int64(0)
	result, found := dataFrames.dataFrames[partitionID].Get("counter")
	if found {
		counter = result.(int64)
	}
	counter += 1
	dataFrames.dataFrames[partitionID].Set("counter", counter, 10*time.Minute)

	err := handler.PublishMessage(EntityTopicName, message)
	if err != nil {
		p.t.Errorf("Failed to publish by entity")
	}
	if topicName != TraceTopicName {
		p.t.Errorf("Wrong topic " + string(topicName) + " instead of " + string(TraceTopicName))
	}
	if partitionID != fixedPartition {
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
func (p entityAnalyzerTest) ProccFunc(topicName TopicType, dataFrames *ProcFuncDataFrames, partitionID int, message pubsub.MessageForBroker, annotations []interface{}, handler *AnalyticsCore) (newAnnotations []interface{}) {
	if len(annotations) != p.priorityValue {
		p.t.Errorf("Improper order of proccFunction calls " + fmt.Sprint(len(annotations)))
	}
	if topicName != EntityTopicName {
		p.t.Errorf("Wrong topic " + string(topicName) + " instead of " + string(EntityTopicName))
	}

	if partitionID != fixedPartition {
		p.t.Errorf("Entity procc is sent to a wrong worker " + fmt.Sprint(partitionID))
	}

	counterProc++
	return append(annotations, 0)
}

func TestAnalyticsCore(t *testing.T) {
	counterProc = 0
	module, _ := newModule(context.TODO(), nil)
	var moduleAnalytics *AnalyticsCore
	switch m := module.(type) {
	case *AnalyticsCore:
		moduleAnalytics = m
	default:
		moduleAnalytics = nil
		t.Errorf("Failed to initialize analytics core")
	}

	moduleAnalytics.AddWorkers(2)
	traceAnalyzer := traceAnalyzerTest{
		t: t,
	}
	moduleAnalytics.RegisterAnalyticsModuleHandler(TraceTopicName, traceAnalyzer)

	entityAnalyzer4 := entityAnalyzerTest{
		priorityValue: 3,
		t:             t,
	}
	moduleAnalytics.RegisterAnalyticsModuleHandler(EntityTopicName, entityAnalyzer4)

	entityAnalyzer3 := entityAnalyzerTest{
		priorityValue: 2,
		t:             t,
	}
	moduleAnalytics.RegisterAnalyticsModuleHandler(EntityTopicName, entityAnalyzer3)
	entityAnalyzer1 := entityAnalyzerTest{
		priorityValue: 0,
		t:             t,
	}
	moduleAnalytics.RegisterAnalyticsModuleHandler(EntityTopicName, entityAnalyzer1)
	entityAnalyzer2 := entityAnalyzerTest{
		priorityValue: 1,
		t:             t,
	}
	moduleAnalytics.RegisterAnalyticsModuleHandler(EntityTopicName, entityAnalyzer2)
	msg := messageForBrokerTest{}
	err := moduleAnalytics.PublishMessage(TraceTopicName, msg)
	if err != nil {
		t.Error("Failed to publish message")
	}
	err = moduleAnalytics.PublishMessage(TraceTopicName, msg)
	if err != nil {
		t.Error("Failed to publish message")
	}
	time.Sleep(3 * time.Second)

	if counterProc != 10 {
		t.Error("Didn't pass all the procc functions")
	}

	// During this test, the same partition is always used
	selectedDataFrame := moduleAnalytics.dataFramesRegistered[traceAnalyzer].dataFrames[fixedPartition]
	result, found := selectedDataFrame.Get("counter")
	if !found {
		t.Errorf("Unable to find counter entry in dataframe[%d]", fixedPartition)
	}
	counter := result.(int64)
	if counter != 2 {
		t.Errorf("Counter has wrong value. Got %d, expected %d", counter, 2)
	}
}
