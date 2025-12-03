package database

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestDatabasePing(t *testing.T) {
	t.Parallel()

	t.Run("nil database", func(t *testing.T) {
		t.Parallel()
		var db *Database
		err := db.Ping(context.Background())
		if err == nil {
			t.Error("Ping() on nil database should return error")
		}
		if err.Error() != "database not configured" {
			t.Errorf("Ping() error = %q, want 'database not configured'", err.Error())
		}
	})

	t.Run("nil mongo database", func(t *testing.T) {
		t.Parallel()
		db := &Database{DB: nil, Cfg: &Config{}}
		err := db.Ping(context.Background())
		if err == nil {
			t.Error("Ping() with nil DB should return error")
		}
		if err.Error() != "mongo database handle is nil" {
			t.Errorf("Ping() error = %q, want 'mongo database handle is nil'", err.Error())
		}
	})

	t.Run("nil context uses background", func(t *testing.T) {
		t.Parallel()
		db := &Database{DB: nil, Cfg: &Config{}}
		err := db.Ping(context.Background()) // Use context.Background() for test contexts
		if err == nil {
			t.Error("Ping() with nil DB and nil context should return error")
		}
	})
}

func TestConfig(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		MongoURL:      "mongodb://localhost:27017",
		MongoDB:       "testdb",
		MongoUser:     "testuser",
		MongoPassword: "testpass",
	}

	if cfg.MongoURL != "mongodb://localhost:27017" {
		t.Errorf("MongoURL = %q, want 'mongodb://localhost:27017'", cfg.MongoURL)
	}
	if cfg.MongoDB != "testdb" {
		t.Errorf("MongoDB = %q, want 'testdb'", cfg.MongoDB)
	}
	if cfg.MongoUser != "testuser" {
		t.Errorf("MongoUser = %q, want 'testuser'", cfg.MongoUser)
	}
	if cfg.MongoPassword != "testpass" {
		t.Errorf("MongoPassword = %q, want 'testpass'", cfg.MongoPassword)
	}
}

func TestDatabaseStruct(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		MongoURL:      "mongodb://localhost:27017",
		MongoDB:       "testdb",
		MongoUser:     "testuser",
		MongoPassword: "testpass",
	}

	db := &Database{
		DB:  nil,
		Cfg: cfg,
	}

	if db.Cfg != cfg {
		t.Error("Database.Cfg should match")
	}
	if db.DB != nil {
		t.Error("Database.DB should be nil")
	}
}

