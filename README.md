# go-microservice-template

A minimal Go service template with Cobra/Viper CLI wiring, ldflags-based versioning, logrus logging, make targets, tests, and GitHub Actions CI (lint/test/build on PRs/main plus auto-tagged releases on `main`).

## Quickstart
- Requirements: Go 1.21+ (module sets 1.23/1.24), GNU `make`.
- Clone and create your branch: `git checkout -b feature/your-branch`.
- **Rename for your service**:
  - Replace `microservice-template` with your module path (e.g., `github.com/yourorg/yourservice`) in `go.mod` and Go imports.
  - Replace binary/entrypoint names: update `APP`, `APP_ENTRY_POINT`, and `GITVER_PKG` in `Makefile`; rename `cmd/microservice-template.go` accordingly.
  - Update CLI command name in `cmd/root/root.go` (`Use: "microservice-template"`).
  - Run `go mod tidy` after renaming.
  - Verify: `make build` and `./<your-binary> --version`.

  **Renaming examples (macOS/BSD sed):**
  ```bash
  # Set your values
  MODULE="github.com/yourorg/yourservice"
  BIN="yourservice"
  CLI="yourservice"

  # Update module
  go mod edit -module "$MODULE"

  # Update imports
  find . -type f -name '*.go' -exec sed -i '' "s|microservice-template|$MODULE|g" {} +

  # Update Makefile
  sed -i '' "s/^APP:=microservice-template/APP:=$BIN/" Makefile
  sed -i '' "s|^APP_ENTRY_POINT:=cmd/microservice-template.go|APP_ENTRY_POINT:=cmd/$BIN.go|" Makefile
  sed -i '' "s|^GITVER_PKG:=microservice-template/pkg/version|GITVER_PKG:=$MODULE/pkg/version|" Makefile

  # Rename entrypoint and CLI command
  mv cmd/microservice-template.go "cmd/$BIN.go"
  sed -i '' "s/Use: \"microservice-template\"/Use: \"$CLI\"/" cmd/root/root.go

  # Tidy deps
  go mod tidy
  ```
- Build: `make build` (binary `./microservice-template`).
- Run: `make run` (invokes `go run -race cmd/microservice-template.go serve`).
- Version: `./microservice-template --version`.
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
- CI: lint/test/build on PRs and `main`; auto-tagged release on `main` if checks pass.
- Tests included for CLI wiring, config defaults, versioning, logger singleton, helpers.
- Rename-friendly: single placeholder name `microservice-template` for binary/module/CLI.

**Trade-offs**
- No HTTP/GRPC server wired yet—skeleton only; you add runtime workloads.
- Viper globals: reset carefully in tests (see patterns in `cmd/root/root_test.go`).
- Release workflow auto-tags incrementally (`v1`, `v2`, …) and creates a GitHub release on `main` (source-only attachments).

## Project Structure
- `cmd/microservice-template.go`: entrypoint; builds root command, adds `serve`, executes CLI.
- `cmd/root`: root command, version template, config init (`initializeConfig`), flag/env binding.
- `cmd/serve`: `serve` command; `PreRun` logs version, `RunE` calls `App.Init`/`Serve`, `PostRun` stops app.
- `internal/application.go`: `App` struct; lifecycle `Init`/`Serve`/`Stop`; helper `CreateAddr`.
- `config/init.go`: Viper defaults (`env=prod` via `setDefaults`); `config/scheme.go` defines config shape.
- `pkg/logger`: logrus singleton `Log()`.
- `pkg/version`: ldflags-driven version metadata and formatted output.
- `Makefile`: build/test/lint/tidy/update targets; ldflags for versioning.
- `.github/workflows/ci.yml`: CI pipeline (lint/test/build on PRs and `main`).
- `.github/workflows/release.yml`: release pipeline on `main` (reruns checks, auto-tags `vN`, creates release).

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
- Add config: update `Scheme`, `setDefaults`, and CLI flags; test binding like in `cmd/root/root_test.go`.
- Add commands: create `cmd/<name>` with `cobra.Command`, register on root in `cmd/microservice-template.go`.
  After renaming the entrypoint file (e.g., `cmd/yourservice.go`), register new commands there.
- Add runtime logic: implement `App.Init/Serve/Stop` with proper context/shutdown handling.
- Add tests: follow table-driven patterns; reset global state (Viper) in `t.Cleanup`.
