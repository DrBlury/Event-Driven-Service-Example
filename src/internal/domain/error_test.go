package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrorNotFound", ErrorNotFound, "not found"},
		{"ErrorBadRequest", ErrorBadRequest, "something you provided was wrong"},
		{"ErrorUpstreamService", ErrorUpstreamService, "upstream service error"},
		{"ErrorNotImplemented", ErrorNotImplemented, "not implemented"},
		{"ErrorInternal", ErrorInternal, "internal error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("got %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestErrValidationsError(t *testing.T) {
	tests := []struct {
		name   string
		errors []string
		want   string
	}{
		{
			name:   "single error",
			errors: []string{"field is required"},
			want:   "1: field is required",
		},
		{
			name:   "multiple errors",
			errors: []string{"field1 is required", "field2 must be positive"},
			want:   "1: field1 is required - 2: field2 must be positive",
		},
		{
			name:   "empty errors",
			errors: []string{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errsCopy := make([]string, len(tt.errors))
			copy(errsCopy, tt.errors)

			e := ErrValidations{Errors: errsCopy}
			got := e.Error()
			if got != tt.want {
				t.Errorf("ErrValidations.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrValidationsImplementsError(t *testing.T) {
	var _ error = ErrValidations{Errors: []string{"test"}}
	// If this compiles, ErrValidations implements error interface
}

func TestErrorsIs(t *testing.T) {
	wrappedNotFound := errors.Join(errors.New("wrapper"), ErrorNotFound)

	if !errors.Is(wrappedNotFound, ErrorNotFound) {
		t.Error("errors.Is should find ErrorNotFound in wrapped error")
	}

	if errors.Is(wrappedNotFound, ErrorBadRequest) {
		t.Error("errors.Is should not find ErrorBadRequest in wrapped error")
	}
}

func TestErrValidationsMultipleErrors(t *testing.T) {
	e := ErrValidations{
		Errors: []string{
			"name is required",
			"age must be positive",
			"email format invalid",
		},
	}

	result := e.Error()

	if !strings.Contains(result, "1:") {
		t.Error("Expected error to contain '1:'")
	}
	if !strings.Contains(result, "2:") {
		t.Error("Expected error to contain '2:'")
	}
	if !strings.Contains(result, "3:") {
		t.Error("Expected error to contain '3:'")
	}

	if !strings.Contains(result, " - ") {
		t.Error("Expected errors to be separated by ' - '")
	}
}

func TestErrValidationsAsError(t *testing.T) {
	originalErr := ErrValidations{Errors: []string{"test error"}}

	// Wrap it
	wrappedErr := errors.Join(errors.New("context"), originalErr)

	var validationErr ErrValidations
	if !errors.As(wrappedErr, &validationErr) {
		t.Error("errors.As should find ErrValidations")
	}

	if len(validationErr.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(validationErr.Errors))
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	// Verify each error is distinct
	errs := []error{
		ErrorNotFound,
		ErrorBadRequest,
		ErrorUpstreamService,
		ErrorNotImplemented,
		ErrorInternal,
	}

	for i, err1 := range errs {
		for j, err2 := range errs {
			if i != j && errors.Is(err1, err2) {
				t.Errorf("Errors should be distinct: %v and %v", err1, err2)
			}
		}
	}
}

func TestErrValidationsWithSpecialCharacters(t *testing.T) {
	e := ErrValidations{
		Errors: []string{
			"field 'name' is required",
			`value must not contain "quotes"`,
			"special: chars & symbols < > \" '",
		},
	}

	result := e.Error()

	if !strings.Contains(result, "'name'") {
		t.Error("Should preserve single quotes")
	}
	if !strings.Contains(result, `"quotes"`) {
		t.Error("Should preserve double quotes")
	}
}
