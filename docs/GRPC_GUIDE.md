# gRPC Guide

This guide explains how to use the gRPC module in the microservice template.

## Overview
- The gRPC module is optional and enabled via `grpc.enabled`.
- Module path: `internal/grpc/` (module pattern: Init → Start → Stop → HealthCheck).
- Health: standard `grpc.health.v1` service is registered.
- Middleware: logging + recovery (no Sentry).
- Protocol sources: the bundled example API in `api/proto/` and, optionally, a shared protocols repository added under `protocols/` via git subtree (see below).

## Configuration

```yaml
grpc:
  enabled: true
  host: 0.0.0.0
  port: 9090
  timeout: 30s
  max_send_msg_size: 62914560   # 60MB
  max_recv_msg_size: 62914560   # 60MB
  num_stream_workers: 0         # 0 uses default
```

Env vars map with uppercase and underscores (e.g., `GRPC_ENABLED`, `GRPC_PORT`).

## Protobuf Workflow

Protobuf code is generated with [Buf](https://buf.build) using the locally installed, version-pinned
`protoc-gen-go` and `protoc-gen-go-grpc` plugins (no `protoc`, no Buf registry access). Generated
`*.pb.go` files are not committed; `make generate` (or `make generate-proto`) produces them.

There are two proto sources:

| Directory | What | Generated import path |
|-----------|------|-----------------------|
| `api/proto/` | Bundled example API (`userservice`) used by `internal/grpc`, `internal/grpcclient` and `pkg/userservice` | `<module>/api/proto/userservice` (explicit `go_package`) |
| `protocols/` | Shared protocols repository added as a git subtree (absent in a fresh template) | `<module>/protocols/<path>` (rewritten by `buf.gen.yaml`) |

The example lives in `api/proto/` (not `protocols/`) so that `make proto-setup` can add a shared
protocols repository at `protocols/` on a fresh template without conflicts, and the example keeps
building next to it until you remove it.

### Shared protocols (git subtree)

```bash
# Requires a clean working tree (git subtree requirement); PROTO_REPO/PROTO_BRANCH are optional
make proto-setup PROTO_REPO=https://github.com/your-org/your-protocols.git PROTO_BRANCH=main
make generate                        # generates api/proto and protocols/
make proto-update                    # later: pull upstream changes into protocols/
```

The `go_package` options written in the shared `.proto` files point at the protocols repository
itself. `buf.gen.yaml` enables Buf managed mode with a `go_package_prefix` override of
`<module>/protocols`, so a file `protocols/user/v1/user.proto` is generated to
`protocols/user/v1/*.pb.go` and imported as `<module>/protocols/user/v1`. `make rename` updates the
prefix together with the module path.

### Targets
```bash
make generate-proto                   # Generate api/proto and (if present) protocols/
make proto-generate PROTO_PACKAGE=user # Generate one package of protocols/
make buf-lint                         # Lint api/proto and protocols/
make buf-breaking                     # Breaking-change check for protocols/ against main
make proto-clean                      # Remove generated *.pb.go
```
`make proto-generate-all` and `make buf-generate-all` are kept as aliases of `make generate-proto`.

## Handler Implementation Pattern

Handlers for specific services should live under `internal/grpc/` and register via `registerHandlers()`. Pull generated protos from the shared protocols repo first, then add conversions and handlers for your services.

### Example pattern
- Validate inputs, call service methods, and convert models to proto.
- Use `google.golang.org/grpc/status` and `codes` for errors.

## Proto Conversion Helpers

Add conversion helpers for your services under `internal/grpc/` (e.g., `conversions.go` per service). These should map between internal models and generated proto types.

## Health Checks

- Standard gRPC health service (`grpc.health.v1.Health`) registered in `Server.RegisterHealthService()`.
- Use `grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check` to probe.

## Running and Testing

1) Enable gRPC:
```bash
export GRPC_ENABLED=true
export GRPC_PORT=9090
```

2) Run the app:
```bash
make run
```

3) Unit + integration tests (gRPC package only):
```bash
make test-grpc
```

4) Test with grpcurl:
```bash
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check
```

## Adding New gRPC Services

1. Pull/update protocols from the shared repo (`make proto-setup` / `make proto-update`).
2. Generate code (Buf preferred): `make buf-generate PROTO_PACKAGE=<service>`.
3. Add conversion helpers in `internal/grpc/` for your types.
4. Implement handlers using `service.IService` (or other deps) and register them in `registerHandlers()`.

## Production Notes
- Keep health checks fast (<2s).
- Adjust keep-alives and message sizes via config if needed.
- Enable TLS/mTLS by extending `GRPCConfig` (not included by default).
- For metrics/tracing, add interceptors later (kept minimal here).
