// Package module provides interfaces and management for optional application components.
package module

import "context"

// Module represents an optional application component with standard lifecycle.
// All modules must implement this interface to be managed by the application.
type Module interface {
	// Name returns the unique module identifier (e.g., "database", "http-server").
	Name() string

	// Init initializes the module resources (establish connections, prepare resources).
	// Called during App.Init() phase before any module starts.
	// Should be idempotent - safe to call multiple times.
	Init(ctx context.Context) error

	// Start begins module operation (start servers, consumers, background workers).
	// Called during App.Serve() phase after all modules are initialized.
	// Should be non-blocking - use goroutines for long-running operations.
	// Return error if the module fails to start.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the module.
	// Called during App.Stop() phase in reverse registration order.
	// Must handle being called even if Start was never called or failed.
	// Should respect context timeout for graceful shutdown.
	Stop(ctx context.Context) error

	// HealthCheck returns the current health status of the module.
	// Used for readiness/liveness probes.
	// Should be quick (< 2 seconds) and return error if module is unhealthy.
	HealthCheck(ctx context.Context) error
}
