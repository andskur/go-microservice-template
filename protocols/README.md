# Protocol Buffers

This directory contains protobuf definitions for gRPC services.

## Current Structure

An example `user` service is included to demonstrate the gRPC pattern. Replace it with your own protocols for production use.

## Replacing with Your Own Protocols

### Option 1: Use Git Subtree (Recommended for shared protocols)

1. Remove example protocols:
   ```bash
   rm -rf protocols/user
   ```

2. Update `PROTO_REPO` in `Makefile` to your protocols repository.

3. Add your protocols as a subtree:
   ```bash
   make proto-setup PROTO_REPO=git@github.com:yourorg/your-protocols.git
   # or
   make proto-setup PROTO_REPO=https://github.com/yourorg/your-protocols.git
   ```

4. Update protocols from remote:
   ```bash
   make proto-update
   ```

### Option 2: Direct Replacement (Simpler for single-repo workflows)

1. Remove example protocols:
   ```bash
   rm -rf protocols/user
   ```

2. Add your own `.proto` files under `protocols/<yourservice>/`.

3. Generate Go code:
   ```bash
   make proto-generate PROTO_PACKAGE=<yourservice>
   ```

## Generating Go Code

After adding or updating `.proto` files:

```bash
# Install tools (one-time)
make proto-install

# Generate code for a specific package
make proto-generate PROTO_PACKAGE=user

# Clean generated files
make proto-clean
```

## Example Service

The included `user` service demonstrates:
- Minimal CRUD surface (CreateUser, GetUser)
- Enum handling (UserStatus)
- Timestamp usage
- UUID as bytes

See `internal/grpc/handlers.go` for handler implementation examples.
