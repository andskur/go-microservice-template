# go-microservice-template

Template for building Go microservices.

## Build / Run / Test
- Build: `make build` (binary: `./template-service`).
- Run: `make run` (runs `go run -race cmd/template-service.go serve`).
- Lint: `make lint` (golangci-lint).
- Test all: `make test` or `go test ./...`.
- Single test: `go test ./... -run TestName -count=1` (e.g., `go test ./cmd/root -run TestInitializeConfig -count=1`).
- Coverage: `make test-coverage` (outputs `coverage.out`).

## Configuration
- Defaults set in `config/init.go` with Viper.
- Schema in `config/scheme.go`; keep exported fields documented.
- Env overrides use `.`→`_` replacer; empty envs allowed.

## Code structure
- Entry: `cmd/template-service.go` wires root + serve and executes CLI.
- Root command: `cmd/root` sets version template and initializes config.
- Serve command: `cmd/serve` logs version, calls `App.Init`/`Serve`, stops app in `PostRun`.
- Application: `internal/application.go` holds config/version, lifecycle, and helpers.
- Config: `config/` for defaults and schema.
- Logger: `pkg/logger` (logrus singleton).
- Version info: `pkg/version` (ldflags set via Makefile).

## Versioning
- Ldflags set in `Makefile` targeting `microservice-template/pkg/version` to inject service name, tag, commit, branch, origin URL, and build date.

## Notes
- Prefer `gofmt` and Go 1.21+.
- Keep tests hermetic and colocated with code.
