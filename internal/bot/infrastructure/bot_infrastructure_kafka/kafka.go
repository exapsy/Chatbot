package bot_infrastructure_kafka

type Topic string

func (t Topic) String() string {
	return string(t)
}

const (
	TopicPrompt = "bot-prompt-message"
)

type Kafka interface {
	Send(topic Topic, msg []byte) error
}
