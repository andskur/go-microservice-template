# Task 3: Protocols (Template)

Objective: define protos per requirements and generate code.

Before starting: read `system-builder/protocols-guide.md` for type conventions (UUID bytes, timestamps, enums, numbering).

Template Steps (fill from requirements)
- Create package directories under protocols repo.
- Write .proto files: models (all fields), enums (UNSPECIFIED zero), services with RPCs using clear verbs.
- Keep imports minimal; choose timestamp format (int64 vs Timestamp) and document choice.
- Configure buf (if used); run generation (buf or protoc) with Makefile targets.
- Verify generated Go code compiles (imports resolve) and commit proto + buf configs (not generated code).

Verification
- Proto files follow conventions; generation succeeds; git clean.

Confirmation required: show proto definitions and generation result, await user before next task.
