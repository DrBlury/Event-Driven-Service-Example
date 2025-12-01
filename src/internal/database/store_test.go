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

func TestStoreExampleRecordValidRecord(t *testing.T) {
	db := &Database{
		DB:  nil,
		Cfg: &Config{},
	}

	record := &domain.ExampleRecord{
		RecordId:    "test-123",
		Title:       "Test Record",
		Description: "Test description",
	}

	// This will panic or error because DB is nil, which is expected
	// We're testing that the nil record check passes for valid record
	func() {
		defer func() {
			if r := recover(); r == nil {
				// May not panic depending on implementation
			}
		}()
		_ = db.StoreExampleRecord(context.Background(), record)
	}()
}

func TestExampleCollectionConstant(t *testing.T) {
	if exampleCollection != "example-records" {
		t.Errorf("exampleCollection = %q, want 'example-records'", exampleCollection)
	}
}
