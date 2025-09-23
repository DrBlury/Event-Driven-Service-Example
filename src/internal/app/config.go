package app

import (
	"drblury/poc-event-signup/internal/domain"
	"drblury/poc-event-signup/internal/server"
	"drblury/poc-event-signup/pkg/logging"
	"drblury/poc-event-signup/pkg/router"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Info   *domain.Info
	Router *router.Config
	Server *server.Config
	Logger *logging.Config
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
		OtelEndpoint:      viper.GetString("LOGGER_OTEL_ENDPOINT"),
		OtelAuthorization: viper.GetString("OTEL_EXPORTER_OTLP_AUTHORIZATION"),
		ServiceName:       viper.GetString("APP_NAME"),
		ServiceVersion:    viper.GetString("VERSION"),
		LogLevel:          viper.GetString("LOGGER_LEVEL"),
		Logger:            viper.GetString("LOGGER"),
	}

	return &Config{
		Info:   infoConfig,
		Router: routerConfig,
		Server: serverConfig,
		Logger: loggerConfig,
	}, nil
}
