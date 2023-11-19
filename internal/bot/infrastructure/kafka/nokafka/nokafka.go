// Package bot_infrastructure_kafka_nokafka is built so you don't need any kafka instance, but there's only a (badly) simulated one
// without any queues or anything
package bot_infrastructure_kafka_nokafka

import (
	bot_infrastructure_kafka "connectly-interview/internal/bot/infrastructure/kafka"
	"errors"
	"sync"
)

// NoKafka is a mock implementation of the Kafka interface that uses channels
type NoKafka struct {
	channels map[bot_infrastructure_kafka.Topic]chan []byte
	lock     sync.RWMutex
}

// NewNoKafka creates a new instance of NoKafka
func NewNoKafka() *NoKafka {
	channels := map[bot_infrastructure_kafka.Topic]chan []byte{
		bot_infrastructure_kafka.TopicPrompt: make(chan []byte),
	}
	return &NoKafka{
		channels: channels,
	}
}

// Send simulates sending a message to a Kafka topic
func (nk *NoKafka) Send(topic bot_infrastructure_kafka.Topic, msg []byte) error {
	nk.lock.RLock()
	defer nk.lock.RUnlock()

	ch, ok := nk.channels[topic]
	if !ok {
		return errors.New("topic not found")
	}

	ch <- msg
	return nil
}

// Listen simulates listening to a Kafka topic
func (nk *NoKafka) Listen(topic bot_infrastructure_kafka.Topic) (<-chan []byte, error) {
	nk.lock.Lock()
	defer nk.lock.Unlock()

	if _, ok := nk.channels[topic]; !ok {
		nk.channels[topic] = make(chan []byte, 100) // Buffer size can be adjusted as needed
	}

	return nk.channels[topic], nil
}
