package service

import (
	"errors"
)

// Sentinel errors for common business logic failures.
// These errors are used by the HTTP layer to map to appropriate HTTP status codes.
var (
	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrInvalidInput indicates the input validation failed.
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized indicates authentication failed.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates insufficient permissions.
	ErrForbidden = errors.New("forbidden")

	// ErrConflict indicates resource already exists.
	ErrConflict = errors.New("conflict")

	// ErrRepositoryUnavailable indicates database module is not enabled.
	ErrRepositoryUnavailable = errors.New("repository unavailable")
)

// IsNotFound checks if error is a not found error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsInvalidInput checks if error is an invalid input error.
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorized checks if error is an unauthorized error.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden checks if error is a forbidden error.
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsConflict checks if error is a conflict error.
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsRepositoryUnavailable checks if error is a repository unavailable error.
func IsRepositoryUnavailable(err error) bool {
	return errors.Is(err, ErrRepositoryUnavailable)
}
