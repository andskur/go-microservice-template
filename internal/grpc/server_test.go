package grpc

import (
	"testing"

	"microservice-template/config"
)

func TestNewServer_ValidConfig(t *testing.T) {
	t.Parallel()

	srv, err := NewServer(&config.GRPCConfig{
		Host:             "127.0.0.1",
		Port:             0,
		Timeout:          "5s",
		MaxSendMsgSize:   1 << 20,
		MaxRecvMsgSize:   1 << 20,
		NumStreamWorkers: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { srv.GracefulStop() })

	if srv == nil || srv.Server() == nil {
		t.Fatalf("expected non-nil server")
	}
}

func TestNewServer_InvalidTimeout(t *testing.T) {
	t.Parallel()

	_, err := NewServer(&config.GRPCConfig{
		Host:    "127.0.0.1",
		Port:    0,
		Timeout: "not-a-duration",
	})
	if err == nil {
		t.Fatalf("expected error for invalid timeout")
	}
}

func TestNewServer_InvalidPort(t *testing.T) {
	t.Parallel()

	_, err := NewServer(&config.GRPCConfig{
		Host:    "127.0.0.1",
		Port:    -1,
		Timeout: "1s",
	})
	if err == nil {
		t.Fatalf("expected error for invalid port")
	}
}

func TestServer_RegisterHealthService(t *testing.T) {
	t.Parallel()

	srv, err := NewServer(&config.GRPCConfig{
		Host:    "127.0.0.1",
		Port:    0,
		Timeout: "1s",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { srv.GracefulStop() })

	if err := srv.RegisterHealthService(); err != nil {
		t.Fatalf("register health failed: %v", err)
	}
}

func TestServer_StreamWorkers(t *testing.T) {
	t.Parallel()

	cfg := &config.GRPCConfig{
		Host:             "127.0.0.1",
		Port:             0,
		Timeout:          "1s",
		NumStreamWorkers: 4,
	}

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { srv.GracefulStop() })

	if srv.Server() == nil {
		t.Fatalf("expected server to be initialized")
	}
}

func TestServer_IsRunningAndMarkRunning(t *testing.T) {
	t.Parallel()

	srv, err := NewServer(&config.GRPCConfig{Host: "127.0.0.1", Port: 0, Timeout: "1s"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv.IsRunning() {
		t.Fatalf("expected not running")
	}

	srv.MarkRunning()
	if !srv.IsRunning() {
		t.Fatalf("expected running after MarkRunning")
	}
}

func TestServer_GracefulStop_Idempotent(t *testing.T) {
	t.Parallel()

	srv, err := NewServer(&config.GRPCConfig{Host: "127.0.0.1", Port: 0, Timeout: "1s"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv.GracefulStop()
	srv.GracefulStop() // second call should not panic
}

func TestServer_MessageSizeLimits(t *testing.T) {
	t.Parallel()

	srv, err := NewServer(&config.GRPCConfig{
		Host:           "127.0.0.1",
		Port:           0,
		Timeout:        "1s",
		MaxSendMsgSize: 2 << 20,
		MaxRecvMsgSize: 3 << 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { srv.GracefulStop() })

	// Option verification is internal to grpc.NewServer; we ensure construction succeeded
	if srv.Server() == nil {
		t.Fatalf("expected server to be initialized")
	}
}

func TestNewServer_KeepsAliveConfigValid(t *testing.T) {
	t.Parallel()

	srv, err := NewServer(&config.GRPCConfig{Host: "127.0.0.1", Port: 0, Timeout: "2s"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { srv.GracefulStop() })

	// If keep-alive options were invalid, NewServer would have errored; reaching here is success.
}
