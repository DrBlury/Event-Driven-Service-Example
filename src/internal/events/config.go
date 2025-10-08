package events

type Config struct {
	// PubSubSystem
	PubSubSystem string // "kafka" or "rabbitmq"

	// Kafka configuration
	KafkaBrokers       []string
	KafkaClientID      string
	KafkaConsumerGroup string

	// RabbitMQ configuration
	RabbitMQURL string

	// === All Queues ===
	PoisonQueue string

	// Demo Queues
	ConsumeQueue string
	PublishQueue string

	// Signup Usecase Queues
	ConsumeQueueSignup string
	PublishQueueSignup string
}
