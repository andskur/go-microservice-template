package models

import "testing"

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{Field: "email", Message: "is required"}
	expected := "email: is required"
	if got := err.Error(); got != expected {
		t.Fatalf("Error() = %q, want %q", got, expected)
	}
}

func TestValidationError_AsError(t *testing.T) {
	var err error = &ValidationError{Field: "name", Message: "too short"}
	verr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError type, got %T", err)
	}
	if verr.Field != "name" {
		t.Fatalf("Field = %q, want %q", verr.Field, "name")
	}
	if verr.Message != "too short" {
		t.Fatalf("Message = %q, want %q", verr.Message, "too short")
	}
}
