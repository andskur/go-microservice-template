package grpc

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"microservice-template/config"
	"microservice-template/internal/models"
	userProto "microservice-template/protocols/user"
)

// mockIntegrationService is a thread-safe in-memory implementation of IService for integration testing.
type mockIntegrationService struct {
	users map[string]*models.User
	mu    sync.Mutex
}

func newMockIntegrationService() *mockIntegrationService {
	return &mockIntegrationService{users: make(map[string]*models.User)}
}

func (m *mockIntegrationService) CreateUser(_ context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.Email] = user
	return nil
}

func (m *mockIntegrationService) GetUserByEmail(_ context.Context, email string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user, ok := m.users[email]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("user not found: %s", email)
}

func TestIntegration_UserHandlers(t *testing.T) {
	t.Parallel()

	cfg := &config.GRPCConfig{
		Enabled:        true,
		Host:           "127.0.0.1",
		Port:           0,
		Timeout:        "2s",
		MaxSendMsgSize: 60 * 1024 * 1024,
		MaxRecvMsgSize: 60 * 1024 * 1024,
	}

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}
	t.Cleanup(func() { srv.GracefulStop() })

	if err := srv.RegisterHealthService(); err != nil {
		t.Fatalf("RegisterHealthService error: %v", err)
	}

	mockSvc := newMockIntegrationService()
	userProto.RegisterUserServiceServer(srv.Server(), NewUserHandlers(mockSvc))

	addr := srv.listener.Addr().String()

	srv.MarkRunning()
	go func() {
		_ = srv.Serve()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Dial error: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	client := userProto.NewUserServiceClient(conn)

	createResp, err := client.CreateUser(ctx, &userProto.CreateUserRequest{Email: "it@example.com", Name: "Integration"})
	if err != nil {
		t.Fatalf("CreateUser error: %v", err)
	}
	if createResp.Email != "it@example.com" || createResp.Name != "Integration" {
		t.Fatalf("unexpected CreateUser response: %+v", createResp)
	}

	getResp, err := client.GetUser(ctx, &userProto.GetUserRequest{Email: "it@example.com"})
	if err != nil {
		t.Fatalf("GetUser error: %v", err)
	}
	if getResp.Email != "it@example.com" || getResp.Name != "Integration" {
		t.Fatalf("unexpected GetUser response: %+v", getResp)
	}
}
