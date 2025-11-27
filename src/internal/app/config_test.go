package app

import (
	"testing"

	"github.com/spf13/viper"
)

func TestSetDefaults(t *testing.T) {
	// Reset viper before test
	viper.Reset()

	SetDefaults()

	tests := []struct {
		key      string
		expected any
	}{
		{"APP_NAME", "example-service"},
		{"APP_SERVER_PORT", "80"},
		{"LOGGER", "json"},
		{"LOGGER_LEVEL", "debug"},
		{"TRACING_ENABLED", false},
		{"SERVICE_NAME", "example-service"},
		{"PROTOFLOW_POISON_QUEUE", "messages-poison"},
		{"PROTOFLOW_WEBUI_ENABLED", true},
		{"PROTOFLOW_WEBUI_PORT", 8081},
		{"EVENTS_DEMO_CONSUME_QUEUE", "messages"},
		{"EVENTS_DEMO_PUBLISH_QUEUE", "messages-processed"},
		{"EVENTS_EXAMPLE_CONSUME_QUEUE", "example-records"},
		{"EVENTS_EXAMPLE_PUBLISH_QUEUE", "example-records-processed"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := viper.Get(tt.key)
			if got != tt.expected {
				t.Errorf("Default %s = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestSetDefaultsCORSValues(t *testing.T) {
	viper.Reset()
	SetDefaults()

	corsHeaders := viper.GetStringSlice("APP_SERVER_CORS_HEADERS")
	if len(corsHeaders) != 1 || corsHeaders[0] != "*" {
		t.Errorf("CORS headers = %v, want [*]", corsHeaders)
	}

	corsMethods := viper.GetStringSlice("APP_SERVER_CORS_METHODS")
	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	if len(corsMethods) != len(expectedMethods) {
		t.Errorf("CORS methods = %v, want %v", corsMethods, expectedMethods)
	}

	corsOrigins := viper.GetStringSlice("APP_SERVER_CORS_ORIGINS")
	if len(corsOrigins) != 1 || corsOrigins[0] != "*" {
		t.Errorf("CORS origins = %v, want [*]", corsOrigins)
	}
}

func testConfigNotNil(t *testing.T, cfg *Config) {
	t.Helper()
	if cfg.Router == nil {
		t.Error("Router config is nil")
	}
	if cfg.Server == nil {
		t.Error("Server config is nil")
	}
	if cfg.Database == nil {
		t.Error("Database config is nil")
	}
	if cfg.Logger == nil {
		t.Error("Logger config is nil")
	}
	if cfg.Tracing == nil {
		t.Error("Tracing config is nil")
	}
	if cfg.Metrics == nil {
		t.Error("Metrics config is nil")
	}
	if cfg.Protoflow == nil {
		t.Error("Protoflow config is nil")
	}
	if cfg.Events == nil {
		t.Error("Events config is nil")
	}
}

func TestLoadConfig(t *testing.T) {
	viper.Reset()

	cfg, err := LoadConfig("1.0.0", "2024-01-01", "test", "abc123", "2024-01-01")

	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig returned nil")
	}

	t.Run("Info", func(t *testing.T) {
		if cfg.Info == nil {
			t.Fatal("Info is nil")
		}
		if cfg.Info.Version != "1.0.0" {
			t.Errorf("Info.Version = %s, want 1.0.0", cfg.Info.Version)
		}
		if cfg.Info.BuildDate != "2024-01-01" {
			t.Errorf("Info.BuildDate = %s, want 2024-01-01", cfg.Info.BuildDate)
		}
		if cfg.Info.Details != "test" {
			t.Errorf("Info.Details = %s, want test", cfg.Info.Details)
		}
		if cfg.Info.CommitHash != "abc123" {
			t.Errorf("Info.CommitHash = %s, want abc123", cfg.Info.CommitHash)
		}
	})

	t.Run("ConfigsNotNil", func(t *testing.T) {
		testConfigNotNil(t, cfg)
	})
}

func TestLoadConfigWithEnvOverrides(t *testing.T) {
	viper.Reset()

	// Set some env overrides
	t.Setenv("APP_SERVER_PORT", "3000")
	t.Setenv("LOGGER", "pretty")
	t.Setenv("LOGGER_LEVEL", "info")

	cfg, err := LoadConfig("2.0.0", "2024-06-01", "prod", "def456", "2024-06-01")

	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	if cfg.Server.Address != "0.0.0.0:3000" {
		t.Errorf("Server.Address = %s, want 0.0.0.0:3000", cfg.Server.Address)
	}
}

func TestLoadLoggerConfigOTel(t *testing.T) {
	viper.Reset()
	t.Setenv("LOGGER", "otel")

	cfg, _ := LoadConfig("1.0.0", "", "", "", "")

	if cfg.Logger.ConsoleEnabled {
		t.Error("Console should be disabled for otel logger")
	}
	if !cfg.Logger.OTel.Enabled {
		t.Error("OTel should be enabled for otel logger")
	}
}

func TestLoadLoggerConfigOTelAndConsole(t *testing.T) {
	viper.Reset()
	t.Setenv("LOGGER", "otel-and-console")

	cfg, _ := LoadConfig("1.0.0", "", "", "", "")

	if !cfg.Logger.ConsoleEnabled {
		t.Error("Console should be enabled for otel-and-console logger")
	}
	if !cfg.Logger.OTel.Enabled {
		t.Error("OTel should be enabled for otel-and-console logger")
	}
	if !cfg.Logger.OTel.MirrorToConsole {
		t.Error("MirrorToConsole should be true for otel-and-console logger")
	}
}

func TestConfigStructFields(t *testing.T) {
	cfg := Config{}

	// Test that all fields are nil initially
	if cfg.Info != nil {
		t.Error("Info should be nil by default")
	}
	if cfg.Router != nil {
		t.Error("Router should be nil by default")
	}
	if cfg.Server != nil {
		t.Error("Server should be nil by default")
	}
}
