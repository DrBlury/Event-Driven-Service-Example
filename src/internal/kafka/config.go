package events

type Config struct {
	KafkaBrokers       []string
	ConsumeTopic       string
	PublishTopic       string
	KafkaClientID      string
	KafkaConsumerGroup string
}
