# Module Development Guide

## Overview

The module system provides a standard way to add optional components to your microservice. This guide explains how to create and integrate custom modules.

## Module Interface

All modules must implement the `module.Module` interface defined in `internal/module/module.go`:

```go
type Module interface {
    Name() string
    Init(ctx context.Context) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    HealthCheck(ctx context.Context) error
}
```

### Lifecycle Methods

- **`Name()`**: Returns a unique identifier for the module (e.g., "database", "http-server")
- **`Init(ctx)`**: Initializes resources (establish connections, prepare resources). Called during `App.Init()` before any module starts. Should be idempotent.
- **`Start(ctx)`**: Begins module operation (start servers, consumers, background workers). Called during `App.Serve()` after all modules are initialized. Should be non-blocking - use goroutines for long-running operations.
- **`Stop(ctx)`**: Gracefully shuts down the module. Called during `App.Stop()` in reverse registration order. Must handle being called even if Start was never called or failed. Should respect context timeout.
- **`HealthCheck(ctx)`**: Returns current health status. Should be quick (< 2 seconds) and return error if module is unhealthy.

## Creating a Module

### Step 1: Define Module Configuration

Add your module's config to `config/scheme.go`:

```go
// MyModuleConfig holds settings for MyModule.
type MyModuleConfig struct {
    Enabled bool   `mapstructure:"enabled"`
    Param1  string `mapstructure:"param1"`
    Param2  int    `mapstructure:"param2"`
}

type Scheme struct {
    Env      string          `mapstructure:"env"`
    // ... other configs
    MyModule *MyModuleConfig `mapstructure:"mymodule"`
}
```

Add defaults in `config/init.go`:

```go
func setDefaults() {
    // ... other defaults
    viper.SetDefault("mymodule.enabled", false)
    viper.SetDefault("mymodule.param1", "default_value")
    viper.SetDefault("mymodule.param2", 100)
}
```

### Step 2: Implement Module Interface

Create `internal/mymodule/module.go`:

```go
package mymodule

import (
    "context"
    "fmt"
    
    "microservice-template/config"
    "microservice-template/pkg/logger"
)

// Module implements module.Module interface for MyModule.
type Module struct {
    config *config.MyModuleConfig
    // Add your dependencies here
}

// NewModule creates a new MyModule instance.
func NewModule(cfg *config.MyModuleConfig) *Module {
    return &Module{
        config: cfg,
    }
}

// Name returns the module identifier.
func (m *Module) Name() string {
    return "mymodule"
}

// Init initializes the module resources.
func (m *Module) Init(ctx context.Context) error {
    logger.Log().Infof("initializing %s module", m.Name())
    
    // Initialize your resources here (connect, prepare, etc.)
    // Example:
    // if err := m.connect(); err != nil {
    //     return fmt.Errorf("connect: %w", err)
    // }
    
    return nil
}

// Start begins module operation.
func (m *Module) Start(ctx context.Context) error {
    logger.Log().Infof("starting %s module", m.Name())
    
    // Start your background workers, servers, etc.
    // Use goroutines for long-running operations
    // Example:
    // go m.runWorker(ctx)
    
    return nil
}

// Stop gracefully shuts down the module.
func (m *Module) Stop(ctx context.Context) error {
    logger.Log().Infof("stopping %s module", m.Name())
    
    // Cleanup resources, close connections, etc.
    // Respect context timeout for graceful shutdown
    
    return nil
}

// HealthCheck returns module health status.
func (m *Module) HealthCheck(ctx context.Context) error {
    // Check if module is healthy
    // Return nil if healthy, error otherwise
    // Keep this quick (< 2 seconds)
    
    return nil
}
```

### Step 3: Register Module in Application

In `internal/application.go`, add to `registerModules()`:

```go
func (app *App) registerModules() error {
    // Register MyModule if enabled
    if app.config.MyModule != nil && app.config.MyModule.Enabled {
        myMod := mymodule.NewModule(app.config.MyModule)
        app.modules.Register(myMod)
    }
    
    logger.Log().Infof("registered %d modules", app.modules.Count())
    return nil
}
```

### Step 4: Handle Module Dependencies

If your module depends on another module, inject dependencies via constructor:

```go
// Module B depends on Module A
func NewModuleB(cfg *config.ModuleBConfig, moduleA *modulea.Module) *ModuleB {
    return &ModuleB{
        config: cfg,
        depA:   moduleA,
    }
}
```

Register in dependency order in `application.go`:

