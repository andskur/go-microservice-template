# Microservices System Builder

Framework for building microservices systems using `go-microservice-template` and `protocols-template`. It defines a clear, repeatable workflow: explore templates → create detailed tasks → execute with confirmations.

## Directory

```
system-builder/
├── README.md
├── requirements-example.md
├── prompt-template.md
├── protocols-guide.md
├── common-issues.md
└── task-templates/
    ├── 00-explore.md
    ├── 01-understand-requirements.md
    ├── 02-scaffolding.md
    ├── 03-protocols.md
    ├── 04-microservices.md
    ├── 05-docker-compose.md
    └── 06-integration-tests.md
```

## Workflow

1) **Phase 0: Explore** (automatic)
   - Clone and read: `go-microservice-template`, `protocols-template`
   - Read `AGENTS.md`, `README.md`, `docs/*` (module, gRPC, HTTP guides)
2) **Phase 1: Plan**
   - Read `requirements.md`
   - Fill task templates 02-06 with concrete steps
   - STOP for user approval
3) **Phase 2: Execute** (tasks 2→6)
   - Before each task: present summary, request confirmation
   - After each task: show results, request confirmation to continue

## Requirements Format (text, Markdown)

Use natural language; include:
- System overview (name, description, root directory)
- Protocols: packages, models, enums, services + RPCs
- Microservices: modules on/off, DB needs, business logic, config, handlers, tests
- Docker compose: services, networks, health checks
- Integration tests: scenarios and expected outcomes

See `requirements-example.md` for a complete sample (user management system).

## Prompt Template

`prompt-template.md` gives ready-to-use instructions for the agent: explore → plan (approval) → execute (per-task confirmations).

## Task Templates

Located in `task-templates/`. Task 1 fills 02-06 based on requirements.
- 00: Explore templates (no confirmation)
- 01: Understand requirements & create detailed tasks (requires approval after)
- 02: Scaffolding
- 03: Protocols
- 04: Microservices
- 05: Docker Compose
- 06: Integration Tests

## Success Criteria (per prompt)
- All services build (`make build`)
- Unit tests pass (`make test` per service)
- docker-compose brings up healthy stack
- Integration tests pass end-to-end
- Each component has its own git repository
- All requirements implemented

## Example Output Structure

```
<system-root>/
├── protocols/            (git repo)
├── <service-a>/          (git repo)
├── <service-b>/          (git repo)
├── docker-compose.yml
└── integration-tests/
```

## Tips
- Be explicit in requirements; more detail ⇒ better tasks.
- Follow template patterns for modules, handlers, repos, and tests.
- Use Makefile targets; skip lints initially; focus on correctness.
- Keep user in the loop: approval after planning, confirmation before each task run.
- See `protocols-guide.md` for proto conventions and `common-issues.md` for quick fixes.
