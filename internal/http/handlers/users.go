package handlers

import (
	"context"
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"microservice-template/internal/http/formatter"
	"microservice-template/internal/http/models"
	"microservice-template/internal/http/server/operations/users"
	"microservice-template/internal/service"
	"microservice-template/pkg/logger"
)

// NewGetUserByEmail creates new GetUserByEmail handler.
func NewGetUserByEmail(svc service.IService) *GetUserByEmail {
	return &GetUserByEmail{service: svc}
}

// GetUserByEmail handler retrieves a user by email address.
type GetUserByEmail struct {
	service service.IService
}

// Handle GetUserByEmail endpoint.
func (h *GetUserByEmail) Handle(params users.GetUserByEmailParams, _ *models.User) middleware.Responder {
	// Extract email from query parameter (convert from strfmt.Email to string)
	email := string(params.Email)

	// Validate email not empty (should be caught by swagger validation, but double-check)
	if email == "" {
		logger.Log().Error("email parameter is empty")
		return users.NewGetUserByEmailBadRequest().
			WithPayload(DefaultError(http.StatusBadRequest, service.ErrInvalidInput, nil))
	}

	// Call service layer
	ctx := context.Background()
	user, err := h.service.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Log().Errorf("get user by email %s: %s", email, err.Error())

		// Map service errors to HTTP status codes
		switch {
		case service.IsInvalidInput(err):
			return users.NewGetUserByEmailBadRequest().
				WithPayload(DefaultError(http.StatusBadRequest, err, nil))
		case service.IsNotFound(err):
			return users.NewGetUserByEmailNotFound().
				WithPayload(DefaultError(http.StatusNotFound, err, nil))
		case service.IsRepositoryUnavailable(err):
			return users.NewGetUserByEmailServiceUnavailable().
				WithPayload(DefaultError(http.StatusServiceUnavailable, err, nil))
		default:
			return users.NewGetUserByEmailInternalServerError().
				WithPayload(DefaultError(http.StatusInternalServerError, err, nil))
		}
	}

	// Convert domain model to API model and return
	return users.NewGetUserByEmailOK().WithPayload(formatter.UserToAPI(user))
}
