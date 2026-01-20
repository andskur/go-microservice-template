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

### Step 1: Define Module Configuration (when needed)

Only add configuration when the module is configurable. Some modules (like the service layer) may not need a dedicated config; they can be always-on and rely on optional dependencies.

Example with configuration in `config/scheme.go`:

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

Add defaults in `config/init.go` when config exists:

```go
func setDefaults() {
    // ... other defaults
    viper.SetDefault("mymodule.enabled", false)
    viper.SetDefault("mymodule.param1", "default_value")
    viper.SetDefault("mymodule.param2", 100)
}
```

If a module is always on and has no config (like the service module), skip the config struct and defaults, and document its optional dependencies instead.

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

In `internal/application.go`, add to `registerModules()` in dependency order. Example with optional dependency injection:

```go
func (app *App) registerModules() error {
    // Infrastructure first: repository (enabled only when database is enabled)
    var repoMod *repository.Module
    if app.config.Database != nil && app.config.Database.Enabled {
        repoMod = repository.NewModule(app.config.Database)
        app.modules.Register(repoMod)
    }

    // Business logic: service module always registers; repository is optional
    var repo repository.IRepository
    if repoMod != nil {
        repo = repoMod.Repository()
    }
    svcMod := service.NewModule(repo)
    app.modules.Register(svcMod)

    // Add more modules (HTTP, gRPC, queues) after business logic

    logger.Log().Infof("registered %d modules", app.modules.Count())
    return nil
}
```

Service module guidance:
- Service is always registered.
- Dependencies (repository, cache, events, etc.) are injected explicitly and may be nil.
- Service methods should handle missing dependencies gracefully (return clear errors).

### Step 4: Handle Module Dependencies

Inject dependencies explicitly via constructors; avoid global lookups. Dependencies may be optional—if so, accept nil and handle gracefully.

```go
// Module B depends on Module A
func NewModuleB(cfg *config.ModuleBConfig, moduleA *modulea.Module) *ModuleB {
    return &ModuleB{
        config: cfg,
        depA:   moduleA, // can be nil if optional
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

    // Register Module B (depends on A); pass nil if A not enabled
    var depA *modulea.Module
    if modA != nil {
        depA = modA
    }
    modB := moduleb.NewModule(app.config.ModuleB, depA)
    app.modules.Register(modB)

    return nil
}
```

Guidance:
- Keep constructor injection explicit and typed.
- If a dependency is optional, document the behavior when nil (e.g., service returns `repository not available` errors when DB is disabled).
- Avoid service locators or global registries; pass what you need.

## Best Practices

### 1. Idempotent Init
Make `Init()` safe to call multiple times:

```go
func (m *Module) Init(ctx context.Context) error {
    if m.initialized {
        return nil // Already initialized
    }
```

### 2. Optional Dependencies
- Accept optional dependencies as constructor params; allow nil.
- Clearly document and handle behavior when a dependency is missing (return explicit errors, not panics).
- Example: the service module always registers; when repository is nil (database disabled), methods return `repository not available` errors.

### 3. Dependency Order
- Register infrastructure before business logic; business logic before transports.
- Stop happens in reverse order automatically (LIFO) via the manager.

### 4. Models & Database Integration

**Model Structure:**
- Models live in `internal/models/` with go-pg struct tags for database mapping (e.g., `pg:"column_name,pk"`).
- Use go-pg hooks for lifecycle: `BeforeInsert`, `BeforeUpdate`, `AfterSelect`.
- Keep validation in `Validate()`; use hooks for DB-specific conversions.

**Status Enums with Database:**
- Use dual fields: `Status UserStatus pg:"-"` (enum, not stored) and `StatusSQL string pg:"status,use_zero"` (string, stored).
- Convert in hooks: `BeforeInsert`/`BeforeUpdate` set `StatusSQL = Status.String()`; `AfterSelect` parses `StatusSQL` back to enum.
- Validation uses enum; database uses string; hooks keep them in sync.

