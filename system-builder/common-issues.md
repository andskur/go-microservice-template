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
t := time.Unix(pb.CreatedAt, 0)
```

**Enum mismatch**
- Cause: Zero value not handled or mapping missing.
- Fix: Add switch mapping with UNSPECIFIED default.

**buf lint: package not versioned**
- Cause: Package missing .v1 suffix.
- Fix: Change `package userservice` to `package userservice.v1` (and update go_package).

**buf lint: directory mismatch**
- Cause: File path doesn’t match package.
- Fix: `package userservice.v1` → place file in `userservice/v1/`.

**buf lint: unused import**
- Cause: Import declared but not used.
- Fix: Remove the unused import.

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

**Handlers/contract mismatch**
- Cause: Implementing RPCs/HTTP endpoints not defined in proto/swagger (extra CRUD), or wrong method signatures.
- Fix: Implement only the contract; remove unsupported handlers and align signatures to proto.

**HTTP handler signature mismatch (go-swagger)**
- Cause: Manual handlers don’t match generated signature (missing `principal` arg, wrong params type).
- Fix: Match generated signature exactly, including principal (`func (h *X) Handle(params <op>Params, principal *apimodels.User) ...`).

**Interface vs concrete (mockability)**
- Cause: Casting interfaces to concrete modules (e.g., `grpcclient.IClient` → `*grpcclient.Module`) blocks mocks.
- Fix: Accept and pass interfaces end-to-end; avoid concrete type assertions in handlers/modules.

**Incomplete mocks**
- Cause: Mock doesn’t implement all methods of the interface (Name/Init/HealthCheck or RPCs).
- Fix: Ensure mock implements full interface; add no-op methods as needed.

**Swagger response constructors not generated**
- Cause: Using response helpers not defined in swagger (e.g., specific 503 constructor missing).
- Fix: Use generated constructors only; fall back to available responses (e.g., InternalServerError) or update swagger to declare the response.

**Swagger global security still enforces auth (401 with mock_auth)**
- Cause: `security: - jwt: []` at root in swagger.yaml forces Authorization header before reaching custom auth.
- Fix: Remove/adjust global security or send Authorization header; if making endpoints public, delete security requirements and regenerate code.

**Handler signature changes after security removal**
- Cause: Removing security removes `principal` from generated handler signatures.
- Fix: Update handlers/tests to match new signatures; regenerate code.

**Stale configure_api_*.go after regeneration**
- Cause: go-swagger does not overwrite `configure_api_*` files if they exist.
- Fix: Delete old configure_api_* file, regenerate (`make generate-api`).

**Legacy/template code lingering**
- Cause: Unused modules or foreign protocols (e.g., userservice code left in other services).
- Fix: Delete unused dirs/protocols; fix imports/go.mod; rebuild.

**Unknown gRPC service / discovery failures**
- Cause: Service not registered (name mismatch), or gRPC server/module not enabled (GRPC_ENABLED missing in env/compose).
- Fix: Ensure service registration name matches proto (case-sensitive); enable gRPC module via env; check compose env vars.

**go_package mismatch / go mod tidy fails on generated code**
- Cause: Generated proto go_package paths don’t match repo module path.
- Fix: Set go_package correctly in proto before generation; regenerate; avoid post-generation sed hacks when possible.

**Protocol drift between services**
- Cause: Gateway expects RPC/endpoints not implemented by backend (or vice versa).
- Fix: Align swagger/proto with actual services; either implement missing RPCs/endpoints or update specs/handlers to match what exists.

**gRPC conversion bugs (zero UUID)**
- Cause: Wrong byte/string conversions in handlers/converters.
- Fix: Use `uuid.UUID.Bytes()` and `uuid.FromBytes()` consistently.

**Postgres EOF / connection churn**
- Cause: High-frequency requests without retries.
- Fix: Set pg.Options `MaxRetries` (e.g., 5).

**Zero-value inserts ignored**
- Cause: go-pg skips zero values without `use_zero` tag.
- Fix: Add struct tag `pg:",use_zero"` for fields that can be zero.

**Charge vs Deposit semantics**
- Cause: Card service only accepts positive deposits; cannot charge.
- Fix: Align contract: add explicit withdraw/charge RPC, or relax validation to allow negative amounts if reusing a single RPC—keep proto/handlers consistent.

**Swagger global security still enforces auth (401 with mock_auth)**
- Cause: `security: - jwt: []` at root in swagger.yaml forces Authorization header before reaching custom auth.
- Fix: Remove/adjust global security or send Authorization header; if making endpoints public, delete security requirements and regenerate code.

**Handler signature changes after security removal**
- Cause: Removing security removes `principal` from generated handler signatures.
- Fix: Update handlers/tests to match new signatures; regenerate code.

**Stale configure_api_*.go after regeneration**
- Cause: go-swagger does not overwrite `configure_api_*` files if they exist.
- Fix: Delete old configure_api_* file, regenerate (`make generate-api`).

**Legacy/template code lingering**
- Cause: Unused modules or foreign protocols (e.g., userservice code left in other services).
- Fix: Delete unused dirs/protocols; fix imports/go.mod; rebuild.

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
