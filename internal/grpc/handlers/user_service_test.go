package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"microservice-template/internal/models"
	"microservice-template/internal/service"
	proto "microservice-template/protocols/userservice"
)

// mockService is a mock implementation of service.IService for testing.
type mockService struct {
	getUserByEmailFunc func(ctx context.Context, email string) (*models.User, error)
	createUserFunc     func(ctx context.Context, user *models.User) error
}

func (m *mockService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) CreateUser(ctx context.Context, user *models.User) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, user)
	}
	return errors.New("not implemented")
}

func TestNewUserServiceHandler(t *testing.T) {
	t.Parallel()

	svc := &mockService{}
	log := logrus.New()
	handler := NewUserServiceHandler(svc, log)

	if handler == nil {
		t.Fatal("NewUserServiceHandler returned nil")
	}

	if handler.service != svc {
		t.Error("handler service not set correctly")
	}

	if handler.log != log {
		t.Error("handler log not set correctly")
	}
}

func TestGetUserByEmail_Success(t *testing.T) {
	t.Parallel()

	userUUID := uuid.Must(uuid.NewV4())
	createdAt := time.Now().UTC().Truncate(time.Second)
	updatedAt := createdAt.Add(time.Hour)

	expectedUser := &models.User{
		UUID:      userUUID,
		Email:     "test@example.com",
		Name:      "Test User",
		Status:    models.UserActive,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	svc := &mockService{
		getUserByEmailFunc: func(_ context.Context, email string) (*models.User, error) {
			if email != "test@example.com" {
				t.Errorf("unexpected email: %s", email)
			}
			return expectedUser, nil
		},
	}

	handler := NewUserServiceHandler(svc, logrus.New())

	req := &proto.EmailRequest{
		Email: "test@example.com",
	}

	resp, err := handler.GetUserByEmail(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.Email != "test@example.com" {
		t.Errorf("email: expected %s, got %s", "test@example.com", resp.Email)
	}

	if resp.Name != "Test User" {
		t.Errorf("name: expected %s, got %s", "Test User", resp.Name)
	}

	if resp.Status != proto.UserStatus_USER_STATUS_ACTIVE {
		t.Errorf("status: expected %v, got %v", proto.UserStatus_USER_STATUS_ACTIVE, resp.Status)
	}
}

func TestGetUserByEmail_NilRequest(t *testing.T) {
	t.Parallel()

	svc := &mockService{
		getUserByEmailFunc: func(_ context.Context, _ string) (*models.User, error) {
			t.Error("service should not be called with nil request")
			return nil, nil
		},
	}

	handler := NewUserServiceHandler(svc, logrus.New())

	resp, err := handler.GetUserByEmail(context.Background(), nil)

	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected code %v, got %v", codes.InvalidArgument, st.Code())
	}
}

func TestGetUserByEmail_EmptyEmail(t *testing.T) {
	t.Parallel()

	svc := &mockService{
		getUserByEmailFunc: func(_ context.Context, _ string) (*models.User, error) {
			t.Error("service should not be called with empty email")
			return nil, nil
		},
	}

	handler := NewUserServiceHandler(svc, logrus.New())

	req := &proto.EmailRequest{
		Email: "",
	}

	resp, err := handler.GetUserByEmail(context.Background(), req)

	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected code %v, got %v", codes.InvalidArgument, st.Code())
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	t.Parallel()

	svc := &mockService{
		getUserByEmailFunc: func(_ context.Context, _ string) (*models.User, error) {
			return nil, service.ErrNotFound
		},
	}

	handler := NewUserServiceHandler(svc, logrus.New())

	req := &proto.EmailRequest{
		Email: "notfound@example.com",
	}

	resp, err := handler.GetUserByEmail(context.Background(), req)

	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.NotFound {
		t.Errorf("expected code %v, got %v", codes.NotFound, st.Code())
	}
}

func TestGetUserByEmail_InvalidInput(t *testing.T) {
	t.Parallel()

	svc := &mockService{
		getUserByEmailFunc: func(_ context.Context, _ string) (*models.User, error) {
			return nil, service.ErrInvalidInput
		},
	}

	handler := NewUserServiceHandler(svc, logrus.New())

	req := &proto.EmailRequest{
		Email: "invalid",
	}

	resp, err := handler.GetUserByEmail(context.Background(), req)

	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected code %v, got %v", codes.InvalidArgument, st.Code())
	}
}

func TestGetUserByEmail_RepositoryUnavailable(t *testing.T) {
	t.Parallel()

	svc := &mockService{
		getUserByEmailFunc: func(_ context.Context, _ string) (*models.User, error) {
			return nil, service.ErrRepositoryUnavailable
		},
	}

	handler := NewUserServiceHandler(svc, logrus.New())

	req := &proto.EmailRequest{
		Email: "test@example.com",
	}

	resp, err := handler.GetUserByEmail(context.Background(), req)

	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.Unavailable {
		t.Errorf("expected code %v, got %v", codes.Unavailable, st.Code())
	}
}

func TestGetUserByEmail_InternalError(t *testing.T) {
	t.Parallel()

	svc := &mockService{
		getUserByEmailFunc: func(_ context.Context, _ string) (*models.User, error) {
			return nil, errors.New("unexpected database error")
		},
	}

	handler := NewUserServiceHandler(svc, logrus.New())

	req := &proto.EmailRequest{
		Email: "test@example.com",
	}

	resp, err := handler.GetUserByEmail(context.Background(), req)

	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.Internal {
		t.Errorf("expected code %v, got %v", codes.Internal, st.Code())
	}
}