**Example User Model with go-pg:**
```go
type User struct {
    tableName struct{} `pg:"users,discard_unknown_columns"` //nolint:unused

    UUID      uuid.UUID  `pg:"uuid,pk,type:uuid"`
    Status    UserStatus `pg:"-"`
    StatusSQL string     `pg:"status,use_zero"`
    Email     string     `pg:"email,unique,notnull"`
    Name      string     `pg:"name,notnull"`
    CreatedAt time.Time  `pg:"created_at,notnull,default:now()"`
    UpdatedAt time.Time  `pg:"updated_at,notnull,default:now()"`
}

func (u *User) BeforeInsert(ctx context.Context) (context.Context, error) {
    if u.UUID == uuid.Nil {
        id, err := uuid.NewV4()
        if err != nil {
            return ctx, fmt.Errorf("generate UUID: %w", err)
        }
        u.UUID = id
    }

    status := u.Status.String()
    if status == "" {
        return ctx, fmt.Errorf("invalid status value: %d", u.Status)
    }
    u.StatusSQL = status

    return ctx, nil
}

func (u *User) BeforeUpdate(ctx context.Context) (context.Context, error) {
    status := u.Status.String()
    if status == "" {
        return ctx, fmt.Errorf("invalid status value: %d", u.Status)
    }
    u.StatusSQL = status
    u.UpdatedAt = time.Now()

    return ctx, nil
}

func (u *User) AfterSelect(_ context.Context) error {
    status, err := UserStatusFromString(u.StatusSQL)
    if err != nil {
        return fmt.Errorf("parse user status: %w", err)
    }
    u.Status = status
    return nil
}
```

**Repository Implementation (go-pg):**
```go
// internal/repository/postgres.go
func (r *PostgresRepository) CreateUser(user *models.User) error {
    if _, err := r.db.Model(user).Returning("*").Insert(); err != nil {
        return fmt.Errorf("insert user %s into db: %w", user.Email, err)
    }
    return nil
}

func (r *PostgresRepository) UserBy(user *models.User, getter UserGetter) error {
    query := r.db.Model(user).Column("user.*")
    if err := getter.Get(query, user); err != nil {
        return fmt.Errorf("parse getter: %w", err)
    }
    if err := query.Select(); err != nil {
        return fmt.Errorf("get user from database by %s: %w", getter.String(), err)
    }
    return nil
}
```

**Key Patterns:**
- Use `Returning("*")` on inserts to populate DB defaults back into the model.
- Use `Column("table.*")` to select all columns explicitly.
- Apply getters via `WherePK()` or `Where()` for flexible queries.
- Keep validation in `Validate()`; hooks handle DB string/enum conversions.


### 5. Fast Health Checks
- Keep `HealthCheck` under 2s; avoid blocking operations.

### 6. No Global Service Locator
- Do not fetch dependencies from globals; use explicit constructor injection.

### 7. Graceful Shutdown
- `Stop` should be idempotent, respect context deadlines, and clean up all resources.
    
    // Do initialization work
    
    m.initialized = true
    return nil
}
```

### 8. Non-blocking Start
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

## gRPC Module Patterns

- Module path: `internal/grpc/` (implements `module.Module`).
- Configuration: `config.GRPCConfig` (`grpc.*` keys) with defaults in `config/init.go`.
- Registration: optional, enabled when `grpc.enabled=true` in config; wired in `internal/application.go` after service module.
- Health: standard `grpc.health.v1` service registered in `Server.RegisterHealthService()`.
- Middleware: logging and recovery interceptors (no Sentry).
- Handler registration: add your handlers in `module.go -> registerHandlers()` (currently empty; populate after generating protos).
- Conversions: add proto helpers under `internal/grpc/` (keeps models package free of proto deps).
- Protocols: pull from the shared repo (`https://github.com/andskur/protocols-template.git`) via subtree; no bundled example is kept locally.

