package pubsub

import "math"

type topicSubscriptions struct {
	paritions     map[int]chan interface{}
	numPartitions int
}

type Handler struct {
	useKafkaInterface bool
	subscriptions     map[string]*topicSubscriptions
}

func NewHandler() (_ *Handler) {
	h := &Handler{useKafkaInterface: false, subscriptions: make(map[string]*topicSubscriptions)}
	return h
}

func (h *Handler) AddSubscriptionShard(topicName string, partitionId int) (_ chan interface{}) {
	i := make(chan interface{}, 1000)
	_, ok := h.subscriptions[topicName]
	if !ok {
		h.subscriptions[topicName] = &topicSubscriptions{paritions: make(map[int]chan interface{})}
	}
	topicPartitions := h.subscriptions[topicName]

	_, ok = topicPartitions.paritions[partitionId]
	if !ok {
		topicPartitions.paritions[partitionId] = i
	}
	topicPartitions.numPartitions = int(math.Max(float64(topicPartitions.numPartitions), float64(partitionId)) + 1)

	return i
}

func (h *Handler) PublishByPartitionKey(topicName string, partitionKey int64, message interface{}) (err bool) {
	_, ok := h.subscriptions[topicName]
	if !ok {
		return true
	}
	topicPartitions := h.subscriptions[topicName]

	partitionId := int(partitionKey % int64(topicPartitions.numPartitions))

	return h.Publish(topicName, partitionId, message)

}

func (h *Handler) Publish(topicName string, partitionId int, message interface{}) (err bool) {
	_, ok := h.subscriptions[topicName]
	if !ok {
		return true
	}
	topicPartitions := h.subscriptions[topicName]

	i, okShards := topicPartitions.paritions[partitionId]
	if !okShards {
		return true
	}
	i <- message
	return false
}
