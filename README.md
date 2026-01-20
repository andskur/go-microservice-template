# go-microservice-template

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](https://go.dev/)

A minimal Go microservice template with Cobra/Viper CLI wiring, ldflags-driven versioning, logrus logging, Makefile targets, tests, and GitHub Actions CI/CD (lint/test/build on PRs/main plus auto-tagged releases on `main`). The structure is intentionally simple so you can plug in your runtime workloads quickly.

## Quickstart
- Requirements: Go 1.21+ (module sets 1.23/1.24), GNU `make`.
- Clone and create your branch: `git checkout -b feature/your-branch`.
- Build: `make build` (binary `./microservice-template`).
- Run: `make run` (invokes `go run -race cmd/microservice-template.go serve`).
- Version: `./microservice-template --version`.
- Lint: `make lint` (golangci-lint).
- Test: `make test` or single test `go test ./... -run TestName -count=1`.
- Generate + test: `make test-with-gen` (runs proto + swagger generation first).
- Generate + lint: `make lint-with-gen` (runs proto + swagger generation first).
- gRPC tests: `make test-grpc` (runs gRPC package including integration).
- HTTP tests: `make test-http` (runs HTTP package tests).
- Coverage: `make test-coverage` (writes `coverage.out`).
- Tidy deps: `make tidy`; update deps: `make update`.
- HTTP quickstart: see [docs/HTTP_SWAGGER_GUIDE.md](./docs/HTTP_SWAGGER_GUIDE.md); enable with `HTTP_ENABLED=true`, generate API with `make generate-api`, test with curl.
- gRPC quickstart: see [docs/GRPC_GUIDE.md](./docs/GRPC_GUIDE.md); enable with `GRPC_ENABLED=true`, test with grpcurl; use shared protocols from `https://github.com/andskur/protocols-template.git`.

### Renaming the project
- Command: `make rename NEW_NAME=my-service` (required parameter).
- Valid NEW_NAME: lowercase letters, numbers, hyphens, optional `/` segments (e.g., `my-service`, `github.com/yourorg/my-service`).
- Updates: module path and imports, Makefile vars, entrypoint file, Cobra root `Use`, swagger API struct name, Dockerfile binary, README/AGENTS references, optional git remote.
- After rename regenerate generated code: `make generate-all` (or at least `make generate-api` + proto generation as needed).
- Verify after rename: `go test ./...`, `make build`, `./<new-binary> --version`.

## Features
- Simple, small footprint using standard libs plus Cobra/Viper/logrus.
- **Module system** for optional components (repository, service, HTTP, gRPC, queue, etc.).
- **HTTP REST API** with Swagger/OpenAPI spec-first approach using go-swagger.
- **gRPC API** with protocol buffer code generation and health checks.
- Version metadata injected via ldflags (`pkg/version`).
- Structured logging via `pkg/logger` singleton.
- Makefile targets for build/run/lint/test/tidy/update/generate-api.
- CI pipeline: lint/test/build on PRs and `main`; release pipeline auto-tags on `main` and publishes a GitHub release (source-only).
- Tests included for CLI wiring, config defaults, versioning, logger singleton, helpers, HTTP and gRPC modules.
- Rename-friendly: single placeholder name with automated `make rename` target.

## Project Structure
```
go-microservice-template/
├── api/                        # Swagger/OpenAPI specifications
├── cmd/                        # CLI entry + commands
├── config/                     # Viper defaults and scheme
├── db/migrations/              # Database migration files (golang-migrate)
├── docs/                       # Additional guides (HTTP_SWAGGER_GUIDE, GRPC_GUIDE)
├── internal/
│   ├── application.go          # App wiring + module registration
│   ├── grpc/                   # gRPC module (server, interceptors)
│   ├── http/                   # HTTP module (handlers, middleware, auth)
│   ├── module/                 # Module interface/manager
│   ├── repository/             # Repository module (optional)
│   ├── service/                # Business logic module
│   └── models/                 # Domain models/enums
├── pkg/                        # Reusable packages (logger, version)
├── protocols/                  # Protocol definitions pulled via subtree (no bundled example)
├── scripts/                    # Automation scripts (rename)
├── .github/workflows/          # CI/CD pipelines
├── Dockerfile                  # Multi-stage container build
├── docker-compose.yml          # Local stack (Postgres/Redis/app)
├── Makefile                    # Build/run/lint/test/proto/swagger targets
├── README.md, AGENTS.md        # Docs and guidelines
└── go.mod, go.sum              # Dependencies
```

## Module System


This template uses a **module-based architecture** for optional components. Modules provide a standard lifecycle (Init → Start → Stop) and can be enabled/disabled via configuration.

### Available Module Slots

The template includes configuration placeholders for common modules:

| Module | Purpose | Config Key | Status |
|--------|---------|-----------|--------|
| Repository | Database-backed persistence (wraps DB connection) | `database` | ✅ Implemented (enabled when `database.enabled` is true) |
| Service | Business logic orchestrator (optional deps) | n/a | ✅ Implemented (always registered; repository optional) |
| HTTP | HTTP REST API server with Swagger/OpenAPI | `http` | ✅ Implemented (enabled when `http.enabled` is true) |
| gRPC Server | gRPC API server | `grpc` | ✅ Implemented (enabled when `grpc.enabled` is true) |
| gRPC Client | External service client for microservice communication | `grpc_client` | ✅ Implemented (enabled when `grpc_client.enabled` is true) |

### Enabling Modules

Configuration can come from env vars (recommended) or a config file (`config.yaml` is optional). Viper merges: flags > env vars > config file.

**Env example (preferred):**
```bash
export DATABASE_ENABLED=true
export DATABASE_DRIVER=postgres
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
```

**config.yaml example (optional):**
```yaml
database:
  enabled: true
  driver: postgres
  host: localhost
  port: 5432
  # ... other settings
```

- The repository module registers only when `database.enabled` (or `DATABASE_ENABLED`) is `true`.
- The service module always registers; if no repository is available, database operations return clear errors.

See `config/scheme.go` for configuration structure definitions.

### Database Setup

The repository module requires PostgreSQL when enabled.

**Quick start with Docker:**
```bash
docker run --name postgres-dev \
  -e POSTGRES_USER=dev \
  -e POSTGRES_PASSWORD=dev \
  -e POSTGRES_DB=microservice_dev \
  -p 5432:5432 \
  -d postgres:16-alpine
```

**Install migration tool (one-time):**
```bash
make migrate-install
```

**Run migrations:**
```bash
# Apply all pending migrations
make migrate-up

# Check current migration version
make migrate-version
```

**Available migration targets:**
```bash
make migrate-install      # Install golang-migrate CLI
make migrate-create       # Create new migration (requires NAME=)
make migrate-up           # Apply all pending migrations
make migrate-down         # Rollback last migration
make migrate-force        # Force migration version (requires VERSION=)
make migrate-version      # Show current migration version
make migrate-drop         # Drop all tables (⚠️ DANGER - requires confirmation)
```

**Local development with Docker Compose:**
```bash
# Start Postgres, Redis, and auto-run migrations (uses db/migrations)
make compose-up

# Stop services
make compose-down

# Restart services
make compose-restart
```

For production deployments, run migrations before starting the application or use a separate migration job in your deployment pipeline.

### HTTP REST API Setup

The HTTP module provides a REST API with Swagger/OpenAPI specification support using go-swagger.

**Install go-swagger (one-time):**
```bash
make swagger-install
```

**Generate API server code from spec:**
```bash
# Validate the swagger spec
make swagger-validate

# Generate server code from api/swagger.yaml
make generate-api
```

**Enable HTTP module:**
```bash
export HTTP_ENABLED=true
export HTTP_PORT=8080
export HTTP_MOCK_AUTH=true  # For local development (bypasses JWT validation)

# Run the service
make run
```

**Test the HTTP endpoints:**
```bash
# Health check (public endpoint)
curl http://localhost:8080/health

# Get user by email (requires auth in production, mock mode for dev)
curl -H "Authorization: Bearer test-token" \
  "http://localhost:8080/users?email=test@example.com"
```

**Available HTTP targets:**
```bash
make swagger-install      # Install go-swagger CLI
make swagger-validate     # Validate swagger spec
make generate-api         # Generate server code from api/swagger.yaml
make swagger-clean        # Remove generated code
make test-http           # Run HTTP module tests
```

**Configuration options:**
- `http.enabled` - Enable/disable HTTP server (default: false)
- `http.host` - Server host (default: localhost)
- `http.port` - Server port (default: 8080)
- `http.mock_auth` - Use mock authentication for development (default: false)
- `http.cors.enabled` - Enable CORS (default: true)
- `http.rate_limit.enabled` - Enable rate limiting (default: true)
- `http.rate_limit.requests_per_second` - Rate limit (default: 100)

For detailed HTTP development guide including adding new endpoints, authentication, and middleware, see [docs/HTTP_SWAGGER_GUIDE.md](./docs/HTTP_SWAGGER_GUIDE.md).

### gRPC Client Setup

The gRPC client module enables communication with external gRPC microservices. HTTP handlers use the client to fetch data from external services, making this template ideal for gateway/BFF (Backend for Frontend) patterns.

**Architecture Overview:**
- **Hybrid approach**: Pure gRPC client in `pkg/userservice/` (proto types only) + module wrapper in `internal/grpcclient/` (conversions + lifecycle)
- **HTTP integration**: HTTP handlers receive grpcClient as dependency and use it to fetch from external services
- **Service independence**: Service layer remains focused on local business logic only

**Enable gRPC client:**
```bash
export GRPC_CLIENT_ENABLED=true
export GRPC_CLIENT_ADDRESS="user-service:9090"
export GRPC_CLIENT_TIMEOUT="30s"

# Run the service
make run
```

**Configuration options:**
- `grpc_client.enabled` - Enable/disable client (default: false)
- `grpc_client.address` - External service address (default: "localhost:9090")
- `grpc_client.timeout` - Request timeout (default: "30s")
- `grpc_client.keep_alive.time` - Keep-alive ping interval (default: "10s")
- `grpc_client.keep_alive.timeout` - Keep-alive timeout (default: "1s")
- `grpc_client.keep_alive.permit_without_stream` - Send pings without streams (default: true)

**Handler integration patterns:**

The template demonstrates fetching from external service only. Alternative patterns are documented in handler code:

```go
// Pattern 1: External only (current implementation)
user, err := h.grpcClient.GetUserByEmail(ctx, email)

// Pattern 2: Local database only (commented alternative)
// user, err := h.service.GetUserByEmail(ctx, email)

// Pattern 3: External with local fallback (commented alternative)
// user, err := h.grpcClient.GetUserByEmail(ctx, email)
// if err != nil {
//     user, err = h.service.GetUserByEmail(ctx, email)
// }

// Pattern 4: Aggregate from both sources (commented alternative)
// externalUser, _ := h.grpcClient.GetUserByEmail(ctx, email)
// localUser, _ := h.service.GetUserByEmail(ctx, email)
// user = mergeUsers(externalUser, localUser)
```

**Test endpoints:**
```bash
# Requires external user-service running on configured address
export HTTP_ENABLED=true
export HTTP_MOCK_AUTH=true
export GRPC_CLIENT_ENABLED=true
export GRPC_CLIENT_ADDRESS="user-service:9090"

make run

# Test the endpoint
curl -H "Authorization: Bearer test-token" \
  "http://localhost:8080/users?email=test@example.com"
```

**Error handling:**
- Returns **503 Service Unavailable** when grpcClient is not configured
- Maps gRPC errors: `not found` → 404, `invalid input` → 400, `unavailable` → 503

For detailed gRPC development guide including adding new services and proto definitions, see [docs/GRPC_GUIDE.md](./docs/GRPC_GUIDE.md).

### Adding Custom Modules

See [Module Development Guide](./docs/MODULE_DEVELOPMENT.md) for creating custom modules. The module system provides:

- **Standard lifecycle**: Init → Start → Stop with health checks
- **Dependency injection**: Modules depend on each other via constructor injection (explicit; no service locator)
- **Configuration-driven**: Enable/disable modules via YAML/env vars (repository depends on `database.enabled`; service always registers)
- **Graceful shutdown**: Automatic cleanup in reverse registration order

## CLI
- Root command name: `microservice-template`.
- Subcommands: `serve` (current runtime hook). Add more via `cmd/<name>` and register on root.
- Version output: `./microservice-template --version` (ldflags populate `pkg/version`).
- `serve` lifecycle: `PreRun` logs version; `RunE` should start your workloads; `PostRun` always stops app.
- Adding a new command (example):
  ```go
  // cmd/health/health.go
  package health

  import "github.com/spf13/cobra"

  func Cmd() *cobra.Command {
      return &cobra.Command{
          Use:   "health",
          Short: "Health probe",
          RunE: func(_ *cobra.Command, _ []string) error {
              // add checks here
              return nil
          },
      }
  }
  ```
  Register it in `cmd/microservice-template.go`: `rootCmd.AddCommand(health.Cmd())`.

## Models & Enums
- Location: `internal/models` with go-pg struct tags/hooks for database integration.
- Validation: implement `Validate() error` and return `*models.ValidationError` (`Field`, `Message`) for structured errors.
- Enums: typed ints with `String()` and case-insensitive `UserStatusFromString()`; add proto/JWT conversions later if needed.
- Hooks: `BeforeInsert`/`BeforeUpdate` convert enums to strings and ensure UUID/timestamps; `AfterSelect` converts strings back to enums.

### Example: User Model
```go
// internal/models/user.go
user := &models.User{
    Email:  "test@example.com",
    Name:   "Jane Doe",
    Status: models.UserActive,
}

if err := user.Validate(); err != nil {
    if verr, ok := err.(*models.ValidationError); ok {
        // structured error with field context
        log.Printf("field=%s msg=%s", verr.Field, verr.Message)
    }
    return err
}
```

### Creating a New Model (pattern)
```go
// internal/models/widget.go
package models

type Widget struct {
    ID    uuid.UUID
    Name  string
    State WidgetState
}

func (w *Widget) Validate() error {
    if w.Name == "" {
        return newValidationError("name", "is required")
    }
    if w.State < WidgetActive || w.State >= widgetStateUnsupported {
        return newValidationError("state", "invalid value")
    }
    return nil
}
```

### Example: UserStatus Enum
```go
// internal/models/user_status.go
status := models.UserActive
fmt.Println(status.String()) // "active"

parsed, err := models.UserStatusFromString("DELETED")
if err != nil {
    // invalid value
}
fmt.Println(parsed == models.UserDeleted) // true
```

### Creating a New Enum (pattern)
```go
// internal/models/widget_state.go
package models

type WidgetState int

const (
    WidgetActive WidgetState = iota
    WidgetDisabled
    widgetStateUnsupported
)

var widgetStates = [...]string{
    WidgetActive:   "active",
    WidgetDisabled: "disabled",
}

func (s WidgetState) String() string {
    if s < 0 || int(s) >= len(widgetStates) {
        return ""
    }
    return widgetStates[s]
}

func WidgetStateFromString(v string) (WidgetState, error) {
    for i, r := range widgetStates {
        if strings.EqualFold(v, r) {
            return WidgetState(i), nil
        }
    }
    return widgetStateUnsupported, fmt.Errorf("invalid widget state %q", v)
}
```

## Limitations
This is a basic, generic Go microservice template designed to provide a clear structure and foundational tooling. It remains intentionally minimal.

## Configuration
- Defaults: `env` defaults to `prod` (`config/init.go:setDefaults`).
- Precedence: flags > env vars > config file.
- Env var naming: dots become underscores (Viper replacer).
- To add a config field:
  ```go
  // config/scheme.go
  type Scheme struct {
      Env  string // existing
      Port int    // new
  }

  // config/init.go
  func setDefaults() {
      viper.SetDefault("env", "prod")
      viper.SetDefault("port", 8080)
  }

  // cmd/root (bind a flag)
  cmd.Flags().Int("port", 0, "port to listen on")
  ```
  Precedence will ensure flag > env > config file for `port` as well.

## CLI
- Root command name: `microservice-template`.
- Subcommands: `serve` (current runtime hook). Add more via `cmd/<name>` and register on root.
- Version output: `./microservice-template --version` (ldflags populate `pkg/version`).
- `serve` lifecycle: `PreRun` logs version; `RunE` should start your workloads; `PostRun` always stops app.
- Adding a new command (example):
  ```go
  // cmd/health/health.go
  package health

  import "github.com/spf13/cobra"

  func Cmd() *cobra.Command {
      return &cobra.Command{
          Use:   "health",
          Short: "Health probe",
          RunE: func(_ *cobra.Command, _ []string) error {
              // add checks here
              return nil
          },
      }
  }
  ```
  Register it in `cmd/microservice-template.go`: `rootCmd.AddCommand(health.Cmd())`.

## Development Workflow
- Format: `gofmt` (used via go tooling).
- Lint: `make lint` (golangci-lint; see `.golangci.yml`).
- Tests: `make test` or `go test ./...`; single test example `go test ./cmd/root -run TestInitializeConfig -count=1`.
- Build: `make build` (CGO disabled; ldflags inject version info).
- Deps: `make tidy` after changes; `make update` to bump modules.

## CI/CD
- Workflows: `.github/workflows/ci.yml` and `.github/workflows/release.yml`.
  - CI (`ci.yml`): on PRs and `main`, runs `make lint`, `make test`, `make build` (Go 1.24) with module caching.
  - Release (`release.yml`): on `main`, reruns lint/test/build, determines next incremental tag (`v1`, `v2`, …), pushes the tag, and creates a GitHub release with autogenerated notes (source-only). Uses `GITHUB_TOKEN`; no extra secrets needed.
- Branch protection (recommended): require CI checks (`lint`, `test`, `build`) to pass before merging to `main` and limit direct pushes.

## Versioning
- `Makefile` injects name/tag/commit/branch/remote/build date into `pkg/version` via ldflags.
- `pkg/version` formats a multi-line version string and handles unspecified values.
- Sample output:
  ```
  Template-service v0.0.0
  Branch main, commit hash: abcdef123
  Origin repository: https://github.com/org/repo
  Compiled at: 2026-01-16 20:58:09 +0000 UTC
  ©2026
  ```

## Logging
- `pkg/logger.Log()` returns a logrus logger with full timestamps.
- Example:
  ```go
  log := logger.Log()
  log.Infof("starting service", "env=%s", cfg.Env)
  log.Errorf("failed to start: %v", err)
  ```

## Extending the template
> **Note:** To rename an existing project, see the [Renaming the project](#renaming-the-project) section in Quickstart.

- Add config: update `Scheme`, `setDefaults`, and CLI flags; test binding like in `cmd/root/root_test.go`.
- Add commands: create `cmd/<name>` with `cobra.Command`, register on root in `cmd/microservice-template.go`.
  After renaming the entrypoint file (e.g., `cmd/yourservice.go`), register new commands there.
- Add runtime logic: implement `App.Init/Serve/Stop` with proper context/shutdown handling and graceful shutdown.
- Add tests: follow table-driven patterns; reset global state (Viper) in `t.Cleanup`.

## Keeping Up-to-Date with Template Changes

This project can receive updates from the upstream template: [go-microservice-template](https://github.com/andskur/go-microservice-template).

### Initial setup (downstream projects)
```bash
make template-setup
```
This will:
- Add the template remote (`template`)
- Fetch the latest template changes
- Create `.template-version` to track sync state

### Checking for updates
```bash
make template-status
make template-diff       # summary diff vs template/main
make template-diff v1.2.0 # diff against a specific tag
```

### Syncing updates
```bash
make template-fetch      # fetch latest template changes
make template-sync       # merge template/main into current branch
make template-sync v1.2.0 # merge a specific tag
```

After merging:
- Resolve any conflicts manually
- Run tests: `make test` (and `make build` if desired)
- Commit with a clear message (e.g., `chore: sync from template v1.2.0`)

### Files likely to need attention during sync
- `README.md`, `AGENTS.md` (project-specific docs)
- `internal/application.go` (module registration)
- `config/scheme.go` and `config/init.go` (config schema/defaults)
- `Makefile` (custom targets)

### Best practices
- Sync regularly to reduce conflicts
- Keep sync commits separate from feature work
- Review `make template-diff` before merging
- Use `.template-version` to record the last synced template ref (updated automatically on successful sync)

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

For development guidelines and best practices, see [AGENTS.md](./AGENTS.md).

## License
This project is licensed under the MIT License — see the [LICENSE](./LICENSE) file for details.

## Author
Copyright (c) 2022 Andrey Skurlatov