### Adding a New gRPC Service
1. Pull/update protocols: `make proto-setup` / `make proto-update` (subtree from protocols-template).
2. Generate code (Buf recommended): `make buf-generate PROTO_PACKAGE=<service>` or use `make proto-generate PROTO_PACKAGE=<service>`.
3. Add conversion helpers in `internal/grpc/` for your types and enums.
4. Implement handlers in `internal/grpc/` using `service.IService` (or other deps); return gRPC status errors.
5. Register handlers in `module.go -> registerHandlers()`.
6. Keep HealthCheck fast (<2s); server already registers standard health service.

## HTTP Module Patterns

The HTTP module provides a REST API using go-swagger for spec-first development with Swagger 2.0.

- Module path: `internal/http/` (implements `module.Module`).
- Configuration: `config.HTTPConfig` (`http.*` keys) with defaults in `config/init.go`.
- Registration: optional, enabled when `http.enabled=true` in config; wired in `internal/application.go` after service module.
- API Specification: `api/swagger.yaml` (Swagger 2.0 format); server code generated to `internal/http/server/` (gitignored).
- Handlers: implemented in `internal/http/handlers/` as structs with dependencies and `Handle()` method.
- Formatters: `internal/http/formatter/` converts domain models ↔ API models (generated from swagger).
- Middleware: Recovery → Logger → CORS → RateLimit chain using `justinas/alice`.
- Authentication: JWT validation in `internal/http/auth/auth.go`; mock mode for development with `http.mock_auth=true`.
- Health endpoint: `GET /health` checks all module health; returns 200 OK with status JSON.

### Swagger Workflow

```bash
# Install go-swagger (one-time)
make swagger-install

# Edit api/swagger.yaml to define your endpoints

# Validate specification
make swagger-validate

# Generate server code (creates internal/http/server/ and internal/http/models/)
make generate-api

# Clean generated code
make swagger-clean
```

**Important**: Use Swagger 2.0 format (not OpenAPI 3.0); go-swagger doesn't support OpenAPI 3.0 yet.

### Adding a New HTTP Endpoint

**Step 1: Define endpoint in `api/swagger.yaml`**

```yaml
paths:
  /widgets/{id}:
    get:
      summary: Get widget by ID
      operationId: getWidget
      parameters:
        - name: id
          in: path
          required: true
          type: string
          format: uuid
      responses:
        200:
          description: Widget found
          schema:
            $ref: '#/definitions/Widget'
        404:
          description: Widget not found
          schema:
            $ref: '#/definitions/Error'
      security:
        - Bearer: []

definitions:
  Widget:
    type: object
    required:
      - id
      - name
    properties:
      id:
        type: string
        format: uuid
      name:
        type: string
      status:
        type: string
        enum: [active, inactive]
```

**Step 2: Generate server code**

```bash
make generate-api
```

**Step 3: Create handler in `internal/http/handlers/widgets.go`**

