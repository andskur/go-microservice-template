# Common Issues & Solutions

Quick reference for frequent build/runtime problems.

## Protocol Issues

**Cannot find import**
- Cause: Include paths missing.
- Fix: Ensure buf/protoc include directories cover referenced paths.

**UUID type mismatch**
- Cause: Proto uses bytes; Go uses uuid.UUID.
- Fix:
```go
u, err := uuid.FromBytes(pb.Uuid)
```

**Timestamp mismatch**
- Cause: Proto int64 vs Go time.Time.
- Fix:
```go
```

**Enum mismatch**
- Cause: Zero value not handled or mapping missing.
- Fix: Add switch mapping with UNSPECIFIED default.

## Build Issues

**Unused imports / module code**
- Cause: Unused HTTP/GRPC/repository code present.
- Fix: Remove unused module directories; update application.go; run `go mod tidy`.

**Binary not found in Docker**
- Cause: ENTRYPOINT name mismatch with built binary.
- Fix: Align Dockerfile COPY/ENTRYPOINT with Makefile APP name.

**Cannot access ../protocols during build**
- Cause: Docker build context too narrow.
- Fix: Copy protocols/ into service dir before docker build; copy it early in Dockerfile.

**Go version mismatch**
- Cause: go.mod version newer than base image.
- Fix: Align go.mod with image (e.g., go 1.24; use golang:1.24 base).

## Runtime Issues

**Module not enabled**
- Cause: Defaults are disabled.
- Fix: Set env vars explicitly (e.g., GRPC_ENABLED=true, DATABASE_ENABLED=true, HTTP_ENABLED=true, GRPC_CLIENT_ENABLED=true).

**Table does not exist**
- Cause: Migrations not run.
- Fix: Add migrate init container or run `make migrate-up` from host before start.

**401 with mock_auth**
- Cause: Swagger middleware still checks Authorization header presence.
- Fix: Send `Authorization: Bearer mock-token` even in mock mode.

## Docker Issues

**Service exits immediately**
- Cause: Missing CMD or module disabled.
- Fix: Check logs; ensure modules enabled and entrypoint correct.

**Health check failing**
- Cause: Wrong port or protocol.
- Fix: Verify GRPC vs HTTP health probes and exposed ports.
