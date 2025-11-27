package database

import (
	"context"
	"testing"

	"drblury/event-driven-service/internal/domain"
)

func TestStoreExampleRecordNilRecord(t *testing.T) {
	db := &Database{
		DB:  nil,
		Cfg: &Config{},
	}

	err := db.StoreExampleRecord(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil record")
	}
	if err.Error() != "example record is required" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestStoreExampleRecordRequiresRecord(t *testing.T) {
	db := &Database{
		DB:  nil,
		Cfg: &Config{},
	}

	// Test that nil record check happens before DB access
	record := (*domain.ExampleRecord)(nil)
	err := db.StoreExampleRecord(context.Background(), record)
	if err == nil {
		t.Error("Expected error for nil record pointer")
	}
}
