package module

import (
	"context"
	"fmt"
	"sync"

	"microservice-template/pkg/logger"
)

// Manager orchestrates the lifecycle of registered modules.
type Manager struct {
	modules []Module
	mu      sync.RWMutex
}

// NewManager creates a new module manager.
func NewManager() *Manager {
	return &Manager{
		modules: make([]Module, 0),
	}
}

// Register adds a module to the manager.
// Modules will be initialized, started, and stopped in registration order.
// Stop will be called in reverse order for proper cleanup.
func (m *Manager) Register(module Module) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Log().Infof("registering module: %s", module.Name())
	m.modules = append(m.modules, module)
}

// InitAll initializes all registered modules in registration order.
// If any module fails to initialize, the process stops and returns the error.
func (m *Manager) InitAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, mod := range m.modules {
		logger.Log().Infof("initializing module: %s", mod.Name())
		if err := mod.Init(ctx); err != nil {
			return fmt.Errorf("failed to init module %s: %w", mod.Name(), err)
		}
	}

	logger.Log().Info("all modules initialized successfully")
	return nil
}

// StartAll starts all registered modules in registration order.
// If any module fails to start, the process stops and returns the error.
func (m *Manager) StartAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, mod := range m.modules {
		logger.Log().Infof("starting module: %s", mod.Name())
		if err := mod.Start(ctx); err != nil {
			return fmt.Errorf("failed to start module %s: %w", mod.Name(), err)
		}
	}

	logger.Log().Info("all modules started successfully")
	return nil
}

// StopAll stops all registered modules in reverse registration order.
// All modules will be stopped even if some fail - errors are logged but don't stop the process.
// This ensures proper cleanup sequence.
func (m *Manager) StopAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var firstError error

	// Stop in reverse order for proper cleanup (LIFO)
	for i := len(m.modules) - 1; i >= 0; i-- {
		mod := m.modules[i]
		logger.Log().Infof("stopping module: %s", mod.Name())

		if err := mod.Stop(ctx); err != nil {
			logger.Log().Errorf("failed to stop module %s: %v", mod.Name(), err)
			if firstError == nil {
				firstError = err
			}
		}
	}

	if firstError != nil {
		return fmt.Errorf("one or more modules failed to stop: %w", firstError)
	}

	logger.Log().Info("all modules stopped successfully")
	return nil
}

// HealthCheckAll checks the health of all registered modules.
// Returns a map of module names to their health check results.
// Modules that are healthy will have nil error.
func (m *Manager) HealthCheckAll(ctx context.Context) map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make(map[string]error, len(m.modules))

	for _, mod := range m.modules {
		results[mod.Name()] = mod.HealthCheck(ctx)
	}

	return results
}

// Count returns the number of registered modules.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.modules)
}