```go
package handlers

import (
    "errors"
    "net/http"

    "github.com/go-openapi/runtime"
    "github.com/gofrs/uuid"
    "github.com/sirupsen/logrus"

    "microservice-template/internal/http/formatter"
    "microservice-template/internal/http/models"
    "microservice-template/internal/service"
)

// GetWidgetHandler handles GET /widgets/{id} requests.
type GetWidgetHandler struct {
    service service.IService
    log     *logrus.Logger
}

// NewGetWidgetHandler creates a new GetWidgetHandler.
func NewGetWidgetHandler(svc service.IService, log *logrus.Logger) *GetWidgetHandler {
    return &GetWidgetHandler{
        service: svc,
        log:     log,
    }
}

// Handle processes the request.
func (h *GetWidgetHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // Extract path parameter
    vars := mux.Vars(r)
    idStr := vars["id"]
    
    // Parse UUID
    id, err := uuid.FromString(idStr)
    if err != nil {
        DefaultError(w, http.StatusBadRequest, "Invalid ID format")
        return
    }
    
    // Call service
    widget, err := h.service.GetWidgetByID(r.Context(), id)
    if err != nil {
        // Map service errors to HTTP status codes
        if errors.Is(err, service.ErrNotFound) {
            DefaultError(w, http.StatusNotFound, "Widget not found")
            return
        }
        if errors.Is(err, service.ErrRepositoryUnavailable) {
            DefaultError(w, http.StatusServiceUnavailable, "Service temporarily unavailable")
            return
        }
        h.log.Errorf("get widget by id: %v", err)
        DefaultError(w, http.StatusInternalServerError, "Internal server error")
        return
    }
    
    // Convert domain model to API model
    apiWidget := formatter.WidgetToAPI(widget)
    
    // Return success response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(apiWidget); err != nil {
        h.log.Errorf("encode response: %v", err)
    }
}
```

**Step 4: Create formatter in `internal/http/formatter/widget.go`**

```go
package formatter

import (
    domainModels "microservice-template/internal/models"
    apiModels "microservice-template/internal/http/models"
)

// WidgetToAPI converts domain Widget to API Widget.
func WidgetToAPI(widget *domainModels.Widget) *apiModels.Widget {
    if widget == nil {
        return nil
    }
    
    return &apiModels.Widget{
        ID:     widget.ID.String(),
        Name:   widget.Name,
        Status: widget.Status.String(),
    }
}

// WidgetFromAPI converts API Widget to domain Widget.
func WidgetFromAPI(apiWidget *apiModels.Widget) (*domainModels.Widget, error) {
    if apiWidget == nil {
        return nil, nil
    }
    
    id, err := uuid.FromString(apiWidget.ID)
    if err != nil {
        return nil, fmt.Errorf("parse id: %w", err)
    }
    
    status, err := domainModels.WidgetStatusFromString(apiWidget.Status)
    if err != nil {
        return nil, fmt.Errorf("parse status: %w", err)
    }
    
    return &domainModels.Widget{
        ID:     id,
        Name:   apiWidget.Name,
        Status: status,
    }, nil
}
```

**Step 5: Register handler in `internal/http/module.go`**

```go
func (m *Module) setupRoutes() http.Handler {
    // ... existing middleware setup ...
    
    // Create handlers
    getUserHandler := handlers.NewGetUserHandler(m.service, m.log)
    getWidgetHandler := handlers.NewGetWidgetHandler(m.service, m.log)
    healthHandler := handlers.NewHealthHandler(m.manager, m.log)
    
    // Setup router
    router := mux.NewRouter()
    
    // Public routes
    router.HandleFunc("/health", healthHandler.Handle).Methods(http.MethodGet)
    
    // Protected routes (require JWT)
    protected := router.PathPrefix("").Subrouter()
    protected.Use(m.auth.Authenticate)
    protected.HandleFunc("/users", getUserHandler.Handle).Methods(http.MethodGet)
    protected.HandleFunc("/widgets/{id}", getWidgetHandler.Handle).Methods(http.MethodGet)
    
    // Apply middleware chain
    return m.middleware.Then(router)
}
```

**Step 6: Test the handler in `internal/http/handlers/widgets_test.go`**