func TestCreateUser_Success(t *testing.T) {
	t.Parallel()

	var capturedUser *models.User

	svc := &mockService{
		createUserFunc: func(_ context.Context, user *models.User) error {
			capturedUser = user
			// Simulate repository setting UUID and timestamps
			user.UUID = uuid.Must(uuid.NewV4())
			user.CreatedAt = time.Now().UTC()
			user.UpdatedAt = user.CreatedAt
			return nil
		},
	}

	handler := NewUserServiceHandler(svc, logrus.New())

	req := &proto.CreateUserRequest{
		Email:  "new@example.com",
		Name:   "New User",
		Status: proto.UserStatus_USER_STATUS_ACTIVE,
	}

	resp, err := handler.CreateUser(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if capturedUser == nil {
		t.Fatal("service was not called")
	}

	if capturedUser.Email != "new@example.com" {
		t.Errorf("email: expected %s, got %s", "new@example.com", capturedUser.Email)
	}

	if capturedUser.Name != "New User" {
		t.Errorf("name: expected %s, got %s", "New User", capturedUser.Name)
	}

	if capturedUser.Status != models.UserActive {
		t.Errorf("status: expected %v, got %v", models.UserActive, capturedUser.Status)
	}

	if resp.Email != "new@example.com" {
		t.Errorf("response email: expected %s, got %s", "new@example.com", resp.Email)
	}

	if resp.Name != "New User" {
		t.Errorf("response name: expected %s, got %s", "New User", resp.Name)
	}
}

func TestCreateUser_NilRequest(t *testing.T) {
	t.Parallel()

	svc := &mockService{
		createUserFunc: func(_ context.Context, _ *models.User) error {
			t.Error("service should not be called with nil request")
			return nil
		},
	}

	handler := NewUserServiceHandler(svc, logrus.New())

	resp, err := handler.CreateUser(context.Background(), nil)

	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected code %v, got %v", codes.InvalidArgument, st.Code())
	}
}

func TestCreateUser_InvalidStatus(t *testing.T) {
	t.Parallel()

	svc := &mockService{
		createUserFunc: func(_ context.Context, _ *models.User) error {
			t.Error("service should not be called with invalid request")
			return nil
		},
	}

	handler := NewUserServiceHandler(svc, logrus.New())

	req := &proto.CreateUserRequest{
		Email:  "test@example.com",
		Name:   "Test User",
		Status: proto.UserStatus(999), // Invalid status
	}

	resp, err := handler.CreateUser(context.Background(), req)

	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected code %v, got %v", codes.InvalidArgument, st.Code())
	}
}

func TestCreateUser_ServiceInvalidInput(t *testing.T) {
	t.Parallel()

	svc := &mockService{
		createUserFunc: func(_ context.Context, _ *models.User) error {
			return service.ErrInvalidInput
		},
	}

	handler := NewUserServiceHandler(svc, logrus.New())

	req := &proto.CreateUserRequest{
		Email:  "invalid",
		Name:   "Test User",
		Status: proto.UserStatus_USER_STATUS_ACTIVE,
	}

	resp, err := handler.CreateUser(context.Background(), req)

	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected code %v, got %v", codes.InvalidArgument, st.Code())
	}
}

func TestCreateUser_RepositoryUnavailable(t *testing.T) {
	t.Parallel()

	svc := &mockService{
		createUserFunc: func(_ context.Context, _ *models.User) error {
			return service.ErrRepositoryUnavailable
		},
	}

	handler := NewUserServiceHandler(svc, logrus.New())

	req := &proto.CreateUserRequest{
		Email:  "test@example.com",
		Name:   "Test User",
		Status: proto.UserStatus_USER_STATUS_ACTIVE,
	}

	resp, err := handler.CreateUser(context.Background(), req)

	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.Unavailable {
		t.Errorf("expected code %v, got %v", codes.Unavailable, st.Code())
	}
}

func TestCreateUser_InternalError(t *testing.T) {
	t.Parallel()

	svc := &mockService{
		createUserFunc: func(_ context.Context, _ *models.User) error {
			return errors.New("unexpected database error")
		},
	}

	handler := NewUserServiceHandler(svc, logrus.New())

	req := &proto.CreateUserRequest{
		Email:  "test@example.com",
		Name:   "Test User",
		Status: proto.UserStatus_USER_STATUS_ACTIVE,
	}

	resp, err := handler.CreateUser(context.Background(), req)

	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.Internal {
		t.Errorf("expected code %v, got %v", codes.Internal, st.Code())
	}
}

func TestGetUserByUUID_Unimplemented(t *testing.T) {
	t.Parallel()

	svc := &mockService{}
	handler := NewUserServiceHandler(svc, logrus.New())

	req := &proto.UUIDRequest{
		Uuid: uuid.Must(uuid.NewV4()).Bytes(),
	}

	resp, err := handler.GetUserByUUID(context.Background(), req)

	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}

	if st.Code() != codes.Unimplemented {
		t.Errorf("expected code %v, got %v", codes.Unimplemented, st.Code())
	}
}
