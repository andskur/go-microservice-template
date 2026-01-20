# Swagger API Specification Guide

This directory contains the OpenAPI 3.0 specification for the microservice template HTTP API. The specification follows a **spec-first** approach: you define the API contract in `swagger.yaml`, then generate Go server code using go-swagger.

## Quick Start

```bash
# Validate the specification
make swagger-validate

# Generate server code
make generate-api

# Clean generated code (if needed)
make swagger-clean
```

## Swagger Spec Structure

### OpenAPI 3.0 Sections

1. **info**: API metadata (title, version, description, contact)
2. **servers**: API server URLs (development, staging, production)
3. **security**: Global security requirements (JWT by default)
4. **paths**: API endpoints (operations, parameters, responses)
5. **components**: Reusable definitions (schemas, security schemes)

### Current Endpoints

| Method | Path       | Description           | Auth Required |
|--------|------------|-----------------------|---------------|
| GET    | /users     | Get user by email     | ✅ JWT        |
| GET    | /health    | Health check          | ❌ No         |

## Adding a New Endpoint

### Step 1: Define the Path

Add a new path to the `paths` section:

```yaml
paths:
  /users/{id}:
    get:
      summary: Get user by UUID
      description: Retrieves a user by their unique identifier
      operationId: getUserByID
      tags:
        - users
      security:
        - jwt: []
      parameters:
        - name: id
          in: path
          required: true
          description: User UUID
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: User found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: User not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
```

### Step 2: Define Request/Response Models

If you need new models, add them to `components/schemas`:

```yaml
components:
  schemas:
    CreateUserRequest:
      type: object
      required:
        - email
        - name
      properties:
        email:
          type: string
          format: email
        name:
          type: string
          minLength: 1
          maxLength: 255
        status:
          type: string
          enum:
            - active
            - deleted
          default: active
```

### Step 3: Regenerate Code

```bash
make generate-api
```

This generates:
- `internal/http/server/models/` - API model types
- `internal/http/server/restapi/operations/` - Handler interfaces
- `internal/http/server/restapi/` - Server configuration

### Step 4: Implement the Handler

Create handler in `internal/http/handlers/`:

```go
// internal/http/handlers/user_by_id.go
package handlers

import (
    "net/http"
    
    "github.com/go-openapi/runtime/middleware"
    "github.com/gofrs/uuid"
    
    "microservice-template/internal/http/formatter"
    "microservice-template/internal/http/server/models"
    "microservice-template/internal/http/server/restapi/operations/users"
    "microservice-template/internal/service"
    "microservice-template/pkg/logger"
)

func NewGetUserByID(svc service.IService) *GetUserByID {
    return &GetUserByID{service: svc}
}

type GetUserByID struct {
    service service.IService
}

func (h *GetUserByID) Handle(params users.GetUserByIDParams, principal *models.User) middleware.Responder {
    userUUID, err := uuid.FromString(params.ID.String())
    if err != nil {
        logger.Log().Errorf("parse user UUID: %s", err.Error())
        return users.NewGetUserByIDDefault(http.StatusBadRequest).
            WithPayload(handlers.DefaultError(http.StatusBadRequest, err, nil))
    }
    
    // Call service layer
    user, err := h.service.GetUserByUUID(ctx, userUUID)
    if err != nil {
        // Handle errors with appropriate HTTP codes
        return handleServiceError(err, users.NewGetUserByIDDefault)
    }
    
    return users.NewGetUserByIDOK().WithPayload(formatter.UserToAPI(user))
}
```

### Step 5: Register Handler

In `internal/http/api.go`, register the handler:

```go
api.UsersGetUserByIDHandler = handlers.NewGetUserByID(m.service)
```

## Parameter Types

### Path Parameters

```yaml
parameters:
  - name: id
    in: path
    required: true
    schema:
      type: string
      format: uuid
```

### Query Parameters

```yaml
parameters:
  - name: email
    in: query
    required: true
    schema:
      type: string
      format: email
  - name: limit
    in: query
    required: false
    schema:
      type: integer
      minimum: 1
      maximum: 100
      default: 20
```

### Request Body

```yaml
requestBody:
  required: true
  content:
    application/json:
      schema:
        $ref: '#/components/schemas/CreateUserRequest'
```

### Header Parameters

```yaml
parameters:
  - name: X-Request-ID
    in: header
    required: false
    schema:
      type: string
      format: uuid
```

## Data Types & Formats

### Common Types

| Type    | Format       | Example                              | Validation          |
|---------|--------------|--------------------------------------|---------------------|
| string  | -            | "hello"                              | minLength, maxLength|
| string  | email        | "user@example.com"                   | RFC 5322            |
| string  | uuid         | "550e8400-e29b-41d4-a716-446655440000" | UUID v4          |
| string  | date-time    | "2024-01-20T10:30:00Z"               | RFC 3339            |
| integer | int32        | 123                                  | minimum, maximum    |
| integer | int64        | 9223372036854775807                  | minimum, maximum    |
| number  | float        | 3.14                                 | minimum, maximum    |
| boolean | -            | true                                 | -                   |
| array   | -            | ["a", "b", "c"]                      | minItems, maxItems  |
| object  | -            | {"key": "value"}                     | required, properties|

### Enums

```yaml
status:
  type: string
  enum:
    - active
    - deleted
    - suspended
  default: active
```

### Arrays

```yaml
tags:
  type: array
  items:
    type: string
  minItems: 1
  maxItems: 10
  uniqueItems: true
```

### Objects

```yaml
address:
  type: object
  required:
    - street
    - city
  properties:
    street:
      type: string
    city:
      type: string
    postal_code:
      type: string
      pattern: '^\d{5}(-\d{4})?$'
```

## Validation Rules

### String Validation

