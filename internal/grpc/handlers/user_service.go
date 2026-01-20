package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"microservice-template/internal/service"
	proto "microservice-template/protocols/userservice"
)

// UserServiceHandler implements the UserService gRPC service.
type UserServiceHandler struct {
	proto.UnimplementedUserServiceServer
	service service.IService
	log     *logrus.Logger
}

// NewUserServiceHandler creates a new UserService gRPC handler.
// Dependencies:
//   - service: Required. Business logic layer for user operations.
//   - log: Optional. Structured logger (uses default if nil).
func NewUserServiceHandler(svc service.IService, log *logrus.Logger) *UserServiceHandler {
	return &UserServiceHandler{
		service: svc,
		log:     log,
	}
}

// GetUserByEmail retrieves a user by their email address.
func (h *UserServiceHandler) GetUserByEmail(ctx context.Context, req *proto.EmailRequest) (*proto.User, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.Email == "" {
		h.log.Error("email is required but was empty")
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	h.log.Infof("gRPC GetUserByEmail: %s", req.Email)

	user, err := h.service.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Map service errors to gRPC status codes
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("user with email %s not found", req.Email))
		}

		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if errors.Is(err, service.ErrRepositoryUnavailable) {
			h.log.Errorf("repository unavailable: %v", err)
			return nil, status.Error(codes.Unavailable, "service temporarily unavailable")
		}

		// Unknown error - log it and return internal error
		h.log.Errorf("get user by email failed: %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	// Convert domain model to proto
	protoUser := UserToProto(user)
	if protoUser == nil {
		h.log.Error("failed to convert user to proto")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return protoUser, nil
}

// CreateUser creates a new user.
func (h *UserServiceHandler) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.User, error) {
	fmt.Println("HER")

	fmt.Printf("%s\n", req)

	if req == nil {

		fmt.Println("POPAOPAOA")

		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	h.log.Infof("gRPC CreateUser: %s", req.Email)

	fmt.Println(req)

	// Convert proto request to domain model
	user, err := CreateUserRequestToModel(req)
	if err != nil {
		h.log.Errorf("invalid create user request: %v", err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid request: %v", err))
	}

	// Call service to create user

	if err := h.service.CreateUser(ctx, user); err != nil {
		// Map service errors to gRPC status codes
		if errors.Is(err, service.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if errors.Is(err, service.ErrRepositoryUnavailable) {
			h.log.Errorf("repository unavailable: %v", err)
			return nil, status.Error(codes.Unavailable, "service temporarily unavailable")
		}

		// Unknown error - log it and return internal error
		h.log.Errorf("create user failed: %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	// Convert created user to proto and return
	protoUser := UserToProto(user)
	if protoUser == nil {
		h.log.Error("failed to convert created user to proto")
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return protoUser, nil
}

// GetUserByUUID retrieves a user by their UUID.
// TODO: Implement this method when service layer adds GetUserByUUID support.
// Currently the service.IService interface only has GetUserByEmail.
// To implement:
//  1. Add GetUserByUUID(ctx context.Context, uuid uuid.UUID) (*models.User, error) to service.IService
//  2. Implement it in service.Service
//  3. Add UserBy with UUID getter to repository layer if not already present
//  4. Uncomment and implement this handler method
func (h *UserServiceHandler) GetUserByUUID(_ context.Context, _ *proto.UUIDRequest) (*proto.User, error) {
	return nil, status.Error(codes.Unimplemented, "GetUserByUUID is not yet implemented")
}