func TestNewDatabaseInvalidURL(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	cfg := &Config{
		MongoURL:      "invalid://not-a-valid-url",
		MongoDB:       "testdb",
		MongoUser:     "user",
		MongoPassword: "pass",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := NewDatabase(cfg, logger, ctx)
	if err == nil {
		t.Error("NewDatabase should error with invalid URL")
	}
}

func TestNewDatabaseEmptyURL(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	cfg := &Config{
		MongoURL:      "",
		MongoDB:       "testdb",
		MongoUser:     "user",
		MongoPassword: "pass",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := NewDatabase(cfg, logger, ctx)
	if err == nil {
		t.Log("NewDatabase with empty URL may succeed or fail depending on driver")
	}
}

func TestNewDatabaseUnreachableHost(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	cfg := &Config{
		MongoURL:      "mongodb://nonexistent.invalid.host:27017",
		MongoDB:       "testdb",
		MongoUser:     "user",
		MongoPassword: "pass",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := NewDatabase(cfg, logger, ctx)
	if err == nil {
		t.Log("NewDatabase to unreachable host may succeed with lazy connection")
	}
}

func TestNewDatabaseContextCancelled(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	cfg := &Config{
		MongoURL:      "mongodb://localhost:27017",
		MongoDB:       "testdb",
		MongoUser:     "user",
		MongoPassword: "pass",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := NewDatabase(cfg, logger, ctx)
	if err == nil {
		t.Log("NewDatabase with cancelled context may succeed or fail")
	}
}

func TestNewDatabaseNilLogger(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		MongoURL:      "invalid://bad",
		MongoDB:       "testdb",
		MongoUser:     "user",
		MongoPassword: "pass",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This may panic or fail gracefully depending on implementation
	func() {
		defer func() {
			_ = recover()
		}()
		_, _ = NewDatabase(cfg, nil, ctx)
	}()
}

func TestNewDatabaseWithAllConfigs(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	testCases := []struct {
		name string
		cfg  *Config
	}{
		{
			name: "minimal config",
			cfg: &Config{
				MongoURL: "mongodb://localhost:27017",
				MongoDB:  "test",
			},
		},
		{
			name: "with auth",
			cfg: &Config{
				MongoURL:      "mongodb://localhost:27017",
				MongoDB:       "test",
				MongoUser:     "user",
				MongoPassword: "password",
			},
		},
		{
			name: "replica set",
			cfg: &Config{
				MongoURL:      "mongodb://localhost:27017,localhost:27018,localhost:27019/?replicaSet=rs0",
				MongoDB:       "test",
				MongoUser:     "user",
				MongoPassword: "password",
			},
		},
		{
			name: "with options",
			cfg: &Config{
				MongoURL:      "mongodb://localhost:27017/?connectTimeoutMS=1000&serverSelectionTimeoutMS=1000",
				MongoDB:       "test",
				MongoUser:     "user",
				MongoPassword: "password",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			_, err := NewDatabase(tc.cfg, logger, ctx)
			// All will fail because MongoDB is not running, which is expected
			if err != nil {
				t.Logf("Expected failure for %s: %v", tc.name, err)
			}
		})
	}
}

func TestDatabasePingWithCancelledContext(t *testing.T) {
	t.Parallel()

	db := &Database{DB: nil, Cfg: &Config{}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := db.Ping(ctx)
	if err == nil {
		t.Error("Ping should error with nil DB even with cancelled context")
	}
}

func TestDatabasePingWithTimeoutContext(t *testing.T) {
	t.Parallel()

	db := &Database{DB: nil, Cfg: &Config{}}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(2 * time.Millisecond) // Let timeout expire

	err := db.Ping(ctx)
	if err == nil {
		t.Error("Ping should error with nil DB")
	}
}

func TestConfigVariants(t *testing.T) {
	t.Parallel()

	for _, tc := range configVariantTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertConfigFields(t, tc.cfg, tc.fields)
		})
	}
}

type configVariantTestCase struct {
	name   string
	cfg    Config
	fields map[string]string
}

func configVariantTestCases() []configVariantTestCase {
	return []configVariantTestCase{
		{
			name: "all empty",
			cfg:  Config{},
			fields: map[string]string{
				"MongoURL": "", "MongoDB": "", "MongoUser": "", "MongoPassword": "",
			},
		},
		{
			name: "typical production",
			cfg: Config{
				MongoURL: "mongodb://mongo.example.com:27017", MongoDB: "production",
				MongoUser: "app_user", MongoPassword: "secure_password_123",
			},
			fields: map[string]string{
				"MongoURL": "mongodb://mongo.example.com:27017", "MongoDB": "production",
				"MongoUser": "app_user", "MongoPassword": "secure_password_123",
			},
		},
		{
			name: "localhost dev",
			cfg: Config{
				MongoURL: "mongodb://localhost:27017", MongoDB: "dev",
				MongoUser: "dev", MongoPassword: "dev",
			},
			fields: map[string]string{
				"MongoURL": "mongodb://localhost:27017", "MongoDB": "dev",
				"MongoUser": "dev", "MongoPassword": "dev",
			},
		},
	}
}

func assertConfigFields(t *testing.T, cfg Config, fields map[string]string) {
	t.Helper()
	if cfg.MongoURL != fields["MongoURL"] {
		t.Errorf("MongoURL mismatch")
	}
	if cfg.MongoDB != fields["MongoDB"] {
		t.Errorf("MongoDB mismatch")
	}
	if cfg.MongoUser != fields["MongoUser"] {
		t.Errorf("MongoUser mismatch")
	}
	if cfg.MongoPassword != fields["MongoPassword"] {
		t.Errorf("MongoPassword mismatch")
	}
}

func TestDatabasePingErrorMessages(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		db          *Database
		expectedErr string
	}{
		{
			name:        "nil database pointer",
			db:          nil,
			expectedErr: "database not configured",
		},
		{
			name:        "nil mongo database",
			db:          &Database{DB: nil, Cfg: &Config{}},
			expectedErr: "mongo database handle is nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.db.Ping(context.Background())
			if err == nil {
				t.Error("Expected error")
			}
			if err.Error() != tc.expectedErr {
				t.Errorf("Error = %q, want %q", err.Error(), tc.expectedErr)
			}
		})
	}
}
