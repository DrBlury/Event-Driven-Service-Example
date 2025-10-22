package app

import (
	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/events"
	"drblury/event-driven-service/internal/server"
	"drblury/event-driven-service/pkg/logging"
	"drblury/event-driven-service/pkg/metrics"
	"drblury/event-driven-service/pkg/router"
	"drblury/event-driven-service/pkg/tracing"
	"time"

	"drblury/event-driven-service/internal/database"

	"github.com/spf13/viper"
)

type Config struct {
	Info     *domain.Info
	Router   *router.Config
	Server   *server.Config
	Database *database.Config
	Logger   *logging.Config
	Tracing  *tracing.Config
	Metrics  *metrics.Config
	Events   *events.Config
}

func SetDefaults() {
	viper.SetDefault("APP_NAME", "example-service")
	viper.SetDefault("APP_SERVER_PORT", "80")
	viper.SetDefault("APP_SERVER_TIMEOUT", 60*time.Second)
	viper.SetDefault("APP_SERVER_CORS_HEADERS", []string{"*"})
	viper.SetDefault("APP_SERVER_CORS_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("APP_SERVER_CORS_ORIGINS", []string{"*"})

	// Logger
	viper.SetDefault("LOGGER", "json")
	viper.SetDefault("LOGGER_LEVEL", "debug")

	// Tracing
	viper.SetDefault("TRACING_ENABLED", false)
	viper.SetDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
	viper.SetDefault("SERVICE_NAME", "example-service")
}

// nolint: funlen
func LoadConfig(
	version string,
	buildDate string,
	details string,
	commitHash string,
	commitDate string,
) (*Config, error) {
	SetDefaults()
	viper.AutomaticEnv()

	infoConfig := &domain.Info{
		Version:    version,
		BuildDate:  buildDate,
		Details:    details,
		CommitHash: commitHash,
		CommitDate: commitDate,
	}

	routerConfig := &router.Config{
		Timeout: viper.GetDuration("APP_SERVER_TIMEOUT"),
		CORS: router.CORSConfig{
			AllowCredentials: viper.GetBool("APP_SERVER_CORS_ALLOW_CREDENTIALS"),
			Headers:          viper.GetStringSlice("APP_SERVER_CORS_HEADERS"),
			Methods:          viper.GetStringSlice("APP_SERVER_CORS_METHODS"),
			Origins:          viper.GetStringSlice("APP_SERVER_CORS_ORIGINS"),
		},
		QuietdownRoutes: viper.GetStringSlice("APP_SERVER_QUIETDOWN_ROUTES"),
		HideHeaders:     viper.GetStringSlice("APP_SERVER_HIDE_HEADERS"),
	}

	serverConfig := &server.Config{
		Address: "0.0.0.0:" + viper.GetString("APP_SERVER_PORT"),
		BaseURL: viper.GetString("APP_SERVER_BASE_URL"),
	}

	loggerConfig := &logging.Config{
		OtelEndpoint:   viper.GetString("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"),
		Headers:        viper.GetString("OTEL_EXPORTER_OTLP_LOGS_HEADERS"),
		ServiceName:    viper.GetString("APP_NAME"),
		ServiceVersion: viper.GetString("VERSION"),
		LogLevel:       viper.GetString("LOGGER_LEVEL"),
		Logger:         viper.GetString("LOGGER"),
	}

	tracingConfig := &tracing.Config{
		Enabled:            viper.GetBool("TRACING_ENABLED"),
		OtelEndpoint:       viper.GetString("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"),
		Headers:            viper.GetString("OTEL_EXPORTER_OTLP_TRACES_HEADERS"),
		OTELTracesExporter: viper.GetString("OTEL_TRACES_EXPORTER"),
		ServiceName:        viper.GetString("APP_NAME"),
		ServiceVersion:     viper.GetString("VERSION"),
	}

	metricsConfig := &metrics.Config{
		Enabled:             viper.GetBool("METRICS_ENABLED"),
		OtelEndpoint:        viper.GetString("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"),
		Headers:             viper.GetString("OTEL_EXPORTER_OTLP_METRICS_HEADERS"),
		OTELMetricsExporter: viper.GetString("OTEL_METRICS_EXPORTER"),
		ServiceName:         viper.GetString("APP_NAME"),
		ServiceVersion:      viper.GetString("VERSION"),
	}

	eventsConfig := &events.Config{
		// PubSubSystem
		PubSubSystem: viper.GetString("PUBSUB_SYSTEM"), // "kafka", "aws" or "rabbitmq"

		// Kafka configuration
		KafkaBrokers:       viper.GetStringSlice("KAFKA_BROKERS_URL"),
		KafkaClientID:      viper.GetString("KAFKA_CLIENT_ID"),
		KafkaConsumerGroup: viper.GetString("KAFKA_CONSUMER_GROUP_ID"),

		// RabbitMQ configuration
		RabbitMQURL: viper.GetString("RABBITMQ_URL"),

		// AWS configuration
		AWSRegion:          viper.GetString("AWS_REGION"),
		AWSAccessKeyID:     viper.GetString("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: viper.GetString("AWS_SECRET_ACCESS_KEY"),
		AWSAccountID:       viper.GetString("AWS_ACCOUNT_ID"),
		AWSEndpoint:        viper.GetString("AWS_ENDPOINT_URL"),

		// === All Queues ===
		PoisonQueue: viper.GetString("POISON_QUEUE"),

		// Example Usecase Queues
		ConsumeQueue: viper.GetString("QUEUE"),
		PublishQueue: viper.GetString("QUEUE_PROCESSED"),

		// Signup Usecase Queues
		ConsumeQueueSignup: viper.GetString("QUEUE_SIGNUP"),
		PublishQueueSignup: viper.GetString("QUEUE_SIGNUP_PROCESSABLE"),
	}

	databaseConfig := &database.Config{
		MongoURL:      viper.GetString("MONGO_URL"),
		MongoDB:       viper.GetString("MONGO_DB"),
		MongoUser:     viper.GetString("MONGO_USER"),
		MongoPassword: viper.GetString("MONGO_PASSWORD"),
	}

	return &Config{
		Info:     infoConfig,
		Router:   routerConfig,
		Server:   serverConfig,
		Database: databaseConfig,
		Logger:   loggerConfig,
		Tracing:  tracingConfig,
		Metrics:  metricsConfig,
		Events:   eventsConfig,
	}, nil
}
