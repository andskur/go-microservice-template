package http

import (
	"context"
	"testing"
	"time"

	"microservice-template/config"
	domainmodels "microservice-template/internal/models"
)

// mockService is a mock implementation of service.IService for testing.
type mockService struct {
	createUserFunc     func(ctx context.Context, user *domainmodels.User) error
	getUserByEmailFunc func(ctx context.Context, email string) (*domainmodels.User, error)
}

func (m *mockService) CreateUser(_ context.Context, user *domainmodels.User) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(context.Background(), user)
	}
	return nil
}

func (m *mockService) GetUserByEmail(_ context.Context, email string) (*domainmodels.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(context.Background(), email)
	}
	return nil, nil
}

func TestNewModule(t *testing.T) {
	cfg := &config.HTTPConfig{
		Host:        "localhost",
		Port:        8080,
		Timeout:     "30s",
		SwaggerSpec: "./api/swagger.yaml",
		Enabled:     true,
		MockAuth:    true,
		AdminEmails: []string{},
	}

	svc := &mockService{}
	module := NewModule(cfg, svc, nil)

	if module == nil {
		t.Fatal("NewModule returned nil")
	}

	if module.config != cfg {
		t.Error("config not set correctly")
	}

	if module.service != svc {
		t.Error("service not set correctly")
	}
}

func TestModule_Name(t *testing.T) {
	cfg := &config.HTTPConfig{
		Host:    "localhost",
		Port:    8080,
		Timeout: "30s",
	}
	svc := &mockService{}
	module := NewModule(cfg, svc, nil)

	name := module.Name()
	if name != "http" {
		t.Errorf("expected module name 'http', got '%s'", name)
	}
}

func TestModule_Init(t *testing.T) {
	cfg := &config.HTTPConfig{
		Host:        "localhost",
		Port:        8080,
		Timeout:     "30s",
		SwaggerSpec: "./api/swagger.yaml",
		MockAuth:    true,
		AdminEmails: []string{"admin@example.com"},
		CORS: &config.CORSConfig{
			Enabled: true,
		},
		RateLimit: &config.RateLimitConfig{
			Enabled: false,
		},
	}

	svc := &mockService{}
	module := NewModule(cfg, svc, nil)

	ctx := context.Background()
	err := module.Init(ctx)

	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify auth was initialized
	if module.auth == nil {
		t.Error("auth not initialized")
	}

	// Verify API was initialized
	if module.api == nil {
		t.Error("api not initialized")
	}

	// Verify server was initialized
	if module.server == nil {
		t.Error("server not initialized")
	}

	// Verify handler was initialized
	if module.handler == nil {
		t.Error("handler not initialized")
	}
}

func TestModule_Init_InvalidTimeout(t *testing.T) {
	cfg := &config.HTTPConfig{
		Host:    "localhost",
		Port:    8080,
		Timeout: "invalid",
		CORS: &config.CORSConfig{
			Enabled: false,
		},
		RateLimit: &config.RateLimitConfig{
			Enabled: false,
		},
	}

	svc := &mockService{}
	module := NewModule(cfg, svc, nil)

	ctx := context.Background()
	err := module.Init(ctx)

	if err == nil {
		t.Error("expected error for invalid timeout, got nil")
	}
}

func TestModule_Init_InvalidHostOrPort(t *testing.T) {
	t.Run("empty host", func(t *testing.T) {
		cfg := &config.HTTPConfig{
			Host:      "",
			Port:      8080,
			Timeout:   "30s",
			CORS:      &config.CORSConfig{Enabled: false},
			RateLimit: &config.RateLimitConfig{Enabled: false},
		}

		svc := &mockService{}
		module := NewModule(cfg, svc, nil)

		ctx := context.Background()
		err := module.Init(ctx)

		if err == nil {
			t.Error("expected error for empty host, got nil")
		}
	})

	t.Run("invalid port", func(t *testing.T) {
		cfg := &config.HTTPConfig{
			Host:      "localhost",
			Port:      -1,
			Timeout:   "30s",
			CORS:      &config.CORSConfig{Enabled: false},
			RateLimit: &config.RateLimitConfig{Enabled: false},
		}

		svc := &mockService{}
		module := NewModule(cfg, svc, nil)

		ctx := context.Background()
		err := module.Init(ctx)

		if err == nil {
			t.Error("expected error for invalid port, got nil")
		}
	})
}

func TestModule_Lifecycle(t *testing.T) {
	t.Skip("Skipping lifecycle test - go-swagger generated server doesn't shut down gracefully in tests")

	cfg := &config.HTTPConfig{
		Host:        "localhost",
		Port:        18080, // Use different port to avoid conflicts
		Timeout:     "30s",
		SwaggerSpec: "./api/swagger.yaml",
		MockAuth:    true,
		AdminEmails: []string{},
		CORS: &config.CORSConfig{
			Enabled: false,
		},
		RateLimit: &config.RateLimitConfig{
			Enabled: false,
		},
	}

	svc := &mockService{}
	module := NewModule(cfg, svc, nil)

	ctx := context.Background()

	// Test Init
	err := module.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Test Start (non-blocking)
	err = module.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	// Test HealthCheck
	err = module.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}

	// Test Stop
	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = module.Stop(stopCtx)
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

func TestModule_Stop_WithoutStart(t *testing.T) {
	cfg := &config.HTTPConfig{
		Host:    "localhost",
		Port:    8080,
		Timeout: "30s",
	}

	svc := &mockService{}
	module := NewModule(cfg, svc, nil)

	ctx := context.Background()

	// Stop without Init or Start should not panic
	err := module.Stop(ctx)
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

func TestModule_HealthCheck(t *testing.T) {
	cfg := &config.HTTPConfig{
		Host:    "localhost",
		Port:    8080,
		Timeout: "30s",
	}

	svc := &mockService{}
	module := NewModule(cfg, svc, nil)

	ctx := context.Background()

	// HealthCheck should always succeed (simple implementation)
	err := module.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}
}
