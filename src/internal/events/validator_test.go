package events

import (
	"testing"
)

func TestNewValidator(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}
	if v == nil {
		t.Fatal("NewValidator() returned nil")
	}
	if v.validator == nil {
		t.Error("validator field is nil")
	}
}

func TestNewValidatorMultipleCalls(t *testing.T) {
	// Ensure we can create multiple validators without issues
	for i := 0; i < 3; i++ {
		v, err := NewValidator()
		if err != nil {
			t.Fatalf("NewValidator() iteration %d error = %v", i, err)
		}
		if v == nil {
			t.Errorf("NewValidator() iteration %d returned nil", i)
		}
	}
}
