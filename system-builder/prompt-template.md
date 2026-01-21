# Task: Build Microservices System

## Context
You are starting in a NEW empty directory. First, explore these templates:
- Microservice: https://github.com/andskur/go-microservice-template
- Protocols: https://github.com/andskur/protocols-template

Read AGENTS.md, README.md, and docs/ in the template to learn patterns.

## Requirements
Read `requirements.md` (text/Markdown) provided for the target system. It describes protocols, microservices, docker-compose, and integration tests.

## Workflow

### Phase 0: Exploration (automatic, no confirmation)
- Clone and read both templates
- Read `system-builder/protocols-guide.md` (type conventions and conversions)
- Skim `system-builder/common-issues.md` (known pitfalls)
- Study: module lifecycle, gRPC patterns, HTTP/swagger patterns, repository patterns, Makefile targets, testing guidance

### Phase 1: Planning
**Task 1**: Create detailed tasks from `requirements.md` into `tasks/`:
- `tasks/02-scaffolding.md`
- `tasks/03-protocols.md`
- `tasks/04-microservices.md`
- `tasks/05-docker-compose.md`
- `tasks/06-integration-tests.md`

Then STOP, present the tasks, and **request user approval** before any execution.

### Phase 2: Execution (tasks 2→6)
For each task:
1) Present task summary
2) **Request user confirmation to start**
3) Execute
4) Show results (tests/build status)
5) **Request user confirmation to continue**

Order: 02 Scaffolding → 03 Protocols → 04 Microservices → 05 Docker Compose → 06 Integration Tests.

## Guidelines
- Follow template patterns (modules, handlers, repos, conversions, tests)
- Use Makefile targets; skip lint initially; focus on functionality
- Run unit tests after implementing each service
- Each component (protocols, services) is its own git repo
- Always wait for user confirmation between execution tasks
- Modules default to disabled; explicitly enable via env vars in compose/.env

## Success Criteria
- All services build (`make build`)
- Unit tests pass (`make test`)
- docker-compose stack healthy
- Integration tests pass end-to-end
- All requirements in `requirements.md` implemented

## Start
Begin with **Phase 0** (explore templates), then Phase 1 (plan tasks and pause for approval).
