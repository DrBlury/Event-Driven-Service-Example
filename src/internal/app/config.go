package app

import (
	"strings"
	"time"

	"drblury/event-driven-service/internal/database"
	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/server"
	"drblury/event-driven-service/pkg/events"
	"drblury/event-driven-service/pkg/logging"
	"drblury/event-driven-service/pkg/metrics"
	"drblury/event-driven-service/pkg/router"
	"drblury/event-driven-service/pkg/tracing"

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
	viper.SetDefault("APP_SERVER_HIDE_HEADERS", []string{"Authorization", "Proxy-Authorization", "Cookie", "Set-Cookie"})

	// Logger
	viper.SetDefault("LOGGER", "json")
	viper.SetDefault("LOGGER_LEVEL", "debug")

	// Tracing
	viper.SetDefault("TRACING_ENABLED", false)
	viper.SetDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
	viper.SetDefault("SERVICE_NAME", "example-service")

	// Events / Middleware defaults
	viper.SetDefault("EVENTS_RETRY_MAX_RETRIES", 5)
	viper.SetDefault("EVENTS_RETRY_INITIAL_INTERVAL", time.Second)
	viper.SetDefault("EVENTS_RETRY_MAX_INTERVAL", 16*time.Second)
}

func LoadConfig(
	version string,
	buildDate string,
	details string,
	commitHash string,
	commitDate string,
) (*Config, error) {
	SetDefaults()
	viper.AutomaticEnv()

	return &Config{
		Info:     loadInfoConfig(version, buildDate, details, commitHash, commitDate),
		Router:   loadRouterConfig(),
		Server:   loadServerConfig(),
		Database: loadDatabaseConfig(),
		Logger:   loadLoggerConfig(),
		Tracing:  loadTracingConfig(),
		Metrics:  loadMetricsConfig(),
		Events:   loadEventsConfig(),
	}, nil
}

func loadInfoConfig(
	version string,
	buildDate string,
	details string,
	commitHash string,
	commitDate string,
) *domain.Info {
	return &domain.Info{
		Version:    version,
		BuildDate:  buildDate,
		Details:    details,
		CommitHash: commitHash,
		CommitDate: commitDate,
	}
}

func loadRouterConfig() *router.Config {
	return &router.Config{
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
}

func loadServerConfig() *server.Config {
	return &server.Config{
		Address: "0.0.0.0:" + viper.GetString("APP_SERVER_PORT"),
		BaseURL: viper.GetString("APP_SERVER_BASE_URL"),
	}
}

func loadLoggerConfig() *logging.Config {
	loggerSelection := strings.ToLower(viper.GetString("LOGGER"))
	consoleFormat := logging.ParseFormat(loggerSelection)
	if consoleFormat == "" {
		consoleFormat = logging.FormatJSON
	}

	consoleEnabled := true
	otelEnabled := false
	otelMirror := false

	switch loggerSelection {
	case "otel":
		consoleEnabled = false
		otelEnabled = true
	case "otel-and-console":
		consoleEnabled = true
		otelEnabled = true
		otelMirror = true
		consoleFormat = logging.FormatPretty
	}

	return &logging.Config{
		Level:          viper.GetString("LOGGER_LEVEL"),
		Format:         consoleFormat,
		AddSource:      true,
		ConsoleEnabled: consoleEnabled,
		SetAsDefault:   true,
		OTel: logging.OTelConfig{
			Enabled:         otelEnabled,
			MirrorToConsole: otelMirror,
			Endpoint:        viper.GetString("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"),
			Headers:         viper.GetString("OTEL_EXPORTER_OTLP_LOGS_HEADERS"),
			ServiceName:     viper.GetString("APP_NAME"),
			ServiceVersion:  viper.GetString("VERSION"),
		},
	}
}

func loadTracingConfig() *tracing.Config {
	return &tracing.Config{
		Enabled:            viper.GetBool("TRACING_ENABLED"),
		OtelEndpoint:       viper.GetString("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"),
		Headers:            viper.GetString("OTEL_EXPORTER_OTLP_TRACES_HEADERS"),
		OTELTracesExporter: viper.GetString("OTEL_TRACES_EXPORTER"),
		ServiceName:        viper.GetString("APP_NAME"),
		ServiceVersion:     viper.GetString("VERSION"),
	}
}

func loadMetricsConfig() *metrics.Config {
	return &metrics.Config{
		Enabled:             viper.GetBool("METRICS_ENABLED"),
		OtelEndpoint:        viper.GetString("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"),
		Headers:             viper.GetString("OTEL_EXPORTER_OTLP_METRICS_HEADERS"),
		OTELMetricsExporter: viper.GetString("OTEL_METRICS_EXPORTER"),
		ServiceName:         viper.GetString("APP_NAME"),
		ServiceVersion:      viper.GetString("VERSION"),
	}
}

func loadEventsConfig() *events.Config {
	return &events.Config{
		PubSubSystem:         viper.GetString("PUBSUB_SYSTEM"),
		KafkaBrokers:         viper.GetStringSlice("KAFKA_BROKERS_URL"),
		KafkaClientID:        viper.GetString("KAFKA_CLIENT_ID"),
		KafkaConsumerGroup:   viper.GetString("KAFKA_CONSUMER_GROUP_ID"),
		RabbitMQURL:          viper.GetString("RABBITMQ_URL"),
		AWSRegion:            viper.GetString("AWS_REGION"),
		AWSAccessKeyID:       viper.GetString("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey:   viper.GetString("AWS_SECRET_ACCESS_KEY"),
		AWSAccountID:         viper.GetString("AWS_ACCOUNT_ID"),
		AWSEndpoint:          viper.GetString("AWS_ENDPOINT_URL"),
		PoisonQueue:          viper.GetString("POISON_QUEUE"),
		ConsumeQueue:         viper.GetString("QUEUE"),
		PublishQueue:         viper.GetString("QUEUE_PROCESSED"),
		ConsumeQueueSignup:   viper.GetString("QUEUE_SIGNUP"),
		PublishQueueSignup:   viper.GetString("QUEUE_SIGNUP_PROCESSABLE"),
		RetryMaxRetries:      viper.GetInt("EVENTS_RETRY_MAX_RETRIES"),
		RetryInitialInterval: viper.GetDuration("EVENTS_RETRY_INITIAL_INTERVAL"),
		RetryMaxInterval:     viper.GetDuration("EVENTS_RETRY_MAX_INTERVAL"),
	}
}

func loadDatabaseConfig() *database.Config {
	return &database.Config{
		MongoURL:      viper.GetString("MONGO_URL"),
		MongoDB:       viper.GetString("MONGO_DB"),
		MongoUser:     viper.GetString("MONGO_USER"),
		MongoPassword: viper.GetString("MONGO_PASSWORD"),
	}
}
