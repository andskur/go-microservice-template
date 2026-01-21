# Compact Agent Prompt (Template Already Cloned)

```markdown
# Task: Build {SYSTEM_NAME} Microservices System

## Before Starting
Read these template docs (in order):
1. `AGENTS.md`
2. `system-builder/protocols-guide.md`
3. `system-builder/common-issues.md`
4. `system-builder/task-templates/` (00-06)
5. Skim: `docs/MODULE_DEVELOPMENT.md`, `docs/GRPC_GUIDE.md`, `docs/HTTP_SWAGGER_GUIDE.md`

## Requirements
Use a short `requirements.md` (keep it concise):

```
# System Requirements

## Overview
- Name: <system-name>
- Root dir: <root-dir>

## Protocols
- Package(s): <name>
- Models: list fields (uuid bytes, timestamps int64, enums with UNSPECIFIED)
- Services/RPCs: Create*, Get*, List*, etc.

## Microservices (one per section)
- Name: <service>
- Modules: (grpc/http/grpc_client/repository) on/off
- Database: yes/no; migrations needed
- Business logic: bullets for main flows
- Config: required env vars (.env.example)
- Handlers: gRPC/HTTP endpoints to implement

## Docker Compose
- Services: postgres?, migrate init?, service containers
- Build prep: copy protocols into each service; Dockerfile binary name
- Env: explicitly enable modules (GRPC_ENABLED, HTTP_ENABLED, DATABASE_ENABLED, GRPC_CLIENT_ENABLED)

## Integration Tests
- Scenarios: happy path + error cases (list)
```

## Workflow
### Phase 1: Planning
Create detailed tasks in `tasks/`:
- `02-scaffolding.md`
- `03-protocols.md`
- `04-microservices.md`
- `05-docker-compose.md`
- `06-integration-tests.md`

STOP after writing tasks → present summary → **request user approval**.

### Phase 2: Execution (tasks 2→6)
For each task: present → **request confirmation** → execute → show results → **request confirmation** before next.

## Critical Points
- **Module removal** (avoid build errors):
  - gRPC-only: `rm -rf internal/http/ api/ internal/grpcclient/`
  - HTTP-only: `rm -rf internal/grpc/ internal/repository/` (keep grpcclient if proxying)
  - Update `internal/application.go`; `go mod tidy && make build`
- **Type conversions**:
  - UUID: `uuid.FromBytes(pb.Uuid)` / `u.Bytes()`
  - Time: `time.Unix(pb.CreatedAt, 0)` / `t.Unix()`
  - Enums: StatusFromProto/StatusToProto helpers
- **Docker prep**:
  - `cp -r protocols/ <service>/protocols/`
  - Dockerfile: copy protocols first; correct binary names
  - Compose: explicitly enable modules (GRPC_ENABLED, HTTP_ENABLED, DATABASE_ENABLED, GRPC_CLIENT_ENABLED)
- **Migrations**: use migrate init container before service starts.

## Success Criteria
- All services build & tests pass
- docker-compose services healthy
- Integration tests pass
- Each component in its own git repo

## Reminders
- Modules default to disabled → enable via env
- Remove unused modules
- Pre-copy protocols before docker build
- Wait for user confirmations

Start with Phase 1: create detailed tasks based on `requirements.md`.
```
