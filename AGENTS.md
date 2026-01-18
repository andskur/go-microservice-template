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
- Models live in `internal/models` and stay pure (no DB hooks/tags); use typed enums with String/FromString helpers and structured validation errors (`ValidationError` with Field/Message).
- See `docs/MODULE_DEVELOPMENT.md` for detailed guide.

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
- Validates Go module naming; prompts for confirmation; updates go.mod, imports, Makefile vars, entrypoint, CLI `Use`, Dockerfile, docs; optional git remote update; runs `go mod tidy`.
- New scripts: add to `scripts/`, make executable, document purpose and invocation here and in README.

## CI/CD
- Workflows: `.github/workflows/ci.yml` and `.github/workflows/release.yml`.
  - CI: lint/test/build on PRs and `main` via `make lint`, `make test`, `make build`.
  - Release: reruns lint/test/build, auto-tags incrementally (`v1`, `v2`, …), creates GitHub release on `main`.
- Branch protection recommended: require CI checks before merge; limit direct pushes to `main`.

## Versioning
- `Makefile` injects name/tag/commit/branch/remote/build date into `pkg/version` via ldflags.
- `pkg/version` formats multi-line version output; handles unspecified values.

## Extending the template
- Add config: update `config/scheme.go`, defaults in `config/init.go`, bind flags in `cmd/root`; test bindings.
- Add commands: create `cmd/<name>` with cobra.Command, register on root in entrypoint.
- Add runtime logic: implement `App.Init/Serve/Stop` with proper shutdown; use contexts.
- Add tests: table-driven, reset globals in `t.Cleanup`.

## When unsure
- Ask for clarification via issues/PR description.
- Default to Go community conventions when unspecified.
- Keep changes minimal and reversible.
- Run lint/tests before submitting changes.
- If adding tools or scripts to `scripts/`, document invocation and purpose in this file.
