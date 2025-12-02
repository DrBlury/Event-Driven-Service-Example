package app

import (
	"os"
	"testing"
	"time"
)

func TestSetDefaults(t *testing.T) {
	// Just ensure it doesn't panic
	SetDefaults()
}

func TestLoadConfig(t *testing.T) {
	// Set some env vars for testing
	os.Setenv("APP_NAME", "test-service")
	os.Setenv("LOGGER", "json")
	os.Setenv("LOGGER_LEVEL", "info")
	defer func() {
		os.Unsetenv("APP_NAME")
		os.Unsetenv("LOGGER")
		os.Unsetenv("LOGGER_LEVEL")
	}()

	cfg, err := LoadConfig(
		"1.0.0",
		"2024-01-01",
		"Test details",
		"abc123",
		"2024-01-01T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	testLoadConfigInfo(t, cfg)
	testLoadConfigNonNil(t, cfg)
}

func testLoadConfigInfo(t *testing.T, cfg *Config) {
	t.Helper()

	if cfg.Info == nil {
		t.Fatal("Info config is nil")
	}
	if cfg.Info.Version != "1.0.0" {
		t.Errorf("Info.Version = %q, want '1.0.0'", cfg.Info.Version)
	}
	if cfg.Info.BuildDate != "2024-01-01" {
		t.Errorf("Info.BuildDate = %q, want '2024-01-01'", cfg.Info.BuildDate)
	}
	if cfg.Info.Details != "Test details" {
		t.Errorf("Info.Details = %q, want 'Test details'", cfg.Info.Details)
	}
	if cfg.Info.CommitHash != "abc123" {
		t.Errorf("Info.CommitHash = %q, want 'abc123'", cfg.Info.CommitHash)
	}
	if cfg.Info.CommitDate != "2024-01-01T00:00:00Z" {
		t.Errorf("Info.CommitDate = %q, want '2024-01-01T00:00:00Z'", cfg.Info.CommitDate)
	}
}

func testLoadConfigNonNil(t *testing.T, cfg *Config) {
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

func TestLoadConfigWithOTelLogger(t *testing.T) {
	os.Setenv("LOGGER", "otel")
	defer os.Unsetenv("LOGGER")

	cfg, err := LoadConfig("1.0.0", "2024-01-01", "Test", "abc123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Logger.ConsoleEnabled {
		t.Error("Logger.ConsoleEnabled should be false for otel")
	}
	if !cfg.Logger.OTel.Enabled {
		t.Error("Logger.OTel.Enabled should be true for otel")
	}
}

func TestLoadConfigWithOTelAndConsoleLogger(t *testing.T) {
	os.Setenv("LOGGER", "otel-and-console")
	defer os.Unsetenv("LOGGER")

	cfg, err := LoadConfig("1.0.0", "2024-01-01", "Test", "abc123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if !cfg.Logger.ConsoleEnabled {
		t.Error("Logger.ConsoleEnabled should be true for otel-and-console")
	}
	if !cfg.Logger.OTel.Enabled {
		t.Error("Logger.OTel.Enabled should be true for otel-and-console")
	}
	if !cfg.Logger.OTel.MirrorToConsole {
		t.Error("Logger.OTel.MirrorToConsole should be true for otel-and-console")
	}
}

func TestLoadConfigServerDefaults(t *testing.T) {
	SetDefaults()

	cfg, err := LoadConfig("1.0.0", "2024-01-01", "Test", "abc123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Server.Address != "0.0.0.0:80" {
		t.Errorf("Server.Address = %q, want '0.0.0.0:80'", cfg.Server.Address)
	}
}

func TestLoadConfigRouterDefaults(t *testing.T) {
	SetDefaults()

	cfg, err := LoadConfig("1.0.0", "2024-01-01", "Test", "abc123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Router.Timeout != 60*time.Second {
		t.Errorf("Router.Timeout = %v, want 60s", cfg.Router.Timeout)
	}
}

func TestLoadConfigEventsDefaults(t *testing.T) {
	SetDefaults()

	cfg, err := LoadConfig("1.0.0", "2024-01-01", "Test", "abc123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Events.DemoConsumeQueue != "messages" {
		t.Errorf("Events.DemoConsumeQueue = %q, want 'messages'", cfg.Events.DemoConsumeQueue)
	}
	if cfg.Events.ExampleConsumeQueue != "example-records" {
		t.Errorf("Events.ExampleConsumeQueue = %q, want 'example-records'", cfg.Events.ExampleConsumeQueue)
	}
}

func TestLoadConfigProtoflowDefaults(t *testing.T) {
	SetDefaults()

	cfg, err := LoadConfig("1.0.0", "2024-01-01", "Test", "abc123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Protoflow.PoisonQueue != "messages-poison" {
		t.Errorf("Protoflow.PoisonQueue = %q, want 'messages-poison'", cfg.Protoflow.PoisonQueue)
	}
	if !cfg.Protoflow.WebUIEnabled {
		t.Error("Protoflow.WebUIEnabled should be true by default")
	}
	if cfg.Protoflow.WebUIPort != 8081 {
		t.Errorf("Protoflow.WebUIPort = %d, want 8081", cfg.Protoflow.WebUIPort)
	}
}

func TestLoadConfigTracingDefaults(t *testing.T) {
	SetDefaults()

	cfg, err := LoadConfig("1.0.0", "2024-01-01", "Test", "abc123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Tracing.Enabled {
		t.Error("Tracing.Enabled should be false by default")
	}
}

func TestLoadConfigMetricsDefaults(t *testing.T) {
	SetDefaults()

	cfg, err := LoadConfig("1.0.0", "2024-01-01", "Test", "abc123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	// Metrics should have defaults set
	if cfg.Metrics == nil {
		t.Fatal("Metrics config is nil")
	}
}

func TestLoadConfigDatabaseConfig(t *testing.T) {
	os.Setenv("MONGO_URL", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "testdb")
	os.Setenv("MONGO_USER", "testuser")
	os.Setenv("MONGO_PASSWORD", "testpass")
	defer func() {
		os.Unsetenv("MONGO_URL")
		os.Unsetenv("MONGO_DB")
		os.Unsetenv("MONGO_USER")
		os.Unsetenv("MONGO_PASSWORD")
	}()

	cfg, err := LoadConfig("1.0.0", "2024-01-01", "Test", "abc123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Database.MongoURL != "mongodb://localhost:27017" {
		t.Errorf("Database.MongoURL = %q, want 'mongodb://localhost:27017'", cfg.Database.MongoURL)
	}
	if cfg.Database.MongoDB != "testdb" {
		t.Errorf("Database.MongoDB = %q, want 'testdb'", cfg.Database.MongoDB)
	}
}

func TestLoadConfigWithCustomServerPort(t *testing.T) {
	os.Setenv("APP_SERVER_PORT", "8080")
	defer os.Unsetenv("APP_SERVER_PORT")

	cfg, err := LoadConfig("1.0.0", "2024-01-01", "Test", "abc123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Server.Address != "0.0.0.0:8080" {
		t.Errorf("Server.Address = %q, want '0.0.0.0:8080'", cfg.Server.Address)
	}
}

func TestLoadConfigWithTracingEnabled(t *testing.T) {
	os.Setenv("TRACING_ENABLED", "true")
	defer os.Unsetenv("TRACING_ENABLED")

	cfg, err := LoadConfig("1.0.0", "2024-01-01", "Test", "abc123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if !cfg.Tracing.Enabled {
		t.Error("Tracing.Enabled should be true when TRACING_ENABLED=true")
	}
}

func TestLoadConfigWithMetricsEnabled(t *testing.T) {
	os.Setenv("METRICS_ENABLED", "true")
	defer os.Unsetenv("METRICS_ENABLED")

	cfg, err := LoadConfig("1.0.0", "2024-01-01", "Test", "abc123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if !cfg.Metrics.Enabled {
		t.Error("Metrics.Enabled should be true when METRICS_ENABLED=true")
	}
}

func TestConfigStructFields(t *testing.T) {
	cfg := &Config{}
	if cfg.Info != nil {
		t.Error("Config.Info should be nil by default")
	}
	if cfg.Router != nil {
		t.Error("Config.Router should be nil by default")
	}
	if cfg.Server != nil {
		t.Error("Config.Server should be nil by default")
	}
}
