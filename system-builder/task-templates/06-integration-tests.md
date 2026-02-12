# Task 6: Integration Tests (Template)

Objective: validate end-to-end flows across services.

Template Steps (fill from requirements)
- Define scenarios (happy path + errors) from requirements.
- Create test scripts or Go tests under integration-tests/.
- Setup: start compose, wait for health, seed if needed.
- Tests: call HTTP/gRPC endpoints, assert responses and DB effects; include cross-service orchestration (e.g., payments → goods price check, card charge).
- Teardown: clean DB, stop compose.

Include checks for
- Contract alignment: gateway/clients call only RPCs/endpoints that exist; proto/swagger in sync.
- Service discovery: gRPC services registered with correct names; GRPC_ENABLED set in compose.
- Data integrity: UUIDs non-zero; zero-value fields persisted (use_zero tags); timestamp/UUID conversions correct.
- DB robustness: pg.MaxRetries (e.g., 5) to reduce EOF under load.

Verification
- All scenarios pass; logs clean; services healthy after tests.

Confirmation required: present results; confirm completion with user.
