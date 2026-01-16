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
