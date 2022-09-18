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

func handlerFunc(t *testing.T, topicName string, shardId int, inChannel chan MessageForBroker, outChannel chan string) {
	message := <-inChannel
	switch m := message.(type) {
	case stringMessageForBroker:
		outChannel <- m.s + "_" + topicName + "_" + strconv.Itoa(shardId)
	default:
		t.Errorf("Wrong message type")
	}
}

func TestPubSubAPI(t *testing.T) {
	handler := NewHandler()
	channelAbc0 := handler.AddSubscriptionShard("abc", 0)
	channelAbc1 := handler.AddSubscriptionShard("abc", 1)
	channelDef2 := handler.AddSubscriptionShard("def", 0)
	channelDef3 := handler.AddSubscriptionShard("def", 1)

	outChannel := make(chan string)

	go handlerFunc(t, "abc", 0, channelAbc0, outChannel)
	go handlerFunc(t, "abc", 1, channelAbc1, outChannel)
	go handlerFunc(t, "def", 0, channelDef2, outChannel)
	go handlerFunc(t, "def", 1, channelDef3, outChannel)

	//channelAbc0 <- stringMessageForBroker{ s: "test" }
	handler.Publish("abc", 0, stringMessageForBroker{s: "test"})
	time.Sleep(3 * time.Second)

	val, ok := <-outChannel
	if !ok {
		t.Errorf("failed to read from channel")
	}
	if val != "test_abc_0" {
		t.Errorf("handlerFunc() got = %v, want %v", val, "test_abc_0")
	}
}
