# go-microservice-template

A minimal Go service template with Cobra/Viper CLI wiring, ldflags-based versioning, logrus logging, make targets, tests, and GitHub Actions CI (lint/test and tagged releases on `main`).

## Quickstart
- Requirements: Go 1.21+ (module sets 1.23/1.24), GNU `make`.
- Clone and create your branch: `git checkout -b feature/your-branch`.
- Build: `make build` (binary `./template-service`).
- Run: `make run` (invokes `go run -race cmd/template-service.go serve`).
- Lint: `make lint` (golangci-lint).
- Test: `make test` or single test `go test ./... -run TestName -count=1`.
- Coverage: `make test-coverage` (writes `coverage.out`).
- Tidy deps: `make tidy`; update deps: `make update`.

## Features / Pros & Cons
**Pros**
- Simple, small footprint; uses standard libs plus Cobra/Viper/logrus.
- Version info injected via ldflags (`pkg/version`).
- Structured logging via singleton `pkg/logger`.
- Makefile with race-enabled run, build, lint, test, tidy/update.
- CI: lint/test on PRs/push; tagged release on `main` if CI passes.
- Tests included for CLI wiring, config defaults, versioning, logger singleton, helpers.

**Trade-offs**
- No HTTP/GRPC server wired yet—skeleton only; you add runtime workloads.
- Viper globals: reset carefully in tests (see patterns in `cmd/root/root_test.go`).
- Release workflow requires `RELEASES_ACTION_GITHUB_TOKEN`; coverage badge branch optional.

## Project Structure
- `cmd/template-service.go`: entrypoint; builds root command, adds `serve`, executes CLI.
- `cmd/root`: root command, version template, config init (`initializeConfig`), flag/env binding.
- `cmd/serve`: `serve` command; `PreRun` logs version, `RunE` calls `App.Init`/`Serve`, `PostRun` stops app.
- `internal/application.go`: `App` struct; lifecycle `Init`/`Serve`/`Stop`; helper `CreateAddr`.
- `config/init.go`: Viper defaults (`env=prod` via `setDefaults`); `config/scheme.go` defines config shape.
- `pkg/logger`: logrus singleton `Log()`.
- `pkg/version`: ldflags-driven version metadata and formatted output.
- `Makefile`: build/test/lint/tidy/update targets; ldflags for versioning.
- `.github/workflows/ci.yml`: CI pipeline (lint/test on PRs/push; release on `main`).

## Configuration
- Defaults: `env` defaults to `prod` (`config/init.go:setDefaults`).
- Precedence: flags > env vars > config file.
- Env var naming: dots become underscores (Viper replacer).
- To add a config field: add to `config/Scheme` with comment; set default in `setDefaults`; bind flag in CLI as needed; document in README/AGENTS.

## CLI
- Root command name: `microservice`.
- Subcommands: `serve` (current runtime hook). Add more via `cmd/<name>` and register on root.
- Version output: `./template-service --version` (ldflags populate `pkg/version`).
- `serve` lifecycle: `PreRun` logs version; `RunE` should start your workloads; `PostRun` always stops app.

## Development Workflow
- Format: `gofmt` (used via go tooling).
- Lint: `make lint` (golangci-lint; see `.golangci.yml`).
- Tests: `make test` or `go test ./...`; single test example `go test ./cmd/root -run TestInitializeConfig -count=1`.
- Build: `make build` (CGO disabled; ldflags inject version info).
- Deps: `make tidy` after changes; `make update` to bump modules.

## CI/CD
- Workflow: `.github/workflows/ci.yml`.
  - On PRs/push: checkout, setup Go from `go.mod`, run golangci-lint, run `go test ./... -coverprofile=cover.out`, compute coverage.
  - On `main`: additionally bump tag and create GitHub release (requires `RELEASES_ACTION_GITHUB_TOKEN`). Optional coverage badge on `main` branch using `action-badges/core`.

## Versioning
- `Makefile` injects name/tag/commit/branch/remote/build date into `pkg/version` via ldflags.
- `pkg/version` formats a multi-line version string and handles unspecified values.

## Logging
- `pkg/logger.Log()` returns a logrus logger with full timestamps. Use at app boundaries; avoid logging secrets.

## Extending the template
- Add config: update `Scheme`, `setDefaults`, and CLI flags; test binding like in `cmd/root/root_test.go`.
- Add commands: create `cmd/<name>` with `cobra.Command`, register on root in `cmd/template-service.go`.
- Add runtime logic: implement `App.Init/Serve/Stop` with proper context/shutdown handling.
- Add tests: follow table-driven patterns; reset global state (Viper) in `t.Cleanup`.
