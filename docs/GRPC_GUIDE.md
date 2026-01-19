# gRPC Guide

This guide explains how to use the gRPC module in the microservice template.

## Overview
- The gRPC module is optional and enabled via `grpc.enabled`.
- Module path: `internal/grpc/` (module pattern: Init → Start → Stop → HealthCheck).
- Health: standard `grpc.health.v1` service is registered.
- Middleware: logging + recovery (no Sentry).
- Protocol source is expected to come from the shared protocols repo (git subtree), not bundled locally.

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

You can use either protoc (existing) or Buf (recommended). The default protocols repo is `https://github.com/andskur/protocols-template.git` and is pulled via git subtree.

### Buf workflow (recommended)
```bash
make buf-install                     # Install Buf CLI
make buf-lint                        # Lint protos
make buf-generate PROTO_PACKAGE=user # Generate Go code for one package
make buf-generate-all                # Generate Go code for all packages
```

### Protoc workflow (traditional)
```bash
make proto-install                   # Install protoc plugins
make proto-setup                     # Add protocols via subtree (uses PROTO_REPO default)
make proto-generate PROTO_PACKAGE=user
make proto-generate-all
make proto-clean
```

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
