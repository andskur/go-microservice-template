# Agent Guide for go-microservice-template

Scope: applies to entire repo.
No other AGENTS.md or Cursor/Copilot rules found.

## Commands
- Prefer Makefile targets when available.
- Build: `make build` (uses ldflags for version info).
- Build binary output: `./microservice-template` in repo root.
- Run app: `make run` (invokes `go run -race cmd/microservice-template.go serve`).
- Show version template at runtime: `./microservice-template --version`.
- `cmd/microservice-template.go`: entry; wires cobra root + serve, executes CLI.
- Keep binary name driven by `APP` variable (`microservice-template`).
- Rename project: `make rename NEW_NAME=my-service` (validates name, prompts, updates module/imports/Makefile/Dockerfile/docs/CLI).
- Rename validation: lowercase letters, numbers, hyphens, optional path segments (`my-service`, `github.com/org/my-service`).
- After rename: verify with `go test ./...`, `make build`, `./<new-binary> --version`.
- For cross-compilation, override `GOOS`/`GOARCH` on make invocations.
- Optimize size with `-w -s`; avoid removing if debug symbols needed locally.
- Local stack helpers: `make compose-up`, `make compose-down`, `make compose-restart` (Postgres + migrations + Redis).

## Testing guidance
- Place tests alongside code (`*_test.go`).
- Use table-driven tests for branches.
- Avoid global state; reset viper changes in tests (use `t.Cleanup`).
- Prefer `require`/`assert`? Not present; stick to stdlib `testing`.
- For CLI commands, test `cobra.Command` behaviors with `ExecuteC`.
- For signal handling, use buffered channels and context where possible.
- Keep tests hermetic; avoid network/filesystem unless tempdir.
- Running single test example: `go test ./cmd/root -run TestInitializeConfig -count=1`.

## Linting & formatting
- Use `gofmt` on all Go files; do not hand-format.
- Golangci-lint config is checked in at `.golangci.yml` with extensive enables (errcheck, govet, staticcheck, revive, misspell, bodyclose, gosec, gocyclo, dupl, etc.).
- Custom settings: gocyclo min-complexity 15; dupl threshold 100; nakedret max 30 lines; line-length 120; errcheck checks type assertions.
- Test files exclude some linters (gocyclo, errcheck, dupl, gosec, goconst) per config.
- Keep lines < 120 chars when reasonable; no trailing whitespace; unix line endings.

## Imports
- Group imports: stdlib, blank line, third-party, blank line, local (`microservice-template/...`).
- Keep deterministic ordering (gofmt).
- Avoid import aliases unless necessary to disambiguate.
- Do not use dot imports.
- Prefer explicit package names over anonymous `_` except for tools.
- Remove unused imports promptly.

## Naming
- Exported identifiers need doc comments in Go style (`Name ...`).
- Keep acronyms in caps (`URL`, `ID`, `HTTP`).
- Functions should be verbs; types nouns.
- Avoid stutter with package names (e.g., `config.Scheme` not `config.ConfigScheme`).
- Avoid one-letter variables except tiny scopes (`i`, `j` in loops).
- Constants in PascalCase for exports, camelCase for internal.
- CLI command variables should be `cmd` or descriptive.

## Error handling
- Prefer `fmt.Errorf` with `%w` when wrapping.
- Return errors; avoid panics except truly unrecoverable.
- Log at the boundary; functions should return errors, not log + return.
- Use `pkg/logger.Log()` (logrus) with appropriate level; avoid noisy Info for errors.
- Include context in messages (`action: %w`).
- Handle config file not found gracefully (already in root initializeConfig).
- When stopping app, log underlying stop errors; do not swallow silently.

## Concurrency & lifecycle
- Use contexts for cancellation; propagate from CLI when adding long-running tasks.
- For goroutines, ensure shutdown via channels/context and `sync.WaitGroup`.
- Signal handling currently in `Serve`; avoid blocking forever without work.
- Use buffered channel size 1 for os.Signal as present.
- Close resources in `Stop`; make it idempotent.
- Avoid global mutable singletons; inject dependencies into `App`.

## Logging
- Use `pkg/logger.Log()` (logrus) accessor.
- Prefer structured messages with format verbs.
- Do not log secrets or PII.
- Keep startup logs concise; include version once.

