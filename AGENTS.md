# Agent Guide for go-microservice-template

Scope: applies to entire repo.
No other AGENTS.md or Cursor/Copilot rules found.

## Commands
- Prefer Makefile targets when available.
- Build: `make build` (uses ldflags for version info).
- Build binary output: `./template-service` in repo root.
- Run app: `make run` (invokes `go run -race cmd/template-service.go serve`).
- Tidy modules: `make tidy`.
- Update deps: `make update`.
- Lint: `make lint` (golangci-lint).
- Install linter if missing: `make lint-install`.
- Test all: `make test` or `go test ./...`.
- Test with coverage: `make test-coverage` (produces `coverage.out`).
- Single package test: `go test ./internal` (replace path as needed).
- Single test by name: `go test ./... -run TestName -count=1`.
- Verbose test output: `go test -v ./...`.
- Clean built binary: `make clean`.
- Show version template at runtime: `./template-service --version`.
- Preferred go toolchain: Go 1.21+ (module uses 1.21 if set).
- Dockerfile builds minimal CGO-disabled binary; keep env CGO_ENABLED=0 unless needed.

## Project layout & code structure
- `cmd/template-service.go`: entry; wires cobra root + serve, executes CLI.
- `cmd/root`: root command, version template, config init via `initializeConfig` + flag/env binding.
- `cmd/serve`: serve command; `PreRun` logs version, `RunE` calls `App.Init`/`Serve`, `PostRun` stops app.
- `internal/application.go`: `App` struct with config/version; lifecycle `Init`/`Serve`/`Stop`; helper `CreateAddr`.
- `config/`: defaults in `init.go`; schema in `scheme.go`.
- `Makefile`: canonical targets for build/test/lint/tidy/run; injects versioner ldflags.
- `Dockerfile`: minimal CGO-disabled build pipeline; honors `APP` name.
- Tests: place alongside code in same package; keep hermetic.

## Config & environment
- Defaults set in `config/init.go`; extend with viper SetDefault.
- Schema in `config/scheme.go`; keep exported fields with comments.
- Config loading: viper reads config file if present; env overrides with `.`→`_` replacer.
- Honor envs `ENV` etc; allow empty env (viper AllowEmptyEnv).
- Keep flags/env names kebab/underscore aligned with viper mapping.
- Prefer `PersistentPreRunE` for config initialization, return wrapped errors.
- Avoid global mutable state; pass config through `*config.Scheme`.
- Document required env vars in README when adding new ones.

## CLI behavior
- Root use string `microservice`; update consistently.
- Add commands via cobra; keep Short concise.
- Use `RunE` for error returns, not `Run`.
- Set version template on root via `app.Version()`; keep output stable.
- PreRun hooks can log version; PostRun should stop app gracefully.
- Return errors with context; cobra will print.

## Build & release
- CGO disabled in Makefile build; enable only if dependency requires.
- Ldflags from Makefile inject name/commit/tag/branch/remote/build date (now using `pkg/version`).
- Do not hardcode version strings; rely on versioner package.
- Keep binary name driven by `APP` variable (`template-service`).
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
- Golangci-lint default config not checked in; rely on defaults.
- Typical enables include `errcheck`, `gosimple`, `staticcheck`; fix violations not silenced.
- If adding config, place at repo root as `.golangci.yml`.
- Keep lines < 120 chars when reasonable.
- No trailing whitespace; keep unix line endings.

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
- Use `github.com/misnaged/annales/logger`'s `Log()` accessor.
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

## Documentation
- Update README when adding features, flags, or envs.
- Keep AGENTS.md in sync with build/test changes.
- Comment exported types and functions; keep TODOs actionable.
- Keep examples minimal and runnable where possible.

## Style preferences
- Favor small, composable functions.
- Validate inputs early; return fast on error.
- Avoid mutating arguments when not obvious.
- Prefer `const` over `var` when values fixed.
- Keep public surface area lean; expose only needed types.
- Avoid over-engineering; simplicity first.

## When unsure
- Ask for clarification via issues/PR description.
- Default to Go community conventions when unspecified.
- Keep changes minimal and reversible.
- Run lint/tests before submitting changes.
- If adding tools or scripts, document invocation in this file.
