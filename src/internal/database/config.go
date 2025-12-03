package database

import (
	"errors"
	"strings"
)

// Config holds MongoDB connection parameters.
// SECURITY NOTE: Credentials should be provided via environment variables
// or a secrets manager, never hardcoded or committed to version control.
type Config struct {
	MongoURL      string
	MongoDB       string
	MongoUser     string
	MongoPassword string
}

// Validate checks that the configuration has the required fields set.
// Returns an error if essential connection parameters are missing.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("database config is nil")
	}
	if strings.TrimSpace(c.MongoURL) == "" {
		return errors.New("MONGO_URL is required")
	}
	if strings.TrimSpace(c.MongoDB) == "" {
		return errors.New("MONGO_DB is required")
	}
	// User and password are optional for local development but recommended
	return nil
}

// Redacted returns a copy of the config with sensitive fields masked.
// Use this for logging to avoid exposing credentials.
func (c *Config) Redacted() Config {
	if c == nil {
		return Config{}
	}
	redacted := *c
	if redacted.MongoPassword != "" {
		redacted.MongoPassword = "[REDACTED]"
	}
	return redacted
}