## Configuration defaults
- Add new defaults in `config/init.go` with comments.
- Keep `Scheme` fields tagged and documented; align names with viper keys.
- Use concrete types; avoid `interface{}`.
- Validate config during `App.Init`; return errors not fatal exit.
- Document required settings in README and AGENTS when changing.

## Dependencies
- Keep `go.mod` tidy via `make tidy` after dependency changes.
- Avoid replacing module paths unless necessary; remove temporary replaces before commit.
- Vendor not used; rely on modules.
- For CLI dependencies, prefer minimal footprint; avoid heavy frameworks.
- Check licenses before adding new deps.

## Module System
- Modules are optional components implementing `module.Module` interface (Init/Start/Stop/HealthCheck).
- Located in `internal/module/` package; Manager handles lifecycle orchestration.
- Lifecycle: Init → Start → Stop (Stop happens in reverse order for LIFO cleanup).
- Registration order determines initialization order; manual wiring in `internal/application.go`.
- Each module may have config in `config/scheme.go` with defaults in `config/init.go`; some modules (e.g., service) are always on and rely on optional dependencies instead of a config flag.
- Module dependencies: use constructor injection (type-safe, explicit); allow nil for optional deps and handle gracefully.
- Keep modules focused and single-purpose; make Init() idempotent.
- Use goroutines in Start() for background work; respect context timeout in Stop().
- HealthCheck() must be fast (< 2s); log all lifecycle events.
- Models live in `internal/models` and stay pure (no DB hooks/tags) when DB is not in use; when database is enabled with go-pg, models include go-pg struct tags and hooks for UUID/status/timestamps.
- See `docs/MODULE_DEVELOPMENT.md` for detailed guide.

### gRPC Module
- Optional; enabled via `grpc.enabled=true` in config.
- Config struct: `config.GRPCConfig` (`grpc.*` keys), defaults in `config/init.go`.
- Module implementation: `internal/grpc/` (Init/Start/Stop/HealthCheck).
- Handler registration is uncommented in `internal/grpc/module.go` (`registerHandlers`).
- Health: standard `grpc.health.v1` service registered automatically.
- Middleware: logging + recovery (no Sentry). Logging at Info for requests, Error for failures.
- Proto conversions live in `internal/grpc/` (keep models free of proto deps).
- Protocols are sourced via subtree from `https://github.com/andskur/protocols-template.git`; no bundled example is kept locally.

### Protobuf Workflow
- Targets in Makefile:
  - `make proto-install`: install protoc plugins (go, go-grpc).
  - `make proto-setup PROTO_REPO=<url>`: add protocols as subtree (default: andskur/protocols-template).
  - `make proto-update`: update subtree.
  - `make proto-generate PROTO_PACKAGE=<name>`: generate Go code from `protocols/<name>/*.proto` (protoc).
  - `make proto-generate-all`: generate Go code from all packages (protoc).
  - `make buf-install`: install Buf CLI.
  - `make buf-lint`: lint protos with Buf.
  - `make buf-breaking`: check breaking changes vs main.
  - `make buf-generate PROTO_PACKAGE=<name>`: generate Go code from `protocols/<name>` (Buf).
  - `make buf-generate-all`: generate Go code for all packages (Buf).
  - `make proto-clean`: remove generated `.pb.go` files.
  - `make test-grpc`: run gRPC package tests (unit + integration).
- Generated files are ignored (.gitignore). Pull the shared protocols repo via subtree before generating.

### gRPC Handler Patterns
- Implement handlers under `internal/grpc/` for your services.
- Depend on `service.IService`; validate inputs; return gRPC status errors.
- Register services in `registerHandlers()`.
- Add conversion helpers under `internal/grpc/` for model↔proto mappings.

### HTTP Module
- Optional; enabled via `http.enabled=true` in config.
- Config struct: `config.HTTPConfig` (`http.*` keys), defaults in `config/init.go`.
- Module implementation: `internal/http/` (Init/Start/Stop/HealthCheck).
- Uses go-swagger for spec-first API development (Swagger 2.0 format).
- API spec: `api/swagger.yaml`; generated server code in `internal/http/server/` (gitignored).
- Middleware chain (justinas/alice): Recovery → Logger → CORS → RateLimit → Handler.
- Auth: JWT validation in `internal/http/auth/auth.go`; mock mode for local dev with `http.mock_auth=true`.
- Handler pattern: struct with dependencies → `Handle(w http.ResponseWriter, r *http.Request)` method.
- Error mapping: service errors map to HTTP status (ErrNotFound → 404, ErrInvalidInput → 400, ErrRepositoryUnavailable → 503).
- Model conversions: `internal/http/formatter/` package converts domain models ↔ API models (generated from swagger).
- Health: `GET /health` returns 200 with status; checks all module health.