```yaml
email:
  type: string
  format: email
  minLength: 5
  maxLength: 255
  pattern: '^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$'
```

### Number Validation

```yaml
age:
  type: integer
  minimum: 18
  maximum: 120
  exclusiveMinimum: false

price:
  type: number
  format: float
  minimum: 0
  exclusiveMinimum: true  # price > 0
  multipleOf: 0.01  # precision to 2 decimal places
```

### Array Validation

```yaml
emails:
  type: array
  items:
    type: string
    format: email
  minItems: 1
  maxItems: 10
  uniqueItems: true
```

## Security Definitions

### JWT Bearer Token

```yaml
components:
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: JWT token from gatekeeper service

# Apply globally
security:
  - jwt: []

# Override for specific endpoint (no auth)
paths:
  /health:
    get:
      security: []  # Public endpoint
```

### API Key (Example)

```yaml
components:
  securitySchemes:
    apiKey:
      type: apiKey
      in: header
      name: X-API-Key
```

## Response Codes

### Success Responses

| Code | Meaning          | Usage                                    |
|------|------------------|------------------------------------------|
| 200  | OK               | Successful GET/PUT/PATCH                 |
| 201  | Created          | Successful POST (resource created)       |
| 204  | No Content       | Successful DELETE (no response body)     |

### Client Error Responses

| Code | Meaning          | Usage                                    |
|------|------------------|------------------------------------------|
| 400  | Bad Request      | Invalid input (validation failed)        |
| 401  | Unauthorized     | Missing/invalid authentication           |
| 403  | Forbidden        | Authenticated but insufficient perms     |
| 404  | Not Found        | Resource doesn't exist                   |
| 409  | Conflict         | Resource already exists (duplicate)      |
| 422  | Unprocessable    | Semantically invalid (valid syntax)      |
| 429  | Too Many Requests| Rate limit exceeded                      |

### Server Error Responses

| Code | Meaning          | Usage                                    |
|------|------------------|------------------------------------------|
| 500  | Internal Error   | Unexpected server error                  |
| 503  | Service Unavailable | Service/dependency down (e.g., DB disabled) |

## Common Patterns

### Pagination

```yaml
parameters:
  - name: offset
    in: query
    schema:
      type: integer
      minimum: 0
      default: 0
  - name: limit
    in: query
    schema:
      type: integer
      minimum: 1
      maximum: 100
      default: 20

responses:
  '200':
    content:
      application/json:
        schema:
          type: object
          properties:
            items:
              type: array
              items:
                $ref: '#/components/schemas/User'
            total:
              type: integer
            offset:
              type: integer
            limit:
              type: integer
```

### Filtering

```yaml
parameters:
  - name: status
    in: query
    schema:
      type: string
      enum: [active, deleted]
  - name: created_after
    in: query
    schema:
      type: string
      format: date-time
```

### Sorting

```yaml
parameters:
  - name: order_by
    in: query
    schema:
      type: string
      enum: [created_at, updated_at, email]
      default: created_at
  - name: order
    in: query
    schema:
      type: string
      enum: [asc, desc]
      default: desc
```

### File Upload

```yaml
requestBody:
  required: true
  content:
    multipart/form-data:
      schema:
        type: object
        properties:
          file:
            type: string
            format: binary
          description:
            type: string
```

## Best Practices

### 1. Use Descriptive Operation IDs

```yaml
# Good
operationId: getUserByEmail

# Bad
operationId: get1
```

### 2. Provide Examples

```yaml
schema:
  type: object
  properties:
    email:
      type: string
      example: "user@example.com"
```

### 3. Document All Fields

```yaml
properties:
  email:
    type: string
    format: email
    description: User email address used for login and notifications
    example: "user@example.com"
```

### 4. Use References for Reusability

```yaml
# Define once
components:
  schemas:
    User:
      type: object
      properties:
        # ...

# Reuse everywhere
responses:
  '200':
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/User'
```

### 5. Version Your API

```yaml
info:
  version: 1.0.0

servers:
  - url: https://api.example.com/v1
```

### 6. Use Tags for Organization

```yaml
tags:
  - name: users
    description: User management operations
  - name: health
    description: Health and monitoring

paths:
  /users:
    get:
      tags:
        - users
```

## Validation & Testing

### Validate Spec

```bash
make swagger-validate
```

### Test with Swagger UI

1. Generate code: `make generate-api`
2. Run service: `make run`
3. Visit: `http://localhost:8080/swagger` (if swagger UI is enabled)

### Test with curl

```bash
# Health check (no auth)
curl http://localhost:8080/health

# Get user (with JWT)
export TOKEN="your-jwt-token"
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/users?email=user@example.com"
```

### Test with Postman

1. Import `swagger.yaml` into Postman
2. Configure environment variables (base URL, JWT token)
3. Test all endpoints

## Troubleshooting

### Validation Errors

```bash
# Error: Invalid reference
Error: path /users references unknown schema

# Solution: Check $ref path matches exactly
$ref: '#/components/schemas/User'
```

### Generation Errors

```bash
# Error: duplicate operationId
Error: duplicate operation ID: getUser

# Solution: Use unique operationId for each endpoint
operationId: getUserByEmail
operationId: getUserByID
```

### Handler Not Found

```bash
# Error: handler not registered
panic: no handler for operation UsersGetUserByEmail

# Solution: Register handler in api.go
api.UsersGetUserByEmailHandler = handlers.NewGetUserByEmail(m.service)
```

## Further Reading

- [OpenAPI 3.0 Specification](https://swagger.io/specification/)
- [go-swagger Documentation](https://goswagger.io/)
- [Swagger Editor](https://editor.swagger.io/) - Online editor for testing specs
- [HTTP Status Codes](https://httpstatuses.com/) - Complete HTTP status code reference
