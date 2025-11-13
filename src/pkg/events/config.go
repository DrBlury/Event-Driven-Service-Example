package events

import "time"

// Config groups the Pub/Sub settings required to initialise the
// Service. Each transport only uses the keys that are relevant to it.
type Config struct {
	// PubSubSystem selects the backing message infrastructure. Supported
	// values: "kafka", "rabbitmq", or "aws" (SNS/SQS).
	PubSubSystem string

	// Kafka configuration.
	KafkaBrokers       []string
	KafkaClientID      string
	KafkaConsumerGroup string

	// RabbitMQ configuration.
	RabbitMQURL string

	// PoisonQueue receives messages that cannot be processed even after
	// retries.
	PoisonQueue string

	// ConsumeQueue and PublishQueue are a generic pair you can reuse for
	// simple pipelines or demos.
	ConsumeQueue string
	PublishQueue string

	// ConsumeQueueSignup and PublishQueueSignup represent a second queue
	// pair that the example application uses to illustrate fan-out flows.
	ConsumeQueueSignup string
	PublishQueueSignup string

	// AWS (SNS/SQS) configuration.
	AWSRegion          string
	AWSAccountID       string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	// AWSEndpoint optionally points to a custom endpoint (for example,
	// Localstack in local development).
	AWSEndpoint string

	// RetryMiddleware tuning. Zero values fall back to library defaults.
	RetryMaxRetries      int
	RetryInitialInterval time.Duration
	RetryMaxInterval     time.Duration
}
