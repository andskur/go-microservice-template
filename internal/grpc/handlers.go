package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"microservice-template/internal/models"
	"microservice-template/internal/service"
	userProto "microservice-template/protocols/user"
)

// UserHandlers implements the gRPC UserService.
type UserHandlers struct {
	userProto.UnimplementedUserServiceServer
	svc service.IService
}

// NewUserHandlers creates a new UserHandlers instance.
func NewUserHandlers(svc service.IService) *UserHandlers {
	return &UserHandlers{svc: svc}
}

// GetUser retrieves a user by email address.
func (h *UserHandlers) GetUser(ctx context.Context, req *userProto.GetUserRequest) (*userProto.User, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	user, err := h.svc.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get user: %v", err)
	}

	return userToProto(user), nil
}

// CreateUser creates a new user.
func (h *UserHandlers) CreateUser(ctx context.Context, req *userProto.CreateUserRequest) (*userProto.User, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	user := &models.User{
		Email:  req.Email,
		Name:   req.Name,
		Status: models.UserActive,
	}

	if err := h.svc.CreateUser(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "create user: %v", err)
	}

	return userToProto(user), nil
}
