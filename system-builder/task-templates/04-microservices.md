# Task 4: Microservices (Template)

Objective: implement all services (can work in parallel) following requirements and template patterns.

Template Steps (per service)
1) Add protocols: pull/copy protocols into service, generate code.
2) Remove unused modules (explicit cleanup):
   - gRPC-only service (no HTTP):
     ```bash
     rm -rf internal/http/ api/
     rm -rf internal/grpcclient/ pkg/userservice/  # if not needed
     # Update internal/application.go: remove HTTP/grpcclient imports/registration
     go mod tidy
     make build
     ```
   - HTTP-only service (no gRPC server, no DB):
     ```bash
     rm -rf internal/grpc/
     rm -rf internal/repository/  # if no DB
     # Keep internal/grpcclient/ if proxying to other services
     # Update internal/application.go: remove gRPC server/repository registration
     go mod tidy
     make build
     ```
   - Remove stray/unused protocols and deps not needed for this service; fix imports/go.mod after cleanup.
3) Configure modules: DO NOT change defaults in config/init.go (modules stay disabled by default). Enable via env/.env/compose. If a module is required, add a startup check in application wiring that fails fast when it’s disabled, but keep defaults off.
4) Models: add domain models/enums; validation.
5) Repository (if DB): interfaces, implementations, migrations, errors, tests.
6) Service layer: business logic, error mapping, validation.
7) Handlers & conversions (critical):
   - Implement ONLY RPCs/endpoints defined in the proto/swagger contract; remove unimplemented CRUD.
   - For go-swagger HTTP handlers: match generated signatures exactly (including `principal` arg and params types).
   - Avoid casting interfaces to concrete types; accept/pass interfaces end-to-end (improves mockability).
   - If you change swagger security (e.g., remove global jwt), regenerate code; handler signatures may drop `principal` and tests/handlers must be updated. If `configure_api_*` exists, delete it before regeneration to avoid stale wiring.
   - Conversions proto ↔ domain ↔ API:
     - UUID bytes ↔ uuid.UUID: `uuid.FromBytes(pb.Uuid)` / `u.Bytes()`
     - Timestamps int64 ↔ time.Time: `time.Unix(pb.CreatedAt, 0)` / `t.Unix()`
     - Enums: mapping helpers with UNSPECIFIED default
   - Service registration: ensure gRPC service names match proto (case-sensitive) and gRPC is enabled via env.
   - Client packages: set proto `go_package` to match module path to avoid go mod tidy issues; fix imports after copying protocols.
8) Config: .env with required vars; document defaults/off-by-default modules.
9) Tests: models, repo, service, handlers, conversions; ensure mocks satisfy interfaces (add missing methods, align signatures; allow injectable behaviors for existence checks); run `make test`.
10) Build: `make build` and sanity run (`./<binary> --version`).

Verification
- Required modules wired; tests passing; builds succeed; configs present.

Confirmation required: present test/build results per service; await user before next task.
