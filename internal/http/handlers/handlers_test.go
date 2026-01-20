package handlers

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/gofrs/uuid"

	"microservice-template/internal/http/models"
	"microservice-template/internal/http/server/operations/users"
	domainmodels "microservice-template/internal/models"
	"microservice-template/internal/service"
)

// mockService is a mock implementation of service.IService for testing.
type mockService struct {
	getUserByEmailFunc func(ctx context.Context, email string) (*domainmodels.User, error)
	createUserFunc     func(ctx context.Context, user *domainmodels.User) error
}

func (m *mockService) GetUserByEmail(ctx context.Context, email string) (*domainmodels.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) CreateUser(ctx context.Context, user *domainmodels.User) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, user)
	}
	return errors.New("not implemented")
}

func TestNewGetUserByEmail(t *testing.T) {
	svc := &mockService{}
	handler := NewGetUserByEmail(svc)

	if handler == nil {
		t.Fatal("NewGetUserByEmail returned nil")
	}

	if handler.service != svc {
		t.Error("handler service not set correctly")
	}
}

func TestGetUserByEmail_Success(t *testing.T) {
	// Setup mock service
	userUUID := uuid.Must(uuid.NewV4())
	expectedUser := &domainmodels.User{
		UUID:   userUUID,
		Email:  "test@example.com",
		Name:   "Test User",
		Status: domainmodels.UserActive,
	}

	svc := &mockService{
		getUserByEmailFunc: func(_ context.Context, email string) (*domainmodels.User, error) {
			if email != "test@example.com" {
				t.Errorf("unexpected email: %s", email)
			}
			return expectedUser, nil
		},
	}

	handler := NewGetUserByEmail(svc)

	// Create params
	email := strfmt.Email("test@example.com")
	params := users.GetUserByEmailParams{
		Email: email,
	}

	// Create principal (mock authenticated user)
	principalEmail := strfmt.Email("admin@example.com")
	principalName := "Admin"
	principalStatus := "active"
	principal := &models.User{
		UUID:   strfmt.UUID(uuid.Must(uuid.NewV4()).String()),
		Email:  &principalEmail,
		Name:   &principalName,
		Status: &principalStatus,
	}

	// Execute handler
	responder := handler.Handle(params, principal)

	// Check response type
	if responder == nil {
		t.Fatal("responder is nil")
	}

	// Type assert to success response
	okResponse, ok := responder.(*users.GetUserByEmailOK)
	if !ok {
		t.Fatalf("expected *users.GetUserByEmailOK, got %T", responder)
	}

	// Verify payload
	if okResponse.Payload == nil {
		t.Fatal("response payload is nil")
	}

	if string(okResponse.Payload.UUID) != userUUID.String() {
		t.Errorf("UUID mismatch: expected %s, got %s", userUUID.String(), okResponse.Payload.UUID)
	}

	if okResponse.Payload.Email == nil || string(*okResponse.Payload.Email) != "test@example.com" {
		t.Error("email mismatch in response")
	}
}

