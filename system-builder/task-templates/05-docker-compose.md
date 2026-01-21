# Task 5: Docker Compose (Template)

Objective: orchestrate all services plus dependencies.

Template Steps (fill from requirements)
- Pre-build: copy protocols into each service and update Dockerfiles to copy protocols first and use correct binary names.
- Migration strategy: use migrate init container (recommended) that runs before the service starts.
- Write docker-compose.yml with services, builds/images, env vars (explicitly enable modules), networks, volumes, health checks.
- Add .env (compose) with shared settings.
- Build and start: `docker-compose build`, `docker-compose up -d`.
- Verify: `docker-compose ps`, health endpoints/grpc health.

Verification
- All containers healthy; services can talk to each other.

Confirmation required: show compose status/logs; await user before integration tests.
