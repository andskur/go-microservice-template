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
3) Configure modules: update scheme/init defaults if adding fields; modules remain disabled by default (enable via env); create .env.example with required vars.
4) Models: add domain models/enums; validation.
5) Repository (if DB): interfaces, implementations, migrations, errors, tests.
6) Service layer: business logic, error mapping, validation.
7) Handlers & conversions (critical): gRPC/HTTP handlers; conversions proto ↔ domain ↔ API.
   - UUID bytes ↔ uuid.UUID: `uuid.FromBytes(pb.Uuid)` / `u.Bytes()`
   - Timestamps int64 ↔ time.Time: `time.Unix(pb.CreatedAt, 0)` / `t.Unix()`
   - Enums: mapping helpers with UNSPECIFIED default
8) Config: .env with required vars; document defaults/off-by-default modules.
9) Tests: models, repo, service, handlers, conversions; run `make test`.
10) Build: `make build` and sanity run (`./<binary> --version`).

Verification
- Required modules wired; tests passing; builds succeed; configs present.

Confirmation required: present test/build results per service; await user before next task.