func TestGetUserByEmail_EmptyEmail(t *testing.T) {
	svc := &mockService{
		getUserByEmailFunc: func(_ context.Context, _ string) (*domainmodels.User, error) {
			t.Error("service should not be called with empty email")
			return nil, nil
		},
	}

	handler := NewGetUserByEmail(svc)

	// Create params with empty email
	params := users.GetUserByEmailParams{
		Email: strfmt.Email(""),
	}

	principal := &models.User{}
	responder := handler.Handle(params, principal)

	// Should return bad request
	badRequestResponse, ok := responder.(*users.GetUserByEmailBadRequest)
	if !ok {
		t.Fatalf("expected *users.GetUserByEmailBadRequest, got %T", responder)
	}

	if badRequestResponse.Payload == nil {
		t.Fatal("error payload is nil")
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	svc := &mockService{
		getUserByEmailFunc: func(_ context.Context, _ string) (*domainmodels.User, error) {
			return nil, service.ErrNotFound
		},
	}

	handler := NewGetUserByEmail(svc)

	email := strfmt.Email("notfound@example.com")
	params := users.GetUserByEmailParams{
		Email: email,
	}

	principal := &models.User{}
	responder := handler.Handle(params, principal)

	// Should return not found
	notFoundResponse, ok := responder.(*users.GetUserByEmailNotFound)
	if !ok {
		t.Fatalf("expected *users.GetUserByEmailNotFound, got %T", responder)
	}

	if notFoundResponse.Payload == nil {
		t.Fatal("error payload is nil")
	}
}

func TestGetUserByEmail_InvalidInput(t *testing.T) {
	svc := &mockService{
		getUserByEmailFunc: func(_ context.Context, _ string) (*domainmodels.User, error) {
			return nil, service.ErrInvalidInput
		},
	}

	handler := NewGetUserByEmail(svc)

	email := strfmt.Email("invalid")
	params := users.GetUserByEmailParams{
		Email: email,
	}

	principal := &models.User{}
	responder := handler.Handle(params, principal)

	// Should return bad request
	badRequestResponse, ok := responder.(*users.GetUserByEmailBadRequest)
	if !ok {
		t.Fatalf("expected *users.GetUserByEmailBadRequest, got %T", responder)
	}

	if badRequestResponse.Payload == nil {
		t.Fatal("error payload is nil")
	}
}

func TestGetUserByEmail_RepositoryUnavailable(t *testing.T) {
	svc := &mockService{
		getUserByEmailFunc: func(_ context.Context, _ string) (*domainmodels.User, error) {
			return nil, service.ErrRepositoryUnavailable
		},
	}

	handler := NewGetUserByEmail(svc)

	email := strfmt.Email("test@example.com")
	params := users.GetUserByEmailParams{
		Email: email,
	}

	principal := &models.User{}
	responder := handler.Handle(params, principal)

	// Should return service unavailable
	unavailableResponse, ok := responder.(*users.GetUserByEmailServiceUnavailable)
	if !ok {
		t.Fatalf("expected *users.GetUserByEmailServiceUnavailable, got %T", responder)
	}

	if unavailableResponse.Payload == nil {
		t.Fatal("error payload is nil")
	}
}

func TestGetUserByEmail_InternalError(t *testing.T) {
	svc := &mockService{
		getUserByEmailFunc: func(_ context.Context, _ string) (*domainmodels.User, error) {
			return nil, fmt.Errorf("unexpected error")
		},
	}

	handler := NewGetUserByEmail(svc)

	email := strfmt.Email("test@example.com")
	params := users.GetUserByEmailParams{
		Email: email,
	}

	principal := &models.User{}
	responder := handler.Handle(params, principal)

	// Should return internal server error
	errorResponse, ok := responder.(*users.GetUserByEmailInternalServerError)
	if !ok {
		t.Fatalf("expected *users.GetUserByEmailInternalServerError, got %T", responder)
	}

	if errorResponse.Payload == nil {
		t.Fatal("error payload is nil")
	}
}

func TestDefaultError(t *testing.T) {
	testError := errors.New("test error message")
	code := 404

	errorResponse := DefaultError(code, testError, nil)

	if errorResponse == nil {
		t.Fatal("DefaultError returned nil")
	}

	if errorResponse.Code == nil {
		t.Fatal("error code is nil")
	}

	if *errorResponse.Code != int64(code) {
		t.Errorf("expected code %d, got %d", code, *errorResponse.Code)
	}

	if errorResponse.Message == nil {
		t.Fatal("error message is nil")
	}

	if *errorResponse.Message != testError.Error() {
		t.Errorf("expected message '%s', got '%s'", testError.Error(), *errorResponse.Message)
	}
}

func TestDefaultError_WithDetails(t *testing.T) {
	testError := errors.New("validation error")
	code := 400
	details := map[string]string{
		"field": "email",
		"error": "invalid format",
	}

	errorResponse := DefaultError(code, testError, details)

	if errorResponse == nil {
		t.Fatal("DefaultError returned nil")
	}

	if errorResponse.Details == nil {
		t.Error("expected details to be set")
	}
}
