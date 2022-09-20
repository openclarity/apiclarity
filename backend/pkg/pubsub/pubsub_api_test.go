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
package pubsub

import (
	"strconv"
	"testing"
	"time"
)

type stringMessageForBroker struct {
	s string
}

func (p stringMessageForBroker) GetPartitionKey() int64 {
	return int64(0)
}

func handlerFunc(t *testing.T, topicName string, shardID int, inChannel chan MessageForBroker, outChannel chan string) {
	message := <-inChannel
	switch m := message.(type) {
	case stringMessageForBroker:
		outChannel <- m.s + "_" + topicName + "_" + strconv.Itoa(shardID)
	default:
		t.Errorf("Wrong message type")
	}
}

func TestPubSubAPI(t *testing.T) {
	handler := NewHandler()
	channelAbc0, _ := handler.AddSubscriptionShard("abc")
	channelAbc1, _ := handler.AddSubscriptionShard("abc")
	channelDef2, _ := handler.AddSubscriptionShard("def")
	channelDef3, _ := handler.AddSubscriptionShard("def")

	outChannel := make(chan string)

	go handlerFunc(t, "abc", 0, channelAbc0, outChannel)
	go handlerFunc(t, "abc", 1, channelAbc1, outChannel)
	go handlerFunc(t, "def", 0, channelDef2, outChannel)
	go handlerFunc(t, "def", 1, channelDef3, outChannel)

	//channelAbc0 <- stringMessageForBroker{ s: "test" }
	err := handler.Publish("abc", 0, stringMessageForBroker{s: "test"})
	if err != nil {
		t.Errorf("failed to publish message")
	}
	time.Sleep(3 * time.Second)

	val, ok := <-outChannel
	if !ok {
		t.Errorf("failed to read from channel")
	}
	if val != "test_abc_0" {
		t.Errorf("handlerFunc() got = %v, want %v", val, "test_abc_0")
	}
}
