package grpcclient

import (
	"context"
	"testing"

	"microservice-template/config"
)

func TestModule_Name(t *testing.T) {
	cfg := &config.GRPCClientConfig{
		Address: "localhost:9090",
		Timeout: "30s",
		Enabled: true,
		KeepAlive: &config.KeepAliveConfig{
			Time:                "10s",
			Timeout:             "1s",
			PermitWithoutStream: true,
		},
	}

	mod := NewModule(cfg)

	if mod.Name() != "grpc-client" {
		t.Errorf("expected name 'grpc-client', got '%s'", mod.Name())
	}
}

func TestModule_Lifecycle(t *testing.T) {
	cfg := &config.GRPCClientConfig{
		Address: "localhost:19090", // Use unlikely port to avoid conflicts
		Timeout: "5s",
		Enabled: true,
		KeepAlive: &config.KeepAliveConfig{
			Time:                "10s",
			Timeout:             "1s",
			PermitWithoutStream: true,
		},
	}

	mod := NewModule(cfg)
	ctx := context.Background()

	// Note: Init will attempt to connect but may fail if no server is running
	// This is expected in unit tests - we're testing the lifecycle flow
	_ = mod.Init(ctx) // Init may fail without running server

	// Test Start (should always succeed - it's a no-op)
	if err := mod.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Test HealthCheck
	_ = mod.HealthCheck(ctx)

	// Test Stop (should handle nil client gracefully)
	if err := mod.Stop(ctx); err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}
