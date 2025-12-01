package database

import (
	"context"
	"testing"
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
		err := db.Ping(nil)
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
