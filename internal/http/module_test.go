package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
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

func lifecycleConfig() *config.HTTPConfig {
	return &config.HTTPConfig{
		Host:        "127.0.0.1",
		Port:        0, // ephemeral port
		Timeout:     "5s",
		MockAuth:    true,
		AdminEmails: []string{},
		CORS:        &config.CORSConfig{Enabled: false},
		RateLimit:   &config.RateLimitConfig{Enabled: true, RequestsPerSec: 1, Burst: 1},
	}
}

// testClient does not keep connections alive, so no idle (or speculatively
// dialed, never used) connection delays the graceful shutdown in Stop.
var testClient = &http.Client{Transport: &http.Transport{DisableKeepAlives: true}, Timeout: 5 * time.Second}

func get(t *testing.T, url string) int {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := testClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	_ = resp.Body.Close()
	return resp.StatusCode
}

func TestModule_Lifecycle(t *testing.T) {
	module := NewModule(lifecycleConfig(), &mockService{}, nil)

	var unhealthy atomic.Bool
	module.SetHealthChecker(func(_ context.Context) map[string]error {
		if unhealthy.Load() {
			return map[string]error{"repository": errors.New("db down"), "http": nil}
		}
		return map[string]error{"repository": nil, "http": nil}
	})

	ctx := context.Background()
	if err := module.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if err := module.HealthCheck(ctx); err == nil {
		t.Error("expected HealthCheck to fail before Start")
	}
	if err := module.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	base := "http://" + module.Addr().String()

	if err := module.HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}

	// Both health paths respond and are exempt from rate limiting (burst 1).
	for i := 0; i < 3; i++ {
		for _, p := range []string{"/health", "/api/v1/health"} {
			if code := get(t, base+p); code != http.StatusOK {
				t.Fatalf("GET %s #%d: got %d, want 200", p, i, code)
			}
		}
	}

	// Health reflects module health checks.
	unhealthy.Store(true)
	for _, p := range []string{"/health", "/api/v1/health"} {
		if code := get(t, base+p); code != http.StatusServiceUnavailable {
			t.Fatalf("GET %s with failing module: got %d, want 503", p, code)
		}
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := module.Stop(stopCtx); err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	// The serve goroutine ends cleanly: Failed is closed without an error.
	select {
	case err, ok := <-module.Failed():
		if ok {
			t.Fatalf("unexpected failure after Stop: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Failed channel not closed after Stop")
	}
}

func TestModule_StartFailsWhenPortInUse(t *testing.T) {
	first := NewModule(lifecycleConfig(), &mockService{}, nil)
	ctx := context.Background()
	if err := first.Init(ctx); err != nil {
		t.Fatal(err)
	}
	if err := first.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = first.Stop(context.Background()) })

	cfg := lifecycleConfig()
	cfg.Port = first.Addr().(*net.TCPAddr).Port
	second := NewModule(cfg, &mockService{}, nil)
	if err := second.Init(ctx); err != nil {
		t.Fatal(err)
	}
	if err := second.Start(ctx); err == nil {
		_ = second.Stop(ctx)
		t.Fatal("expected Start to fail when the port is in use")
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

	module := NewModule(cfg, &mockService{}, nil)

	// Not serving yet: unhealthy.
	if err := module.HealthCheck(context.Background()); err == nil {
		t.Error("expected HealthCheck to fail when the server is not running")
	}
}