### Swagger Workflow
- Targets in Makefile:
  - `make swagger-install`: install go-swagger CLI.
  - `make swagger-validate`: validate `api/swagger.yaml` spec.
  - `make generate-api`: generate server code from spec (creates `internal/http/server/` and `internal/http/models/`).
  - `make swagger-clean`: remove generated code.
  - `make test-http`: run HTTP package tests (unit + integration).
- Generated files are ignored (.gitignore). Edit `api/swagger.yaml` then regenerate with `make generate-api`.
- Use Swagger 2.0 format (not OpenAPI 3.0); go-swagger doesn't support OpenAPI 3.0 yet.

### HTTP Handler Patterns
- Implement handlers in `internal/http/handlers/` as structs with dependencies.
- Handler struct pattern:
  ```go
  type GetUserHandler struct {
      service service.IService
      log     *logrus.Logger
  }
  
  func NewGetUserHandler(svc service.IService, log *logrus.Logger) *GetUserHandler {
      return &GetUserHandler{service: svc, log: log}
  }
  
  func (h *GetUserHandler) Handle(w http.ResponseWriter, r *http.Request) {
      // Extract params, validate, call service, convert response
  }
  ```
- Register handlers in `internal/http/module.go` → `setupRoutes()`.
- Use `DefaultError(w, code, msg)` helper for error responses.
- Convert models with formatters: `formatter.UserToAPI(domainUser)` → API model.
- Validate inputs; return 400 for invalid input, 404 for not found, 503 for service unavailable.
- Log errors with context; use structured logging (logrus).

### HTTP Error Handling
- Service layer returns sentinel errors (`service.ErrNotFound`, `service.ErrInvalidInput`, `service.ErrRepositoryUnavailable`).
- HTTP handlers map errors to status codes:
  - `service.ErrNotFound` → 404 Not Found
  - `service.ErrInvalidInput` or `models.ValidationError` → 400 Bad Request
  - `service.ErrRepositoryUnavailable` → 503 Service Unavailable
  - Other errors → 500 Internal Server Error
- Use `errors.Is()` to check error types; log internal errors, return generic messages to client.
- Example handler error mapping:
  ```go
  user, err := h.service.GetUserByEmail(ctx, email)
  if err != nil {
      if errors.Is(err, service.ErrNotFound) {
          DefaultError(w, http.StatusNotFound, "User not found")
          return
      }
      if errors.Is(err, service.ErrRepositoryUnavailable) {
          DefaultError(w, http.StatusServiceUnavailable, "Service temporarily unavailable")
          return
      }
      h.log.Errorf("get user by email: %v", err)
      DefaultError(w, http.StatusInternalServerError, "Internal server error")
      return
  }
  ```

### HTTP Authentication
- JWT validation in `internal/http/auth/auth.go` via `Authenticate()` middleware.
- Production: validates JWT signature, checks expiration, extracts claims (user ID, email, admin flag).
- Development: use `http.mock_auth=true` to bypass validation (accepts any Bearer token).
- Gatekeeper integration: detailed TODO in `auth.go` for future external auth service integration.
- Protected routes: apply `Authenticate()` middleware in handler chain.
- Claims stored in request context; retrieve with `auth.GetUserID(r)`, `auth.GetEmail(r)`, `auth.IsAdmin(r)`.

### HTTP Middleware
- Recovery: catch panics, log stack trace, return 500.
- Logger: log request method, path, duration, status code; Info for 2xx/3xx, Error for 4xx/5xx.
- CORS: configurable origins/methods/headers via `http.cors.*` config; preflight support.
- RateLimit: token bucket per IP with `http.rate_limit.requests_per_second` and burst size.
- Chain with justinas/alice: `alice.New(Recovery(), Logger(), CORS(), RateLimit()).Then(handler)`.
- Add new middleware in `internal/http/middlewares/`; register in module's `setupRoutes()`.

