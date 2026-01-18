# gRPC Guide

This guide explains how to use the gRPC module in the microservice template.

## Overview
- The gRPC module is optional and enabled via `grpc.enabled`.
- Module path: `internal/grpc/` (module pattern: Init → Start → Stop → HealthCheck).
- Health: standard `grpc.health.v1` service is registered.
- Middleware: logging + recovery (no Sentry).
- Example service: `protocols/user/` with `GetUser` and `CreateUser` methods.

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

1) Install tooling (one-time):
```bash
make proto-install
```

2) (Optional) Replace example protocols with your own via subtree:
```bash
make proto-setup PROTO_REPO=git@github.com:yourorg/your-protocols.git
# or HTTPS
make proto-setup PROTO_REPO=https://github.com/yourorg/your-protocols.git
```

3) Generate code for a package:
```bash
make proto-generate PROTO_PACKAGE=user
```

4) Clean generated code:
```bash
make proto-clean
```

## Handler Implementation Pattern

- Handlers live in `internal/grpc/handlers.go` (example: `UserHandlers`).
- Each handler depends on the service layer (`service.IService`).
- Register handlers inside `module.go -> registerHandlers()` (already uncommented).

### Example: CreateUser and GetUser
- Handlers validate input, call service methods, and convert models to proto using helpers in `internal/grpc/conversions.go`.
- Status codes use `google.golang.org/grpc/status` and `codes`.

## Proto Conversion Helpers

Located in `internal/grpc/conversions.go`:
- `userToProto(*models.User) *userProto.User`
- `userFromProto(*userProto.User) *models.User`
- `userStatusToProto(models.UserStatus) userProto.UserStatus`
- `userStatusFromProto(userProto.UserStatus) models.UserStatus`

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
grpcurl -plaintext -d '{"email":"test@example.com","name":"Test User"}' localhost:9090 user.v1.UserService/CreateUser
grpcurl -plaintext -d '{"email":"test@example.com"}' localhost:9090 user.v1.UserService/GetUser
```

## Adding New gRPC Services

1. Add `.proto` in `protocols/<service>/`.
2. Run `make proto-generate PROTO_PACKAGE=<service>`.
3. Add conversion helpers in `internal/grpc/conversions.go` (or a new file if distinct types).
4. Implement handlers using `service.IService` (or other dependencies) in `internal/grpc/handlers.go` or new handler files.
5. Register in `module.go -> registerHandlers()`.

## Production Notes
- Keep health checks fast (<2s).
- Adjust keep-alives and message sizes via config if needed.
- Enable TLS/mTLS by extending `GRPCConfig` (not included by default).
- For metrics/tracing, add interceptors later (kept minimal here).
