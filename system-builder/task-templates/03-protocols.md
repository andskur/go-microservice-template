# Task 3: Protocols (Template)

Objective: define protos per requirements and generate code.

Before starting: read `system-builder/protocols-guide.md` (type conventions, versioning, directory layout).

Template Steps (fill from requirements)
- Create versioned package directories: e.g., `protocols/userservice/v1/` (package must be `userservice.v1`).
- Write .proto files with versioned package and matching path; set go_package accordingly.
- Keep imports minimal; remove unused imports; choose timestamp format (int64 vs Timestamp) and document choice.
- Configure buf at protocols root (buf.yaml) with lint/breaking defaults.
- Run buf as primary tool: `buf lint` then `buf generate` (or `make buf-generate-all`).
- Verify generated Go code compiles (imports resolve); commit proto + buf configs (not generated code).

Verification
- Versioned packages (.v1), paths match packages, lint passes, generation succeeds, git clean.

Confirmation required: show proto definitions and buf lint/generate result; await user before next task.
