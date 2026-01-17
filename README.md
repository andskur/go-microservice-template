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
- Coverage: `make test-coverage` (writes `coverage.out`).
- Tidy deps: `make tidy`; update deps: `make update`.

### Renaming the project
- Command: `make rename NEW_NAME=my-service` (required parameter).
- Valid NEW_NAME: lowercase letters, numbers, hyphens, optional `/` segments (e.g., `my-service`, `github.com/yourorg/my-service`).
- Updates: module path and imports, Makefile vars, entrypoint file, Cobra root `Use`, Dockerfile binary, README/AGENTS references, optional git remote.
- Verify after rename: `go test ./...`, `make build`, `./<new-binary> --version`.

## Features
- Simple, small footprint using standard libs plus Cobra/Viper/logrus.
- Version metadata injected via ldflags (`pkg/version`).
- Structured logging via `pkg/logger` singleton.
- Makefile targets for build/run/lint/test/tidy/update.
- CI pipeline: lint/test/build on PRs and `main`; release pipeline auto-tags on `main` and publishes a GitHub release (source-only).
- Tests included for CLI wiring, config defaults, versioning, logger singleton, helpers.
- Rename-friendly: single placeholder name with automated `make rename` target.

## Limitations
This is a basic, generic Go microservice template designed to provide a clear structure and foundational tooling. It remains intentionally minimal:
- No HTTP or gRPC server wired—skeleton only; you add runtime workloads.
- Uses Viper globals; reset carefully in tests (see `cmd/root/root_test.go`).
- Release workflow auto-increments tags (`v1`, `v2`, …) on `main`.

## Project Structure
```
go-microservice-template/
│
├── .github/workflows/
│   ├── ci.yml              # CI: lint, test, build on PRs and main
│   └── release.yml         # Release: auto-tag and GitHub release on main
│
├── cmd/                    # Command-line interface
│   ├── microservice-template.go  # Main entry; builds root command and executes CLI
│   ├── root/               # Root command, version template, config initialization
│   │   ├── root.go
│   │   └── root_test.go
│   └── serve/              # Serve command; lifecycle hooks (PreRun/RunE/PostRun)
│       ├── serve.go
│       └── serve_test.go
│
├── config/                 # Configuration management
│   ├── init.go             # Viper defaults (env=prod)
│   ├── scheme.go           # Configuration structure definition
│   └── init_test.go
│
├── internal/               # Private application code
│   ├── application.go      # App struct with Init/Serve/Stop lifecycle
│   └── application_test.go
│
├── pkg/                    # Public reusable packages
│   ├── logger/             # Logrus singleton for structured logging
│   │   ├── logger.go
│   │   └── logger_test.go
│   └── version/            # Version metadata injected via ldflags
│       ├── version.go
│       └── version_test.go
│
├── scripts/                # Automation scripts
│   └── rename.sh           # Automated project rename script
│
├── .dockerignore           # Docker build context exclusions
├── .golangci.yml           # Linter configuration (extensive rule set)
├── Dockerfile              # Multi-stage build (golang:1.24 -> scratch)
├── Makefile                # Build targets: build, run, test, lint, tidy, update
├── LICENSE                 # MIT License
├── README.md               # Project documentation
├── AGENTS.md               # Development guidelines
├── go.mod                  # Go module definition
└── go.sum                  # Dependency checksums
```

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

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

For development guidelines and best practices, see [AGENTS.md](./AGENTS.md).

## License
This project is licensed under the MIT License — see the [LICENSE](./LICENSE) file for details.

## Author
Copyright (c) 2022 Andrey Skurlatov
