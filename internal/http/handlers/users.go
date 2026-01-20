package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-openapi/runtime/middleware"

	"microservice-template/internal/grpcclient"
	"microservice-template/internal/http/formatter"
	"microservice-template/internal/http/models"
	"microservice-template/internal/http/server/operations/users"
	"microservice-template/internal/service"
	"microservice-template/pkg/logger"
)

// NewGetUserByEmail creates new GetUserByEmail handler.
func NewGetUserByEmail(svc service.IService, grpcClient grpcclient.IClient) *GetUserByEmail {
	return &GetUserByEmail{
		service:    svc,
		grpcClient: grpcClient,
	}
}

// GetUserByEmail handler retrieves a user by email address.
type GetUserByEmail struct {
	service    service.IService
	grpcClient grpcclient.IClient
}

// Handle GetUserByEmail endpoint.
func (h *GetUserByEmail) Handle(params users.GetUserByEmailParams, _ *models.User) middleware.Responder {
	email := string(params.Email)

	if email == "" {
		logger.Log().Error("email parameter is empty")
		return users.NewGetUserByEmailBadRequest().
			WithPayload(DefaultError(http.StatusBadRequest, service.ErrInvalidInput, nil))
	}

	ctx := context.Background()

	// Fetch user from external service via gRPC client
	// NOTE: Currently fetching ONLY from external service.
	// Alternative patterns:
	// 1. Fetch from local DB: user, err := h.service.GetUserByEmail(ctx, email)
	// 2. Try external first, fallback to local if external fails
	// 3. Aggregate data from both external and local sources

	if h.grpcClient == nil {
		// gRPC client not configured/enabled
		logger.Log().Error("grpc client not available")
		return users.NewGetUserByEmailServiceUnavailable().
			WithPayload(DefaultError(http.StatusServiceUnavailable,
				fmt.Errorf("external user service not available"), nil))
	}

	user, err := h.grpcClient.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Log().Errorf("get user by email from external service %s: %s", email, err.Error())

		// Map errors to HTTP status codes
		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "invalid input"):
			return users.NewGetUserByEmailBadRequest().
				WithPayload(DefaultError(http.StatusBadRequest, err, nil))
		case strings.Contains(errStr, "not found"):
			return users.NewGetUserByEmailNotFound().
				WithPayload(DefaultError(http.StatusNotFound, err, nil))
		case strings.Contains(errStr, "unavailable"), strings.Contains(errStr, "timeout"):
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
