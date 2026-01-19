package models

import (
	"errors"
	"testing"
)

func TestUser_Validate_Valid(t *testing.T) {
	user := &User{
		Email:  "test@example.com",
		Name:   "John Doe",
		Status: UserActive,
	}

	if err := user.Validate(); err != nil {
		t.Fatalf("expected valid user, got error: %v", err)
	}
}

func TestUser_Validate_MissingEmail(t *testing.T) {
	user := &User{
		Name:   "John Doe",
		Status: UserActive,
	}

	err := user.Validate()
	assertValidationError(t, err, "email", "is required")
}

func TestUser_Validate_InvalidEmail(t *testing.T) {
	user := &User{
		Email:  "not-an-email",
		Name:   "John Doe",
		Status: UserActive,
	}

	err := user.Validate()
	assertValidationError(t, err, "email", "invalid format")
}

func TestUser_Validate_MissingName(t *testing.T) {
	user := &User{
		Email:  "test@example.com",
		Status: UserActive,
	}

	err := user.Validate()
	assertValidationError(t, err, "name", "is required")
}

func TestUser_Validate_InvalidStatus(t *testing.T) {
	user := &User{
		Email:  "test@example.com",
		Name:   "John Doe",
		Status: userStatusUnsupported,
	}

	err := user.Validate()
	assertValidationError(t, err, "status", "invalid value")
}

func assertValidationError(t *testing.T, err error, field, message string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var verr *ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if verr.Field != field {
		t.Fatalf("Field = %q, want %q", verr.Field, field)
	}
	if verr.Message != message {
		t.Fatalf("Message = %q, want %q", verr.Message, message)
	}
}
