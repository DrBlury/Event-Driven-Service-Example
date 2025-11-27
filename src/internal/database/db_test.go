package database

import (
	"context"
	"testing"
)

func TestDatabasePingNilReceiver(t *testing.T) {
	var db *Database

	err := db.Ping(context.Background())
	if err == nil {
		t.Error("Expected error for nil receiver")
	}
	if err.Error() != "database not configured" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestDatabasePingNilDB(t *testing.T) {
	db := &Database{
		DB:  nil,
		Cfg: &Config{},
	}

	err := db.Ping(context.Background())
	if err == nil {
		t.Error("Expected error for nil DB")
	}
	if err.Error() != "mongo database handle is nil" {
		t.Errorf("Unexpected error: %v", err)
	}
}
