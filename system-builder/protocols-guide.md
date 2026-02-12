# Protocols Guide (Patterns & Conventions)

Use these conventions when defining protos so services, generators, and conversions stay consistent.

## Types

- **UUID**: use `bytes` fields (simple, efficient). Convert with `uuid.FromBytes()` / `uuid.UUID.Bytes()`.
- **Timestamps**: prefer `int64` Unix seconds for simplicity; convert with `time.Unix()` / `t.Unix()`. If you need RFC3339 or nanos, use `google.protobuf.Timestamp` explicitly.
- **Status Enums**: zero value is `UNSPECIFIED`; follow `PREFIX_STATUS_*` style. Map to domain enums with helper functions.
- **Strings/Numbers**: keep required fields explicit in validation, not via `required` keyword (proto3).

## Naming

- Messages: PascalCase (`User`, `CreateUserRequest`).
- RPCs: verbs + nouns (`CreateUser`, `GetUserByEmail`, `ListUsers`).
- Enums: UPPER_SNAKE with prefix (`USER_STATUS_ACTIVE`).
- Packages: single word, lowercase (`userservice`).

## Field Numbering

- Reserve 1–15 for frequently used fields (smaller varint). Do not reuse numbers.
- Keep wire compatibility: never repurpose a number; use new numbers when adding fields.

## Imports

- Standard: `google/protobuf/timestamp.proto` if you choose Timestamp; otherwise use `int64` and avoid the import.
- Custom common types: only introduce if reused across many services; otherwise inline simple fields.
- Keep include paths minimal; verify buf/protoc include dirs.

## Message Patterns (example snippets)

```proto
message User {
  bytes uuid = 1;              // uuid.UUID
  string email = 2;            // validated in service layer
  string name = 3;
  UserStatus status = 4;       // enum below
  int64 created_at = 5;        // Unix seconds
  int64 updated_at = 6;        // Unix seconds
}

enum UserStatus {
  USER_STATUS_UNSPECIFIED = 0;
  USER_STATUS_ACTIVE = 1;
  USER_STATUS_DELETED = 2;
}

message EmailRequest {
  string email = 1;
}

message CreateUserRequest {
  string email = 1;
  string name = 2;
  UserStatus status = 3;
}
```

## Service Patterns

- Keep RPCs unary unless streaming is required.
- Use clear verb names; prefer narrow RPCs (Create, Get, List, Update, Delete).
- Use gRPC status codes via server handlers (map domain errors to codes).

## Conversion Patterns (Go)

UUID (bytes ↔ uuid.UUID):
```go
u, err := uuid.FromBytes(pb.Uuid)
```

Timestamp (int64 ↔ time.Time):
```go
t := time.Unix(pb.CreatedAt, 0)
```

Enum mapping (proto ↔ domain):
```go
func StatusFromProto(s proto.UserStatus) (models.UserStatus, error) { /* switch with default */ }
func StatusToProto(s models.UserStatus) proto.UserStatus { /* switch with default */ }
```

Null Handling:
- Avoid optional unless needed; represent optional fields via validation rules and zero values.
- If optional is required, use `optional` and handle presence checks explicitly.

## Common Pitfalls

- Missing imports: ensure buf.yaml/protoc include paths cover any common packages you reference.
- Reusing field numbers: never repurpose; add new fields with new numbers.
- UUID wrappers: avoid custom UUID messages unless truly needed; bytes are simpler.
- Timestamp mismatch: decide on int64 vs Timestamp up front and document it.

## Checklist for New Protos

- [ ] Zero-value enum is UNSPECIFIED.
- [ ] UUID is bytes (or documented otherwise).
- [ ] Timestamp format chosen and documented.
- [ ] Field numbers unique and stable.
- [ ] Imports minimal and resolvable.
- [ ] RPCs use clear verbs.
- [ ] Conversion helpers planned (UUID, timestamp, enums).

## Package Versioning (Required)

- Always version packages with `.v1` (and increment to v2, v3, ... when making breaking changes).
- Directory must match package:
  - `package userservice.v1` → file path `userservice/v1/<file>.proto`
  - `package common.v1` → path `common/v1/<file>.proto`
- Set go_package accordingly, e.g.:
  ```proto
  // File: protocols/userservice/v1/user.proto
  syntax = "proto3";
  package userservice.v1;
  option go_package = "your-project/protocols/userservice/v1;userservicev1";
  ```

### buf.yaml (root of protocols)
Recommended minimal config:
```yaml
version: v1
breaking:
  use:
    - FILE
lint:
  use:
    - DEFAULT
```

### buf Lint Expectations
- Will enforce versioned packages.
- Will enforce directory matching package.
- Flags unused imports—remove them.
- Follow the errors; don't disable linting.
