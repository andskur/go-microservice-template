package handlers

import (
	"microservice-template/internal/http/models"
)

// DefaultError creates a default error response
func DefaultError(code int, err error, details interface{}) *models.Error {
	message := err.Error()
	code64 := int64(code)

	return &models.Error{
		Code:    &code64,
		Message: &message,
		Details: details,
	}
}
