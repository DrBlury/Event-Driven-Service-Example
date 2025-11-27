package database

import "testing"

func TestConfigFields(t *testing.T) {
	cfg := Config{
		MongoURL:      "mongodb://localhost:27017",
		MongoDB:       "testdb",
		MongoUser:     "testuser",
		MongoPassword: "testpass",
	}

	if cfg.MongoURL != "mongodb://localhost:27017" {
		t.Errorf("MongoURL = %q, want %q", cfg.MongoURL, "mongodb://localhost:27017")
	}
	if cfg.MongoDB != "testdb" {
		t.Errorf("MongoDB = %q, want %q", cfg.MongoDB, "testdb")
	}
	if cfg.MongoUser != "testuser" {
		t.Errorf("MongoUser = %q, want %q", cfg.MongoUser, "testuser")
	}
	if cfg.MongoPassword != "testpass" {
		t.Errorf("MongoPassword = %q, want %q", cfg.MongoPassword, "testpass")
	}
}

func TestConfigZeroValue(t *testing.T) {
	var cfg Config

	if cfg.MongoURL != "" {
		t.Errorf("zero value MongoURL = %q, want empty string", cfg.MongoURL)
	}
	if cfg.MongoDB != "" {
		t.Errorf("zero value MongoDB = %q, want empty string", cfg.MongoDB)
	}
	if cfg.MongoUser != "" {
		t.Errorf("zero value MongoUser = %q, want empty string", cfg.MongoUser)
	}
	if cfg.MongoPassword != "" {
		t.Errorf("zero value MongoPassword = %q, want empty string", cfg.MongoPassword)
	}
}
