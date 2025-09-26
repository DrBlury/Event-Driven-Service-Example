package events

type Config struct {
	KafkaBrokers       []string
	ConsumeTopic       string
	PublishTopic       string
	KafkaClientID      string
	KafkaConsumerGroup string
}

func NewConfig() Config {
	return Config{
		KafkaBrokers:       []string{"kafka:9092"},
		ConsumeTopic:       "events",
		PublishTopic:       "events-processed",
		KafkaClientID:      "mqtt-service-flow-example",
		KafkaConsumerGroup: "mqtt-service-flow-example-group",
	}
}
