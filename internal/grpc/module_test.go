package grpc

import (
	"context"
	"testing"

	"microservice-template/config"
)

func TestModule_Lifecycle(t *testing.T) {
	cfg := &config.GRPCConfig{
		Enabled:          true,
		Host:             "127.0.0.1",
		Port:             0,
		Timeout:          "5s",
		MaxSendMsgSize:   1024 * 1024,
		MaxRecvMsgSize:   1024 * 1024,
		NumStreamWorkers: 0,
	}

	mod := NewModule(cfg, nil)
	ctx := context.Background()

	if err := mod.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if err := mod.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if err := mod.HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}

	if err := mod.Stop(ctx); err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

func TestModule_Name(t *testing.T) {
	mod := NewModule(&config.GRPCConfig{}, nil)

	if mod.Name() != "grpc" {
		t.Errorf("expected name 'grpc', got '%s'", mod.Name())
	}
}

func TestModule_Init_InvalidTimeout(t *testing.T) {
	t.Parallel()

	mod := NewModule(&config.GRPCConfig{Enabled: true, Host: "127.0.0.1", Port: 0, Timeout: "bad"}, nil)
	if err := mod.Init(context.Background()); err == nil {
		t.Fatalf("expected error for invalid timeout")
	}
}

func TestModule_HealthCheck_NotInitialized(t *testing.T) {
	t.Parallel()

	mod := NewModule(&config.GRPCConfig{}, nil)
	if err := mod.HealthCheck(context.Background()); err == nil {
		t.Fatalf("expected error when server not initialized")
	}
}

func TestModule_HealthCheck_NotRunning(t *testing.T) {
	t.Parallel()

	cfg := &config.GRPCConfig{Enabled: true, Host: "127.0.0.1", Port: 0, Timeout: "1s"}
	mod := NewModule(cfg, nil)
	if err := mod.Init(context.Background()); err != nil {
		t.Fatalf("init error: %v", err)
	}

	if err := mod.HealthCheck(context.Background()); err == nil {
		t.Fatalf("expected error when server not running")
	}
}

func TestModule_Stop_Idempotent(t *testing.T) {
	t.Parallel()

	cfg := &config.GRPCConfig{Enabled: true, Host: "127.0.0.1", Port: 0, Timeout: "1s"}
	mod := NewModule(cfg, nil)
	_ = mod.Stop(context.Background())
	_ = mod.Stop(context.Background())
}
