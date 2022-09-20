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

import "fmt"

type topicSubscriptions struct {
	partitions    map[int]chan MessageForBroker
	numPartitions int
}

type Handler struct {
	useKafkaInterface bool
	subscriptions     map[string]*topicSubscriptions
}

type MessageForBroker interface {
	GetPartitionKey() int64
}

func NewHandler() (_ *Handler) {
	h := &Handler{useKafkaInterface: false, subscriptions: make(map[string]*topicSubscriptions)}
	return h
}

func (h *Handler) GetNumPartitions(topicName string) int {
	subTopic, ok := h.subscriptions[topicName]
	if !ok {
		return 0
	}

	return subTopic.numPartitions
}

// WARNING: This function shall be executed only during initialization step. Running this function
// when the message broker is in use may generate synchronization problem due to lack of lock on topics map
// this is done on purpose to avoid unneccessary lock overhead
func (h *Handler) AddSubscriptionShard(topicName string) (_ chan MessageForBroker, partitionId int) {

	_, ok := h.subscriptions[topicName]
	if !ok {
		h.subscriptions[topicName] = &topicSubscriptions{partitions: make(map[int]chan MessageForBroker), numPartitions: 0}
	}
	topicPartitions := h.subscriptions[topicName]

	partitionId = topicPartitions.numPartitions

	_, ok = topicPartitions.partitions[partitionId]
	if !ok {
		i := make(chan MessageForBroker, 1000)
		topicPartitions.partitions[partitionId] = i
		topicPartitions.numPartitions++
		return i, partitionId
	}
	return nil, partitionId
}

func (h *Handler) PublishByPartitionKey(topicName string, message MessageForBroker) (err error) {
	topicPartitions, ok := h.subscriptions[topicName]
	if !ok {
		return fmt.Errorf("no topic '%s' exists for topic", topicName)
	}

	partitionId := int(message.GetPartitionKey() % int64(topicPartitions.numPartitions))

	return h.Publish(topicName, partitionId, message)

}

func (h *Handler) Publish(topicName string, partitionId int, message MessageForBroker) (err error) {
	topicPartitions, ok := h.subscriptions[topicName]
	if !ok {
		return fmt.Errorf("no topic '%s' exists for topic", topicName)
	}

	i, okShards := topicPartitions.partitions[partitionId]
	if !okShards {
		return fmt.Errorf("no partition '%d' exists for topic '%s'", partitionId, topicName)
	}
	i <- message
	return nil
}
