# Protocol Buffers

This directory hosts protobuf definitions consumed by this microservice. The recommended source of truth is the shared protocols template repository.

## Source of Truth

Default protocols repository: `https://github.com/andskur/protocols-template.git` (configured via `PROTO_REPO` in the Makefile). Use git subtree to pull updates into `protocols/`.

## Getting Protocols (Git Subtree, Recommended)

```bash
# Add protocols from the shared repo
make proto-setup PROTO_REPO=https://github.com/andskur/protocols-template.git

# Pull updates later
make proto-update
```

## Code Generation

You can use either protoc (existing) or Buf (recommended). Ensure tools are installed first (`make proto-install` and/or `make buf-install`).

### Buf workflow (recommended)
```bash
make buf-lint                    # Lint protos
make buf-generate PROTO_PACKAGE=user   # Generate Go code for one package
make buf-generate-all            # Generate Go code for all packages
```

### Protoc workflow (traditional)
```bash
make proto-install               # Install protoc plugins
make proto-generate PROTO_PACKAGE=user  # Generate Go code for one package
make proto-generate-all          # Generate Go code for all packages
make proto-clean                 # Remove generated files
```

## Notes
- Generated `.pb.go` files should be consumed by this service but not committed back to the shared protocols repo.
- The bundled example protocols have been removed; pull the shared protocols-template instead.
- See `docs/GRPC_GUIDE.md` for detailed guidance and comparison of Buf vs. protoc workflows.
