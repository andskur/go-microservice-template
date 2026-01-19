# HTTP/Swagger Module Guide

This guide explains how to use the HTTP REST API module in the microservice template. The module uses go-swagger for spec-first API development with OpenAPI/Swagger 2.0.

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Configuration](#configuration)
4. [Swagger Specification](#swagger-specification)
5. [Code Generation](#code-generation)
6. [Handler Implementation](#handler-implementation)
7. [Model Conversions](#model-conversions)
8. [Middleware](#middleware)
9. [Authentication](#authentication)
10. [Testing](#testing)
11. [Common Patterns](#common-patterns)
12. [Production Considerations](#production-considerations)

## Overview

The HTTP module provides:
- **Spec-first development**: Define your API in `api/swagger.yaml`, generate Go server code
- **Standard module lifecycle**: Init → Start → Stop → HealthCheck
- **Middleware chain**: Recovery, logging, CORS, rate limiting
- **JWT authentication**: Bearer token validation with mock mode
- **Error mapping**: Service errors → HTTP status codes
- **Type-safe conversions**: Domain models ↔ API models

### Architecture

```
┌─────────────────────────────────────────────────┐
│  HTTP Request                                   │
└──────────────────┬──────────────────────────────┘
                   │
         ┌─────────▼─────────┐
         │  Middleware Chain │
         │  - Recovery       │
         │  - Logger         │
         │  - CORS           │
         │  - RateLimit      │
         └─────────┬─────────┘
                   │
         ┌─────────▼─────────┐
         │  Authentication   │
         │  JWT / Mock       │
         └─────────┬─────────┘
                   │
         ┌─────────▼─────────┐
         │  Handler          │
         │  - Validate       │
         │  - Call Service   │
         │  - Format Response│
         └─────────┬─────────┘
                   │
         ┌─────────▼─────────┐
         │  Service Layer    │
         └─────────┬─────────┘
                   │
         ┌─────────▼─────────┐
         │  Repository       │
         └───────────────────┘
```

## Quick Start

### 1. Enable HTTP Module

```bash
export HTTP_ENABLED=true
export HTTP_ADDRESS="0.0.0.0:8080"
export HTTP_MOCK_AUTH=true  # For local development
```

### 2. Run the Service

```bash
make run
```

### 3. Test Endpoints

```bash
# Health check (no authentication required)
curl http://localhost:8080/health

# Get user by email (requires JWT token in mock mode)
curl -H "Authorization: Bearer test-token" \
  "http://localhost:8080/users?email=test@example.com"
```

## Configuration

### Full Configuration Example

```yaml
http:
  enabled: true
  address: "0.0.0.0:8080"
  timeout: "30s"
  swagger_spec: "./api/swagger.yaml"
  mock_auth: false
  admin_emails:
    - "admin@example.com"
  
  cors:
    enabled: true
    allowed_origins:
      - "https://myapp.com"
      - "https://app.myapp.com"
    allowed_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "OPTIONS"
      - "PATCH"
    allowed_headers:
      - "*"
    max_age: 3600
  
  rate_limit:
    enabled: true
    requests_per_sec: 100.0
    burst: 20
  
  gatekeeper:
    address: "localhost:9091"
    timeout: "5s"
```

### Environment Variables

Configuration can be overridden with environment variables (dots become underscores):

```bash
HTTP_ENABLED=true
HTTP_ADDRESS="0.0.0.0:8080"
HTTP_TIMEOUT="30s"
HTTP_MOCK_AUTH=true
HTTP_ADMIN_EMAILS="admin1@example.com,admin2@example.com"

# CORS
HTTP_CORS_ENABLED=true
HTTP_CORS_ALLOWED_ORIGINS="https://myapp.com,https://app.myapp.com"

# Rate limiting
HTTP_RATE_LIMIT_ENABLED=true
HTTP_RATE_LIMIT_REQUESTS_PER_SEC=100.0
HTTP_RATE_LIMIT_BURST=20

# Gatekeeper
HTTP_GATEKEEPER_ADDRESS="localhost:9091"
HTTP_GATEKEEPER_TIMEOUT="5s"
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | false | Enable HTTP module |
| `address` | string | "0.0.0.0:8080" | Host:port to listen on |
| `timeout` | string | "30s" | Request timeout |
| `swagger_spec` | string | "./api/swagger.yaml" | Path to swagger spec |
| `mock_auth` | bool | false | Use mock authentication |
| `admin_emails` | []string | [] | Admin user emails |
| `cors.enabled` | bool | true | Enable CORS |
| `cors.allowed_origins` | []string | ["*"] | Allowed origins |
| `cors.allowed_methods` | []string | ["GET",...] | Allowed HTTP methods |
| `cors.allowed_headers` | []string | ["*"] | Allowed headers |
| `cors.max_age` | int | 3600 | Preflight cache duration (seconds) |
| `rate_limit.enabled` | bool | false | Enable rate limiting |
| `rate_limit.requests_per_sec` | float64 | 100.0 | Requests per second |
| `rate_limit.burst` | int | 20 | Burst size |
| `gatekeeper.address` | string | "localhost:9091" | Gatekeeper gRPC address |
| `gatekeeper.timeout` | string | "5s" | Gatekeeper request timeout |

## Swagger Specification

### Current Spec Structure

The `api/swagger.yaml` uses Swagger 2.0 format:

```yaml
swagger: "2.0"
info:
  title: Microservice Template API
  version: 1.0.0

paths:
  /users:
    get:
      operationId: getUserByEmail
      parameters:
        - name: email
          in: query
          type: string
          format: email
      responses:
        200:
          schema:
            $ref: "#/definitions/User"
```

### Adding a New Endpoint

**Step 1**: Add to `api/swagger.yaml`

```yaml
paths:
  /users/{id}:
    get:
      summary: Get user by UUID
      operationId: getUserByID
      tags:
        - users
      security:
        - jwt: []
      parameters:
        - name: id
          in: path
          required: true
          type: string
          format: uuid
      responses:
        200:
          description: User found
          schema:
            $ref: "#/definitions/User"
        404:
          description: User not found
          schema:
            $ref: "#/definitions/Error"
```

**Step 2**: Regenerate code

```bash
make generate-api
```

**Step 3**: Implement handler (see [Handler Implementation](#handler-implementation))

**Step 4**: Register handler in `internal/http/module.go` → `initAPI()`

```go
api.UsersGetUserByIDHandler = handlers.NewGetUserByID(m.service)
```

### Defining Models

```yaml
definitions:
  CreateUserRequest:
    type: object
    required:
      - email
      - name
    properties:
      email:
        type: string
        format: email
        description: User email address
      name:
        type: string
        minLength: 1
        maxLength: 255
        description: User full name
      status:
        type: string
        enum:
          - active
          - deleted
        default: active
```

## Code Generation

### Generate API Code

```bash
# Validate swagger spec
make swagger-validate

# Generate server code from spec
make generate-api

# Clean generated code
make swagger-clean
```

### Generated Structure

```
internal/http/
├── models/              # Generated API models
│   ├── error.go
│   ├── health.go
│   └── user.go
└── server/              # Generated server code (gitignored)
    ├── server.go        # HTTP server
    ├── operations/      # Operation handlers
    │   ├── health/
    │   └── users/
    └── embedded_spec.go # Embedded swagger spec
```

### What to Edit vs Generated

**DO NOT EDIT** (regenerated on `make generate-api`):
- `internal/http/server/**`
- `internal/http/models/**`

**DO EDIT** (your implementation):
- `internal/http/handlers/**` - Handler implementations
- `internal/http/formatter/**` - Model conversions
- `internal/http/auth/**` - Authentication logic
- `internal/http/middlewares/**` - Middleware implementations
- `internal/http/module.go` - Module wiring

## Handler Implementation

### Handler Pattern

**File**: `internal/http/handlers/users.go`

```go
package handlers

import (
    "context"
    "net/http"
    
    "github.com/go-openapi/runtime/middleware"
    
    "microservice-template/internal/http/formatter"
    "microservice-template/internal/http/models"
    "microservice-template/internal/http/server/operations/users"
    "microservice-template/internal/service"
    "microservice-template/pkg/logger"
)

// NewGetUserByEmail creates new handler
func NewGetUserByEmail(svc service.IService) *GetUserByEmail {
    return &GetUserByEmail{service: svc}
}

// GetUserByEmail handler
type GetUserByEmail struct {
    service service.IService
}

// Handle processes the request
func (h *GetUserByEmail) Handle(
    params users.GetUserByEmailParams,
    principal *models.User,
) middleware.Responder {
    // 1. Extract and validate parameters
    email := string(params.Email)
    if email == "" {
        return users.NewGetUserByEmailBadRequest().
            WithPayload(DefaultError(http.StatusBadRequest, 
                service.ErrInvalidInput, nil))
    }
    
    // 2. Call service layer
    ctx := context.Background()
    user, err := h.service.GetUserByEmail(ctx, email)
    if err != nil {
        logger.Log().Errorf("get user by email %s: %s", email, err.Error())
        
        // 3. Map service errors to HTTP status codes
        switch {
        case service.IsInvalidInput(err):
            return users.NewGetUserByEmailBadRequest().
                WithPayload(DefaultError(http.StatusBadRequest, err, nil))
        case service.IsNotFound(err):
            return users.NewGetUserByEmailNotFound().
                WithPayload(DefaultError(http.StatusNotFound, err, nil))
        case service.IsRepositoryUnavailable(err):
            return users.NewGetUserByEmailServiceUnavailable().
                WithPayload(DefaultError(http.StatusServiceUnavailable, err, nil))
        default:
            return users.NewGetUserByEmailInternalServerError().
                WithPayload(DefaultError(http.StatusInternalServerError, err, nil))
        }
    }
    
    // 4. Convert domain model to API model and return
    return users.NewGetUserByEmailOK().WithPayload(formatter.UserToAPI(user))
}
```

### Error Mapping

Map service errors to HTTP status codes:

```go
// Service errors → HTTP status codes
service.ErrInvalidInput           → 400 Bad Request
service.ErrNotFound               → 404 Not Found
service.ErrUnauthorized           → 401 Unauthorized
service.ErrForbidden              → 403 Forbidden
service.ErrConflict               → 409 Conflict
service.ErrRepositoryUnavailable  → 503 Service Unavailable
Unknown errors                    → 500 Internal Server Error
```

### POST/PUT Handler Example

```go
func (h *CreateUser) Handle(
    params users.CreateUserParams,
    principal *models.User,
) middleware.Responder {
    // Convert API model to domain model
    domainUser := formatter.UserFromAPI(params.Body)
    
    // Validate
    if err := domainUser.Validate(); err != nil {
        return users.NewCreateUserBadRequest().
            WithPayload(DefaultError(http.StatusBadRequest, err, nil))
    }
    
    // Create user
    ctx := context.Background()
    if err := h.service.CreateUser(ctx, domainUser); err != nil {
        // Handle errors...
        return users.NewCreateUserInternalServerError().
            WithPayload(DefaultError(http.StatusInternalServerError, err, nil))
    }
    
    // Return created user
    return users.NewCreateUserCreated().
        WithPayload(formatter.UserToAPI(domainUser))
}
```

## Model Conversions

### Formatter Pattern

**File**: `internal/http/formatter/user.go`

```go
package formatter

import (
    "github.com/go-openapi/strfmt"
    "github.com/gofrs/uuid"
    
    apimodels "microservice-template/internal/http/models"
    "microservice-template/internal/models"
)

// UserToAPI converts domain User to API User
func UserToAPI(user *models.User) *apimodels.User {
    email := strfmt.Email(user.Email)
    name := user.Name
    status := user.Status.String()
    
    return &apimodels.User{
        UUID:      strfmt.UUID(user.UUID.String()),
        Email:     &email,
        Name:      &name,
        Status:    &status,
        CreatedAt: strfmt.DateTime(user.CreatedAt),
        UpdatedAt: strfmt.DateTime(user.UpdatedAt),
    }
}

// UserFromAPI converts API User to domain User
func UserFromAPI(apiUser *apimodels.User) *models.User {
    status, _ := models.UserStatusFromString(*apiUser.Status)
    
    return &models.User{
        UUID:   uuid.FromStringOrNil(apiUser.UUID.String()),
        Email:  string(*apiUser.Email),
        Name:   *apiUser.Name,
        Status: status,
    }
}
```

### Type Conversions

| Domain Type | API Type | Conversion |
|-------------|----------|------------|
| `string` | `strfmt.Email` | `strfmt.Email(str)` / `string(email)` |
| `uuid.UUID` | `strfmt.UUID` | `strfmt.UUID(id.String())` / `uuid.FromStringOrNil(string(id))` |
| `time.Time` | `strfmt.DateTime` | `strfmt.DateTime(t)` / `time.Time(dt)` |
| `UserStatus` | `string` | `status.String()` / `UserStatusFromString(str)` |

## Middleware

### Middleware Chain

Middleware is applied in order (alice chain):

```go
Recovery() → Logger() → Cors() → RateLimit() → Handler
```

### Recovery Middleware

Catches panics and returns 500 errors:

```go
// Automatically logs stack traces
// Returns: {"code":500,"message":"internal server error"}
```

### Logger Middleware

Logs all requests:

```
HTTP GET /users?email=test@example.com 200 45.2ms
HTTP POST /users 201 123.5ms
HTTP GET /users?email=notfound@example.com 404 12.3ms
```

### CORS Middleware

Configure CORS in `config.yaml`:

```yaml
http:
  cors:
    enabled: true
    allowed_origins:
      - "https://myapp.com"
    allowed_methods:
      - "GET"
      - "POST"
    allowed_headers:
      - "Authorization"
      - "Content-Type"
    max_age: 3600
```

Handles preflight OPTIONS requests automatically.

### Rate Limit Middleware

Token bucket rate limiter per IP address:

```yaml
http:
  rate_limit:
    enabled: true
    requests_per_sec: 100.0  # 100 requests per second
    burst: 20                 # Allow burst of 20
```

Returns `429 Too Many Requests` when limit exceeded.

## Authentication

### Mock Mode (Development)

For local development without gatekeeper:

```yaml
http:
  mock_auth: true
```

Mock mode returns a test user for any Bearer token:
- UUID: `fa734dc4-22e6-41c5-a913-30c302c1ca68`
- Email: `test@example.com`
- Name: `Test User`
- Status: `active`

### Gatekeeper Integration

See detailed TODO in `internal/http/auth/auth.go` for integration steps.

**Summary**:
1. Add gatekeeper gRPC client dependency
2. Initialize client in `NewAuth()`
3. Call `ValidateToken()` in `CheckAuth()`
4. Handle errors (invalid, expired, network)
5. Add connection pooling and retry logic
6. Consider token caching for performance

### Admin Role Checking

Configure admin emails:

```yaml
http:
  admin_emails:
    - "admin@example.com"
    - "superadmin@example.com"
```

Check in handlers:

```go
if !m.auth.IsAdmin(*principal.Email) {
    return operations.NewDeleteUserForbidden().
        WithPayload(DefaultError(http.StatusForbidden, 
            service.ErrForbidden, nil))
}
```

### Protected vs Public Endpoints

**Protected** (requires JWT):
```yaml
paths:
  /users:
    get:
      security:
        - jwt: []  # Requires authentication
```

**Public** (no auth):
```yaml
paths:
  /health:
    get:
      security: []  # No authentication required
```

## Testing

### Unit Tests

Test handlers with mock service:

```go
func TestGetUserByEmail_Success(t *testing.T) {
    // Setup mock
    svc := &mockService{
        getUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
            return &models.User{
                UUID:  uuid.Must(uuid.NewV4()),
                Email: email,
                Name:  "Test User",
            }, nil
        },
    }
    
    handler := NewGetUserByEmail(svc)
    
    // Execute
    params := users.GetUserByEmailParams{
        Email: strfmt.Email("test@example.com"),
    }
    responder := handler.Handle(params, principal)
    
    // Assert
    okResponse, ok := responder.(*users.GetUserByEmailOK)
    if !ok {
        t.Fatalf("expected success response, got %T", responder)
    }
}
```

### Testing with curl

```bash
# Health check
curl -v http://localhost:8080/health

# Get user (with auth)
curl -v \
  -H "Authorization: Bearer test-token" \
  "http://localhost:8080/users?email=test@example.com"

# Test error cases
curl -v \
  -H "Authorization: Bearer test-token" \
  "http://localhost:8080/users?email=notfound@example.com"
```

### Run Tests

```bash
# Run all HTTP tests
make test-http

# Run specific package
go test -v ./internal/http/handlers/...

# Run with coverage
go test -coverprofile=coverage.out ./internal/http/...
go tool cover -html=coverage.out
```

## Common Patterns

### Pagination

**Swagger**:
```yaml
parameters:
  - name: offset
    in: query
    type: integer
    default: 0
  - name: limit
    in: query
    type: integer
    default: 20
    maximum: 100
```

**Handler**:
```go
offset := int(params.Offset)
limit := int(params.Limit)
if limit > 100 {
    limit = 100
}

users, total, err := h.service.ListUsers(ctx, offset, limit)
```

### Filtering

```yaml
parameters:
  - name: status
    in: query
    type: string
    enum: [active, deleted]
  - name: created_after
    in: query
    type: string
    format: date-time
```

### Sorting

```yaml
parameters:
  - name: order_by
    in: query
    type: string
    enum: [created_at, updated_at, email]
    default: created_at
  - name: order
    in: query
    type: string
    enum: [asc, desc]
    default: desc
```

### Batch Operations

```yaml
paths:
  /users/batch:
    post:
      operationId: createUsersBatch
      parameters:
        - name: body
          in: body
          schema:
            type: object
            properties:
              users:
                type: array
                items:
                  $ref: "#/definitions/CreateUserRequest"
```

## Production Considerations

### Security

1. **TLS/HTTPS**: Add TLS configuration to generated server
2. **CORS**: Use specific origins, not `*`
3. **Rate Limiting**: Tune based on expected load
4. **Input Validation**: Validate all inputs in handlers
5. **Secrets**: Never log sensitive data (passwords, tokens)

### Performance

1. **Connection Pooling**: Reuse HTTP client connections
2. **Timeouts**: Set appropriate timeouts for all operations
3. **Caching**: Cache validated JWT tokens
4. **Compression**: Enable gzip for responses
5. **Keep-Alive**: Enable HTTP keep-alive

### Monitoring

1. **Metrics**: Add Prometheus metrics middleware
2. **Tracing**: Add distributed tracing (Jaeger, Zipkin)
3. **Health Checks**: Implement real health checks
4. **Logging**: Use structured logging with request IDs

### Scalability

1. **Horizontal Scaling**: Run multiple instances behind load balancer
2. **Rate Limiting**: Implement distributed rate limiting (Redis)
3. **Caching**: Use distributed cache (Redis, Memcached)
4. **Database**: Connection pooling and query optimization

### Example Production Config

```yaml
http:
  enabled: true
  address: "0.0.0.0:8080"
  timeout: "30s"
  mock_auth: false
  admin_emails:
    - "admin@company.com"
  
  cors:
    enabled: true
    allowed_origins:
      - "https://app.company.com"
      - "https://admin.company.com"
    allowed_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
    allowed_headers:
      - "Authorization"
      - "Content-Type"
      - "X-Request-ID"
    max_age: 3600
  
  rate_limit:
    enabled: true
    requests_per_sec: 1000.0
    burst: 50
  
  gatekeeper:
    address: "gatekeeper.company.internal:9091"
    timeout: "5s"
```

## Troubleshooting

### Common Issues

**Issue**: `make generate-api` fails
```bash
# Solution: Validate spec first
make swagger-validate
```

**Issue**: Handler not found
```bash
# Solution: Ensure handler is registered in initAPI()
api.UsersGetUserByEmailHandler = handlers.NewGetUserByEmail(m.service)
```

**Issue**: CORS errors in browser
```bash
# Solution: Check allowed origins and enable CORS
http:
  cors:
    enabled: true
    allowed_origins:
      - "https://yourapp.com"
```

**Issue**: Rate limit too strict
```bash
# Solution: Increase limits or disable for development
http:
  rate_limit:
    enabled: false  # or increase requests_per_sec
```

## Further Reading

- [Swagger 2.0 Specification](https://swagger.io/specification/v2/)
- [go-swagger Documentation](https://goswagger.io/)
- [api/README.md](../api/README.md) - Swagger spec guide
- [MODULE_DEVELOPMENT.md](./MODULE_DEVELOPMENT.md) - Module patterns