```go
package handlers

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gofrs/uuid"
    "github.com/gorilla/mux"

    "microservice-template/internal/models"
    "microservice-template/internal/service"
    "microservice-template/pkg/logger"
)

func TestGetWidgetHandler_Success(t *testing.T) {
    widgetID := uuid.Must(uuid.NewV4())
    
    mockSvc := &MockService{
        GetWidgetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Widget, error) {
            return &models.Widget{
                ID:     widgetID,
                Name:   "Test Widget",
                Status: models.WidgetActive,
            }, nil
        },
    }
    
    handler := NewGetWidgetHandler(mockSvc, logger.Log())
    
    req := httptest.NewRequest(http.MethodGet, "/widgets/"+widgetID.String(), nil)
    req = mux.SetURLVars(req, map[string]string{"id": widgetID.String()})
    w := httptest.NewRecorder()
    
    handler.Handle(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }
}

func TestGetWidgetHandler_NotFound(t *testing.T) {
    mockSvc := &MockService{
        GetWidgetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Widget, error) {
            return nil, service.ErrNotFound
        },
    }
    
    handler := NewGetWidgetHandler(mockSvc, logger.Log())
    
    req := httptest.NewRequest(http.MethodGet, "/widgets/"+uuid.Must(uuid.NewV4()).String(), nil)
    w := httptest.NewRecorder()
    
    handler.Handle(w, req)
    
    if w.Code != http.StatusNotFound {
        t.Errorf("expected status 404, got %d", w.Code)
    }
}
```

### HTTP Error Mapping Pattern

Map service errors to HTTP status codes consistently:

```go
// Service error → HTTP status mapping
if err != nil {
    switch {
    case errors.Is(err, service.ErrNotFound):
        DefaultError(w, http.StatusNotFound, "Resource not found")
    case errors.Is(err, service.ErrInvalidInput):
        DefaultError(w, http.StatusBadRequest, "Invalid input")
    case errors.Is(err, service.ErrRepositoryUnavailable):
        DefaultError(w, http.StatusServiceUnavailable, "Service temporarily unavailable")
    default:
        h.log.Errorf("unexpected error: %v", err)
        DefaultError(w, http.StatusInternalServerError, "Internal server error")
    }
    return
}
```

### HTTP Authentication Pattern

**Production JWT validation:**
```go
// internal/http/auth/auth.go implements JWT validation
// Protected routes get user context from auth middleware
userID := auth.GetUserID(r)
email := auth.GetEmail(r)
isAdmin := auth.IsAdmin(r)
```

**Development mock mode:**
```bash
# Bypass JWT validation for local testing
export HTTP_MOCK_AUTH=true
curl -H "Authorization: Bearer any-token" http://localhost:8080/users
```

**Gatekeeper integration:**
See detailed TODO in `internal/http/auth/auth.go` for external auth service integration steps.

### HTTP Testing Patterns

**Mock service for handler tests:**
```go
type MockService struct {
    GetWidgetByIDFunc func(ctx context.Context, id uuid.UUID) (*models.Widget, error)
}

func (m *MockService) GetWidgetByID(ctx context.Context, id uuid.UUID) (*models.Widget, error) {
    if m.GetWidgetByIDFunc != nil {
        return m.GetWidgetByIDFunc(ctx, id)
    }
    return nil, errors.New("not implemented")
}
```

**Table-driven handler tests:**
```go
func TestGetWidgetHandler(t *testing.T) {
    tests := []struct {
        name           string
        widgetID       string
        mockFunc       func(ctx context.Context, id uuid.UUID) (*models.Widget, error)
        expectedStatus int
    }{
        {
            name:     "success",
            widgetID: uuid.Must(uuid.NewV4()).String(),
            mockFunc: func(ctx context.Context, id uuid.UUID) (*models.Widget, error) {
                return &models.Widget{ID: id, Name: "Test"}, nil
            },
            expectedStatus: http.StatusOK,
        },
        {
            name:     "not found",
            widgetID: uuid.Must(uuid.NewV4()).String(),
            mockFunc: func(ctx context.Context, id uuid.UUID) (*models.Widget, error) {
                return nil, service.ErrNotFound
            },
            expectedStatus: http.StatusNotFound,
        },
        {
            name:           "invalid uuid",
            widgetID:       "not-a-uuid",
            expectedStatus: http.StatusBadRequest,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... test implementation
        })
    }
}
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
- Configuration defaults: `config/init.go`
