package internal

import (
	"testing"

	"microservice-template/config"
)

func TestNewApplication(t *testing.T) {
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if app == nil {
		t.Fatal("expected app, got nil")
	}

	if app.modules == nil {
		t.Error("modules manager should be initialized")
	}

	if app.modules.Count() != 0 {
		t.Errorf("expected 0 modules initially, got %d", app.modules.Count())
	}
}

func TestApp_Init_NoModules(t *testing.T) {
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	err = app.Init()
	if err != nil {
		t.Errorf("Init should succeed with no modules: %v", err)
	}

	// Without database configured, no modules should be registered
	if app.modules.Count() != 0 {
		t.Errorf("expected 0 modules, got %d", app.modules.Count())
	}

	// Service should be nil when database is disabled
	if app.Service() != nil {
		t.Error("Service should be nil when database is not enabled")
	}
}

func TestApp_Stop(t *testing.T) {
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// Should be safe to call Stop even without Init
	err = app.Stop()
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

func TestApp_Modules(t *testing.T) {
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	manager := app.Modules()
	if manager == nil {
		t.Error("Modules() should return manager")
	}
}

func TestApp_Config(t *testing.T) {
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	cfg := app.Config()
	if cfg == nil {
		t.Error("Config() should return configuration")
	}
}

func TestApp_Version(t *testing.T) {
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	version := app.Version()
	if version == "" {
		t.Error("Version() should return non-empty string")
	}
}

func TestApp_RegisterModules_WithDatabase(t *testing.T) {
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	app.config.Database = &config.DatabaseConfig{
		Enabled: true,
		Driver:  "postgres",
		Host:    "localhost",
		Port:    5432,
	}

	if err := app.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if app.modules.Count() != 1 {
		t.Errorf("expected 1 module, got %d", app.modules.Count())
	}

	if app.Service() == nil {
		t.Error("Service should be initialized when database is enabled")
	}
}

func TestCreateAddr(t *testing.T) {
	tests := []struct { //nolint:govet // field alignment not critical in test table
		name string
		host string
		port int
		want string
	}{
		{name: "simple", host: "localhost", port: 8080, want: "localhost:8080"},
		{name: "ip", host: "127.0.0.1", port: 80, want: "127.0.0.1:80"},
		{name: "ipv6", host: "[::1]", port: 443, want: "[::1]:443"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateAddr(tt.host, tt.port)
			if got != tt.want {
				t.Fatalf("CreateAddr(%s, %d) = %s, want %s", tt.host, tt.port, got, tt.want)
			}
		})
	}
}