### HTTP Testing
- Test handlers with mock service implementing `service.IService`.
- Use `httptest.NewRecorder()` and `httptest.NewRequest()`.
- Table-driven tests for branches (success, not found, invalid input, service error).
- Test middleware with chained handlers; verify headers, status codes, response bodies.
- Test module lifecycle (Init/Start/Stop/HealthCheck) with mock config.
- Run with `make test-http` or `go test ./internal/http/...`.
- Example handler test:
  ```go
  func TestGetUserHandler_Success(t *testing.T) {
      mockSvc := &MockService{
          GetUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
              return &models.User{ID: uuid.Must(uuid.NewV4()), Email: email}, nil
          },
      }
      handler := NewGetUserHandler(mockSvc, logger.Log())
      req := httptest.NewRequest(http.MethodGet, "/users?email=test@example.com", nil)
      w := httptest.NewRecorder()
      handler.Handle(w, req)
      if w.Code != http.StatusOK {
          t.Errorf("expected 200, got %d", w.Code)
      }
  }
  ```

### Repository Layer with go-pg
- Repository module wraps `*pg.DB` connection and implements `IRepository` interface.
- Located in `internal/repository/`; registered only when `database.enabled=true`.
- Models use go-pg struct tags (`pg:"column"`) and hooks (`BeforeInsert`, `BeforeUpdate`, `AfterSelect`) for UUID generation, status conversion, timestamps.
- Status enums: use dual fields (`Status UserStatus pg:"-"` + `StatusSQL string pg:"status,use_zero"`); hooks convert between enum and string.
- UUID generation: handled in `BeforeInsert` if UUID is nil; ensures every insert has a UUID.
- Timestamps: DB defaults for `created_at`; DB trigger updates `updated_at` on row updates; Go hooks also set `updated_at` for defense in depth.
- Migrations: use `golang-migrate/migrate` via Makefile; migrations in `db/migrations/` dir; create with `make migrate-create NAME=<name>`; run with `make migrate-up`.
- Connection config: `pg.Options` uses `database.host`, `database.port`, `database.user`, `database.password`, `database.name`, pooling via `max_open_conns` and `max_idle_conns`.
- Health check: `SELECT 1` via `db.WithContext(ctx).Exec` in module `HealthCheck`.
- Graceful shutdown: `db.Close()` in module `Stop`; guard nil before closing.
- Query patterns: `db.Model(model).Column("table.*").Where(...).Select()` for reads; `.Returning("*").Insert()` for creates.
- Error handling: wrap with context (`fmt.Errorf("action: %w", err)`); check `pg.ErrNoRows` for not-found.
- UserGetter pattern: enum with `Get(query *orm.Query, model *Model)`; apply `WherePK()` or `Where()`.

### Database Migrations
- Migrations managed by `golang-migrate/migrate` CLI tool; install via `make migrate-install`.
- Migration files live in `db/migrations/` with sequential numbering (`000001`, `000002`, ...).
- Each migration has `.up.sql` (forward) and `.down.sql` (rollback) files.
- Schema uses PostgreSQL ENUM `user_status` ('active', 'deleted'); DB trigger updates `updated_at` column.
- Create migrations with `make migrate-create NAME=<descriptive_name>`; always test down migrations.
- Apply migrations with `make migrate-up`; rollback with `make migrate-down`; check version with `make migrate-version`; recover with `make migrate-force VERSION=<n>`.
- Use `make migrate-drop` (with confirmation) to drop all tables during local development only.
- Never modify applied migrations; create new migrations to fix issues.
- Use `IF EXISTS`/`IF NOT EXISTS` for idempotent migrations safe to re-run.
- Production: run migrations before app startup or as separate deployment job.

### Adding a New Module
1. Define config struct in `config/scheme.go` when the module is configurable; skip if always-on.
2. Add defaults in `config/init.go` for all config fields you add.
3. Implement `module.Module` interface in `internal/<name>/module.go`.
4. Register in `internal/application.go` → `registerModules()` in dependency order (infrastructure → business logic → transports).
5. Wire dependencies via constructor injection; pass nil for optional deps and document behavior.
6. Add tests for module lifecycle (Init/Start/Stop/HealthCheck) and dependency handling.
7. Document in README.md and MODULE_DEVELOPMENT.md.

