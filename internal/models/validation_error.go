// Package models contains domain model definitions and helpers.
package models

import "fmt"

// ValidationError represents a field-level validation failure.
type ValidationError struct {
	Field   string // Name of the field that failed validation
	Message string // Human-readable error message
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// newValidationError creates a new ValidationError.
func newValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