```go
func (app *App) registerModules() error {
    var modA *modulea.Module
    
    // Register Module A first (dependency)
    if app.config.ModuleA != nil && app.config.ModuleA.Enabled {
        modA = modulea.NewModule(app.config.ModuleA)
        app.modules.Register(modA)
    }
    
    // Register Module B (depends on A)
    if modA != nil && app.config.ModuleB != nil && app.config.ModuleB.Enabled {
        modB := moduleb.NewModule(app.config.ModuleB, modA)
        app.modules.Register(modB)
    }
    
    return nil
}
```

## Best Practices

### 1. Idempotent Init
Make `Init()` safe to call multiple times:

```go
func (m *Module) Init(ctx context.Context) error {
    if m.initialized {
        return nil // Already initialized
    }
    
    // Do initialization work
    
    m.initialized = true
    return nil
}
```

### 2. Non-blocking Start
Use goroutines for long-running operations:

```go
func (m *Module) Start(ctx context.Context) error {
    // Start background worker
    go m.runWorker(ctx)
    
    // Return immediately - don't block
    return nil
}
```

### 3. Graceful Stop
Respect context timeout for graceful shutdown:

```go
func (m *Module) Stop(ctx context.Context) error {
    // Signal workers to stop
    close(m.stopChan)
    
    // Wait for workers with timeout
    select {
    case <-m.doneChan:
        logger.Log().Info("module stopped gracefully")
    case <-ctx.Done():
        logger.Log().Warn("module stop timed out")
        return ctx.Err()
    }
    
    return nil
}
```

### 4. Fast HealthCheck
Keep health checks under 2 seconds:

```go
func (m *Module) HealthCheck(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    // Quick health check
    if err := m.ping(ctx); err != nil {
        return fmt.Errorf("unhealthy: %w", err)
    }
    
    return nil
}
```

### 5. Error Wrapping
Use `fmt.Errorf` with `%w` for error context:

```go
func (m *Module) Init(ctx context.Context) error {
    if err := m.connect(); err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }
    return nil
}
```

### 6. Structured Logging
Log all lifecycle events:

```go
logger.Log().Infof("module %s initialized successfully", m.Name())
logger.Log().Errorf("module %s failed to start: %v", m.Name(), err)
```

## Testing Modules

Create `internal/mymodule/module_test.go`:

```go
package mymodule

import (
    "context"
    "testing"
    
    "microservice-template/config"
)

func TestModule_Lifecycle(t *testing.T) {
    cfg := &config.MyModuleConfig{
        Enabled: true,
        Param1:  "test",
    }
    
    mod := NewModule(cfg)
    ctx := context.Background()
    
    // Test Init
    if err := mod.Init(ctx); err != nil {
        t.Fatalf("Init failed: %v", err)
    }
    
    // Test Start
    if err := mod.Start(ctx); err != nil {
        t.Fatalf("Start failed: %v", err)
    }
    
    // Test HealthCheck
    if err := mod.HealthCheck(ctx); err != nil {
        t.Errorf("HealthCheck failed: %v", err)
    }
    
    // Test Stop
    if err := mod.Stop(ctx); err != nil {
        t.Errorf("Stop failed: %v", err)
    }
}

func TestModule_Name(t *testing.T) {
    mod := NewModule(&config.MyModuleConfig{})
    
    if mod.Name() != "mymodule" {
        t.Errorf("expected name 'mymodule', got '%s'", mod.Name())
    }
}
```

## Module Registration Order

Modules are initialized, started, and stopped in the order they are registered. Plan your registration order based on dependencies:

1. **Infrastructure** (database, cache, queue) - no dependencies
2. **Business Logic** (repositories, services) - depend on infrastructure
3. **Transport** (HTTP, gRPC) - depend on business logic

**Important**: Stop happens in **reverse order** (LIFO) to ensure proper cleanup.

## Example: Database Module

See the following structure for a complete database module example:

```
internal/
├── database/
│   ├── module.go          # Implements module.Module
│   ├── postgres.go        # PostgreSQL-specific code
│   └── repository/        # Repository interfaces
│       └── user.go
├── models/                # Domain models
│   └── user.go
└── repository/            # Repository implementations
    └── postgres/
        └── user.go
```

## Troubleshooting

### Module not starting
- Check if module is enabled in config
- Verify module is registered in `application.go`
- Check logs for initialization errors

### Module fails health check
- Ensure dependencies are available
- Check resource connectivity (database, cache, etc.)
- Verify timeout is sufficient (< 2s)

### Shutdown hangs
- Check if `Stop()` respects context timeout
- Ensure goroutines are properly cleaned up
- Look for blocking operations in `Stop()`

## Additional Resources

- Module interface: `internal/module/module.go`
- Module manager: `internal/module/manager.go`
- Configuration schema: `config/scheme.go`
- Example config: `config.example.yaml`