### Module Best Practices
- Registration order: Infrastructure (DB, cache, queue) → Business logic (repos, services) → Transport (HTTP, gRPC).
- Constructor injection for dependencies: `NewModuleB(cfg, moduleA)`; avoid service locators/globals.
- Non-blocking Start: use `go m.runWorker(ctx)` for long-running operations.
- Graceful Stop: select on done channel and ctx.Done() with timeout.
- Error wrapping: use `fmt.Errorf("action: %w", err)` for context.

## Documentation
- Update README when adding features, flags, or envs.
- Keep AGENTS.md in sync with build/test/tool changes.
- If adding scripts to `scripts/`, document purpose and invocation here and in README.
- Comment exported types and functions; keep TODOs actionable.
- Keep examples minimal and runnable where possible.

## Style preferences
- Favor small, composable functions.
- Validate inputs early; return fast on error.
- Avoid mutating arguments when not obvious.
- Prefer `const` over `var` when values fixed.
- Keep public surface area lean; expose only needed types.
- Avoid over-engineering; simplicity first.

## Development Workflow
- Format: `gofmt` (used via go tooling).
- Lint: `make lint` (golangci-lint; see `.golangci.yml`).
- Tests: `make test` or `go test ./...`; single test example `go test ./cmd/root -run TestInitializeConfig -count=1`.
- Build: `make build` (CGO disabled; ldflags inject version info).
- Deps: `make tidy` after changes; `make update` to bump modules.

## Docker
- Dockerfile: multi-stage build (golang:1.24 builder → scratch); invokes `make build`.
- Binary name in Dockerfile COPY/ENTRYPOINT must match Makefile `APP` variable.
- Current binary: `/microservice-template` (synced with `APP:=microservice-template`).
- `make rename` updates Dockerfile automatically.
- Build: `docker build -t microservice-template .`
- Run: `docker run --rm microservice-template` (defaults to `serve`).
- Best practices: use `.dockerignore` to reduce context; multi-stage keeps final image minimal.
- Cross-compilation: override `GOOS`/`GOARCH` on `make build` before docker build if needed.

## Scripts
- `scripts/rename.sh`: automated project rename; invoked via `make rename NEW_NAME=...`.
- `scripts/template-sync.sh`: template sync helper (setup/status/diff/sync) used by Makefile targets.
- Validates Go module naming; prompts for confirmation; updates go.mod, imports, Makefile vars, entrypoint, CLI `Use`, Dockerfile, docs; optional git remote update; runs `go mod tidy`.
- New scripts: add to `scripts/`, make executable, document purpose and invocation here and in README.

## Template Synchronization (for downstream repos)
- One-time setup: `make template-setup` (adds `template` remote, fetches, creates `.template-version`).
- Check updates: `make template-status`; summary diff: `make template-diff` (optionally with tag).
- Sync updates: `make template-fetch` then `make template-sync` (optionally with tag ref).
- After sync: resolve conflicts if any, run `make test`/`make build`, commit separately (e.g., `chore: sync from template vX.Y.Z`).
- Likely conflict files: README.md, AGENTS.md, config/scheme.go, config/init.go, internal/application.go, Makefile.

## CI/CD
- Workflows: `.github/workflows/ci.yml` and `.github/workflows/release.yml`.
- CI: lint/test/build on PRs and `main` via `make lint`, `make test`, `make build`.
- Release: reruns lint/test/build, auto-tags incrementally (`v1`, `v2`, …), creates GitHub release on `main`.
- Branch protection recommended: require CI checks (`lint`, `test`, `build`) to pass before merging to `main` and limit direct pushes.

## Versioning
- `Makefile` injects name/tag/commit/branch/remote/build date into `pkg/version` via ldflags.
- `pkg/version` formats multi-line version output; handles unspecified values.
- Protocols: source from `https://github.com/andskur/protocols-template.git` via subtree; generate locally with Buf or protoc.

## Extending the template
- Add config: update `config/scheme.go`, defaults in `config/init.go`, bind flags in `cmd/root`; test bindings.
- Add commands: create `cmd/<name>` with `cobra.Command`, register on root in entrypoint.
- Add runtime logic: implement `App.Init/Serve/Stop` with proper shutdown; use contexts.
- Add tests: follow table-driven patterns; reset global state (Viper) in `t.Cleanup`.

## When unsure
- Ask for clarification via issues/PR description.
- Default to Go community conventions when unspecified.
- Keep changes minimal and reversible.
- Run lint/tests before submitting changes.
- If adding tools or scripts to `scripts/`, document invocation and purpose in this file.

