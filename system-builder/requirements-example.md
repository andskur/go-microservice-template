# System Requirements: User Management System

## Overview
Simple user management with gRPC backend and HTTP gateway.

- System name: `user-management-system`
- Root directory: `user-system`

## Protocols

Reference: follow `system-builder/protocols-guide.md` conventions (UUID as bytes, timestamps as int64 Unix, UNSPECIFIED enums).

Package: `userservice`

Models
- User: uuid (bytes), email (string), name (string), status (UserStatus), created_at (int64 Unix), updated_at (int64 Unix)
- UserStatus (enum): USER_STATUS_UNSPECIFIED=0, USER_STATUS_ACTIVE=1, USER_STATUS_DELETED=2
- EmailRequest: email (string)
- CreateUserRequest: email (string), name (string), status (UserStatus)

Services
- UserService
  - CreateUser(CreateUserRequest) returns (User)
  - GetUserByEmail(EmailRequest) returns (User)

## Microservices

### users-service (gRPC + DB)
- Modules: grpc (on), repository (on), service (on), http (off), grpc_client (off)
- Database: Postgres, db name `users_service_db`, migration for users table with unique email
- Business logic:
  - CreateUser: validate input, check email uniqueness, generate UUID, status ACTIVE, insert, return user
  - GetUserByEmail: validate email, fetch by email, return not found when missing
- Config (example env):
  - GRPC_ENABLED=true, GRPC_PORT=9090
  - DATABASE_ENABLED=true, DATABASE_HOST=postgres, DATABASE_PORT=5432, DATABASE_NAME=users_service_db, DATABASE_USER=dev, DATABASE_PASSWORD=dev
  - Modules are disabled by default; set env vars explicitly.
- Handlers: gRPC handlers for CreateUser, GetUserByEmail; conversions domain ↔ proto; proper status codes
- Tests: model validation, repository (with DB), service (mock repo), gRPC handlers (mock service)

### users-api (HTTP gateway with gRPC client)
- Modules: http (on), grpc_client (on), service (on/minimal), repository (off)
- No database
- Business logic: proxy HTTP → gRPC, map errors (not found→404, invalid→400, conflict→409, unavailable→503); handle nil grpc client with 503
- Config (example env):
  - HTTP_ENABLED=true, HTTP_PORT=8080, HTTP_MOCK_AUTH=true
  - GRPC_CLIENT_ENABLED=true, GRPC_CLIENT_ADDRESS=users-service:9090
  - DATABASE_ENABLED=false
  - Modules are disabled by default; set env vars explicitly.
- Swagger endpoints:
  - POST /users (body: email, name) → create user, returns created user
  - GET /users?email=... → get user, returns user or 404
- Tests: HTTP handlers (mock grpc client), formatter conversions, optional integration with real client

## Docker Compose

Build prep (before docker-compose):
1) Copy protocols into each service directory:
   ```bash
   cp -r protocols/ users-service/protocols/
   cp -r protocols/ users-api/protocols/
   echo "protocols/" >> users-service/.gitignore
   echo "protocols/" >> users-api/.gitignore
   ```
2) Update each Dockerfile: copy protocols first, use correct binary name in ENTRYPOINT/COPY.

Migration strategy: init container (migrate/migrate) runs before service starts.

Services
1) postgres: postgres:16-alpine, port 5432, env (db/user/password), volume, health check pg_isready
2) users-service-migrate: migrate image, runs migrations from ./users-service/db/migrations
3) users-service: build ./users-service, port 9090, depends on migrate completion; env for DB+gRPC; gRPC health check
4) users-api: build ./users-api, port 8080, depends on users-service healthy; env for HTTP+gRPC client; HTTP health check /health

Network: bridge network `microservices-network`

## Integration Tests (happy path + errors)

Scenarios
1) Create user: POST /users valid → 201 + user with UUID, status ACTIVE
2) Get user: after create, GET /users?email=... → 200 + correct data
3) Not found: GET unknown email → 404
4) Duplicate: POST same email twice → 409
5) Invalid: POST invalid email → 400

Test setup
- docker-compose up
- wait until healthy
- run tests (script or go tests hitting HTTP)
- clean DB between runs
- docker-compose down

## Success Criteria
- Both services build (`make build`)
- Unit tests pass (per service)
- docker-compose stack healthy
- Integration tests pass
- Each component is a separate git repository
