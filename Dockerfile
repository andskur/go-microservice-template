# Builder: Go toolchain pinned by digest (golang:1.26.11).
FROM golang:1.26.11@sha256:4165ee6bf3b7811fd7b40373290405c83d0167dcb87ee5f29c427cdcef45ff8f AS builder

WORKDIR /app

# Install the pinned code generation tools (swagger, buf, protoc plugins).
# This layer is cached until the Makefile changes.
COPY Makefile go.mod go.sum ./
RUN make tools

# Download Go dependencies (cached until go.mod/go.sum change).
RUN go mod download

# Copy the sources. Generated code is not committed, so it is generated here;
# `git clone && docker build .` therefore works without any local tooling.
# .git is copied (when present) so `make build` can embed version information.
COPY . .
RUN make generate && make build

# Minimal runtime image.
FROM scratch

# CA certificates for outbound TLS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app/microservice-template /microservice-template

# Run as an unprivileged user.
USER 65534:65534

ENTRYPOINT ["/microservice-template"]
CMD ["serve"]
