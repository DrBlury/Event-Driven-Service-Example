package events

type Config struct {
	KafkaBrokers       []string
	KafkaClientID      string
	KafkaConsumerGroup string
	// === All Topics ===

	// Example Usecase Topics
	ConsumeTopic string
	PublishTopic string

	// Signup Usecase Topics
	ConsumeTopicSignup string
	PublishTopicSignup string
}
