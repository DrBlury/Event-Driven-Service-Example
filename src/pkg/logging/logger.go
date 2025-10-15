package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"

	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func SetLogger(loggerConfig *Config) *slog.Logger {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	slogLevel := getSlogLevel(loggerConfig)

	var logger *slog.Logger
	slog.With("Logger", loggerConfig.Logger).Info("Using logger")
	// set slog to json or console
	switch strings.ToLower(loggerConfig.Logger) {
	case "json":
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slogLevel,
		}))
	case "prettyjson":
		logger = slog.New(NewPrettyHandler(&slog.HandlerOptions{
			Level:     slogLevel,
			AddSource: true,
		}))
	case "otel":
		// ===== Open Telemetry =====
		logger = slog.New(createOtelHandler(loggerConfig))
	case "otel-and-console":
		// ===== Open Telemetry and Console =====
		logger = createMultiLogger(loggerConfig, slogLevel)
	default:
		slog.Warn("LOGGER environment variable not set to json, using default console logger")
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slogLevel,
		}))
	}

	logger.With("Loglevel", loggerConfig.LogLevel, "Logformat", loggerConfig.Logger).Info("Logger set")
	logger.Debug("DEBUG MODE IS ACTIVE")
	return logger
}

func createOtelHandler(loggerConfig *Config) *otelslog.Handler {
	otelLog, err := NewOtelLog(context.Background(), loggerConfig)
	if err != nil {
		slog.With("Error", err).Error("Error creating OpenTelemetry logger")
		return nil
	}
	otelHandler := otelslog.NewHandler(loggerConfig.ServiceName,
		otelslog.WithLoggerProvider(otelLog.logProvider),
	)
	return otelHandler
}

func createMultiLogger(loggerConfig *Config, slogLevel slog.Level) *slog.Logger {
	otelHandler := createOtelHandler(loggerConfig)

	prettyHandler := NewPrettyHandler(&slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: true,
	})

	// If creating the otel handler failed, fall back to only console/pretty logger.
	if otelHandler == nil {
		slog.Warn("Falling back to pretty console logger only")
		return slog.New(
			prettyHandler,
		)
	}

	wrapperHandler := NewMyWrapperHandler(otelHandler)
	logger := slog.New(
		slogmulti.Fanout(
			prettyHandler,
			wrapperHandler,
		),
	)
	return logger
}

func getSlogLevel(loggerConfig *Config) slog.Level {
	slogLevel := slog.LevelDebug

	switch loggerConfig.LogLevel {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slog.Warn("Invalid LOGGER_LEVEL environment variable, using default logger level DEBUG")
	}
	return slogLevel
}
