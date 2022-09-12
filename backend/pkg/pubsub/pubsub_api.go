package pubsub

type topicSubscriptions struct {
	shards map[int]chan interface{}
}

type Handler struct {
	useKafkaInterface bool
	subscriptions     map[string]*topicSubscriptions
}

func NewHandler() (_ *Handler) {
	h := &Handler{useKafkaInterface: false, subscriptions: make(map[string]*topicSubscriptions)}
	return h
}

func (h *Handler) AddSubscriptionShard(topicName string, shardId int) (_ chan interface{}) {
	i := make(chan interface{}, 1000)
	_, ok := h.subscriptions[topicName]
	if !ok {
		h.subscriptions[topicName] = &topicSubscriptions{shards: make(map[int]chan interface{})}
	}
	topicShards := h.subscriptions[topicName]

	_, ok = topicShards.shards[shardId]
	if !ok {
		topicShards.shards[shardId] = i
	}

	return i
}

func (h *Handler) Publish(topicName string, shardId int, message interface{}) (err bool) {
	_, ok := h.subscriptions[topicName]
	if !ok {
		return true
	}
	topicShards := h.subscriptions[topicName]

	i, okShards := topicShards.shards[shardId]
	if !okShards {
		return true
	}
	i <- message
	return false
}
