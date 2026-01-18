package module

import (
	"context"
	"errors"
	"testing"
)

// mockModule is a test double for Module interface.
type mockModule struct {
	initErr      error
	startErr     error
	stopErr      error
	healthErr    error
	name         string
	initCalled   bool
	startCalled  bool
	stopCalled   bool
	healthCalled bool
}

func (m *mockModule) Name() string { return m.name }

func (m *mockModule) Init(_ context.Context) error {
	m.initCalled = true
	return m.initErr
}

func (m *mockModule) Start(_ context.Context) error {
	m.startCalled = true
	return m.startErr
}

func (m *mockModule) Stop(_ context.Context) error {
	m.stopCalled = true
	return m.stopErr
}

func (m *mockModule) HealthCheck(_ context.Context) error {
	m.healthCalled = true
	return m.healthErr
}

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("expected manager, got nil")
	}

	if manager.Count() != 0 {
		t.Errorf("expected 0 modules, got %d", manager.Count())
	}
}

func TestManager_Register(t *testing.T) {
	manager := NewManager()

	mod1 := &mockModule{name: "module1"}
	mod2 := &mockModule{name: "module2"}

	manager.Register(mod1)
	manager.Register(mod2)

	if manager.Count() != 2 {
		t.Errorf("expected 2 modules, got %d", manager.Count())
	}
}

func TestManager_InitAll_Success(t *testing.T) {
	manager := NewManager()

	mod1 := &mockModule{name: "module1"}
	mod2 := &mockModule{name: "module2"}

	manager.Register(mod1)
	manager.Register(mod2)

	err := manager.InitAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mod1.initCalled || !mod2.initCalled {
		t.Error("not all modules were initialized")
	}
}

func TestManager_InitAll_Failure(t *testing.T) {
	manager := NewManager()

	mod1 := &mockModule{name: "module1"}
	mod2 := &mockModule{name: "module2", initErr: errors.New("init failed")}

	manager.Register(mod1)
	manager.Register(mod2)

	err := manager.InitAll(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !mod1.initCalled {
		t.Error("module1 should have been initialized")
	}
}

func TestManager_StartAll_Success(t *testing.T) {
	manager := NewManager()

	mod1 := &mockModule{name: "module1"}
	mod2 := &mockModule{name: "module2"}

	manager.Register(mod1)
	manager.Register(mod2)

	err := manager.StartAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mod1.startCalled || !mod2.startCalled {
		t.Error("not all modules were started")
	}
}

func TestManager_StartAll_Failure(t *testing.T) {
	manager := NewManager()

	mod1 := &mockModule{name: "module1"}
	mod2 := &mockModule{name: "module2", startErr: errors.New("start failed")}

	manager.Register(mod1)
	manager.Register(mod2)

	err := manager.StartAll(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !mod1.startCalled {
		t.Error("module1 should have been started")
	}
}

func TestManager_StopAll_Success(t *testing.T) {
	manager := NewManager()

	mod1 := &mockModule{name: "module1"}
	mod2 := &mockModule{name: "module2"}

	manager.Register(mod1)
	manager.Register(mod2)

	err := manager.StopAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mod1.stopCalled || !mod2.stopCalled {
		t.Error("not all modules were stopped")
	}
}

func TestManager_StopAll_ContinuesOnError(t *testing.T) {
	manager := NewManager()

	mod1 := &mockModule{name: "module1"}
	mod2 := &mockModule{name: "module2", stopErr: errors.New("stop failed")}

	manager.Register(mod1)
	manager.Register(mod2)

	err := manager.StopAll(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Both should be stopped despite error
	if !mod1.stopCalled || !mod2.stopCalled {
		t.Error("all modules should be stopped even if one fails")
	}
}

func TestManager_StopAll_ReverseOrder(t *testing.T) {
	manager := NewManager()

	stopOrder := make([]string, 0)
	stopOrderPtr := &stopOrder

	// Create modules that track stop order
	mod1 := &trackingModule{name: "module1", stopOrder: stopOrderPtr}
	mod2 := &trackingModule{name: "module2", stopOrder: stopOrderPtr}
	mod3 := &trackingModule{name: "module3", stopOrder: stopOrderPtr}

	manager.Register(mod1)
	manager.Register(mod2)
	manager.Register(mod3)

	manager.StopAll(context.Background())

	// Should stop in reverse order: module3, module2, module1
	if len(stopOrder) != 3 {
		t.Fatalf("expected 3 stops, got %d", len(stopOrder))
	}

	if stopOrder[0] != "module3" || stopOrder[1] != "module2" || stopOrder[2] != "module1" {
		t.Errorf("wrong stop order: %v (expected [module3, module2, module1])", stopOrder)
	}
}

// trackingModule tracks the order of Stop calls.
type trackingModule struct {
	stopOrder *[]string
	name      string
}

func (m *trackingModule) Name() string { return m.name }

func (m *trackingModule) Init(_ context.Context) error { return nil }

func (m *trackingModule) Start(_ context.Context) error { return nil }

func (m *trackingModule) Stop(_ context.Context) error {
	*m.stopOrder = append(*m.stopOrder, m.name)
	return nil
}

func (m *trackingModule) HealthCheck(_ context.Context) error { return nil }

func TestManager_HealthCheckAll(t *testing.T) {
	manager := NewManager()

	mod1 := &mockModule{name: "module1"}
	mod2 := &mockModule{name: "module2", healthErr: errors.New("unhealthy")}

	manager.Register(mod1)
	manager.Register(mod2)

	results := manager.HealthCheckAll(context.Background())

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	if results["module1"] != nil {
		t.Error("module1 should be healthy")
	}

	if results["module2"] == nil {
		t.Error("module2 should be unhealthy")
	}

	if !mod1.healthCalled || !mod2.healthCalled {
		t.Error("not all modules were health checked")
	}
}

func TestManager_Count(t *testing.T) {
	manager := NewManager()

	if manager.Count() != 0 {
		t.Errorf("expected 0 modules, got %d", manager.Count())
	}

	manager.Register(&mockModule{name: "module1"})

	if manager.Count() != 1 {
		t.Errorf("expected 1 module, got %d", manager.Count())
	}

	manager.Register(&mockModule{name: "module2"})
	manager.Register(&mockModule{name: "module3"})

	if manager.Count() != 3 {
		t.Errorf("expected 3 modules, got %d", manager.Count())
	}
}
