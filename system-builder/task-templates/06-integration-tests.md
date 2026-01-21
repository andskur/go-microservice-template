# Task 6: Integration Tests (Template)

Objective: validate end-to-end flows across services.

Template Steps (fill from requirements)
- Define scenarios (happy path + errors) from requirements.
- Create test scripts or Go tests under integration-tests/.
- Setup: start compose, wait for health, seed if needed.
- Tests: call HTTP/gRPC endpoints, assert responses and DB effects.
- Teardown: clean DB, stop compose.

Verification
- All scenarios pass; logs clean; services healthy after tests.

Confirmation required: present results; confirm completion with user.
