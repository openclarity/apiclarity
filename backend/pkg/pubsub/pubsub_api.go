package pubsub

type topicSubscriptions struct {
	paritions     map[int]chan MessageForBroker
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

func (h *Handler) GetNumPartitionts(topicName string) int {
	subTopic, ok := h.subscriptions[topicName]
	if !ok {
		return 0
	}

	return subTopic.numPartitions
}

func (h *Handler) AddSubscriptionShard(topicName string, partitionId int) (_ chan MessageForBroker) {
	i := make(chan MessageForBroker, 1000)
	_, ok := h.subscriptions[topicName]
	if !ok {
		h.subscriptions[topicName] = &topicSubscriptions{paritions: make(map[int]chan MessageForBroker), numPartitions: 0}
	}
	topicPartitions := h.subscriptions[topicName]

	_, ok = topicPartitions.paritions[partitionId]
	if !ok {
		topicPartitions.paritions[partitionId] = i
	}
	topicPartitions.numPartitions++

	return i
}

func (h *Handler) PublishByPartitionKey(topicName string, message MessageForBroker) (err bool) {
	_, ok := h.subscriptions[topicName]
	if !ok {
		return true
	}
	topicPartitions := h.subscriptions[topicName]

	partitionId := int(message.GetPartitionKey() % int64(topicPartitions.numPartitions))

	return h.Publish(topicName, partitionId, message)

}

func (h *Handler) Publish(topicName string, partitionId int, message MessageForBroker) (err bool) {
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
