package userservice

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	proto "microservice-template/protocols/userservice"
)

// Client implements IUserServiceClient for external user service.
type Client struct {
	client  proto.UserServiceClient
	conn    *grpc.ClientConn
	addr    string
	timeout time.Duration
}

var codeReason = map[codes.Code]string{
	codes.OK:                 "unexpected ok status",
	codes.Canceled:           "canceled",
	codes.Unknown:            "unknown error",
	codes.InvalidArgument:    "invalid argument",
	codes.NotFound:           "not found",
	codes.AlreadyExists:      "already exists",
	codes.PermissionDenied:   "permission denied",
	codes.ResourceExhausted:  "resource exhausted",
	codes.FailedPrecondition: "failed precondition",
	codes.Aborted:            "aborted",
	codes.OutOfRange:         "out of range",
	codes.Unimplemented:      "unimplemented",
	codes.Internal:           "internal error",
	codes.Unavailable:        "service unavailable",
	codes.DeadlineExceeded:   "deadline exceeded",
	codes.DataLoss:           "data loss",
	codes.Unauthenticated:    "unauthenticated",
}

// New creates a new user service client.
func New(address string, timeout time.Duration, kacp keepalive.ClientParameters) (IUserServiceClient, error) {
	c := &Client{
		addr:    address,
		timeout: timeout,
	}

	if err := c.initConn(kacp); err != nil {
		return nil, fmt.Errorf("init connection: %w", err)
	}

	c.client = proto.NewUserServiceClient(c.conn)
	return c, nil
}

// initConn initializes the gRPC connection.
func (c *Client) initConn(kacp keepalive.ClientParameters) error {
	conn, err := grpc.NewClient(
		c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // TODO: Add TLS support via config
		grpc.WithKeepaliveParams(kacp),
	)
	if err != nil {
		return fmt.Errorf("dial %s: %w", c.addr, err)
	}

	c.conn = conn
	return nil
}

// UserByEmail retrieves a user by email address.
func (c *Client) UserByEmail(ctx context.Context, email string) (*proto.User, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &proto.EmailRequest{Email: email}

	resp, err := c.client.GetUserByEmail(ctx, req)
	if err != nil {
		return nil, mapGRPCError(err, "get user by email")
	}

	return resp, nil
}

// UserByUUID retrieves a user by UUID bytes.
func (c *Client) UserByUUID(ctx context.Context, uuidBytes []byte) (*proto.User, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &proto.UUIDRequest{Uuid: uuidBytes}

	resp, err := c.client.GetUserByUUID(ctx, req)
	if err != nil {
		return nil, mapGRPCError(err, "get user by uuid")
	}

	return resp, nil
}

// CreateUser creates a new user.
func (c *Client) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.User, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.CreateUser(ctx, req)
	if err != nil {
		return nil, mapGRPCError(err, "create user")
	}

	return resp, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// mapGRPCError maps gRPC errors to application errors with context.
func mapGRPCError(err error, operation string) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("%s: %w", operation, err)
	}

	reason, found := codeReason[st.Code()]
	if !found {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return fmt.Errorf("%s: %s: %w", operation, reason, err)
}
