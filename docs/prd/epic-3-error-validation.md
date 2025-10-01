# Epic 3: Error Handling & Validation

**Priority**: ⭐⭐⭐⭐⭐ (Highest)
**Estimated Duration**: 1 week
**Status**: Not Started
**Dependencies**: None (hypererror), Epic 1 (hypervalidator)

---

## Overview

Implement comprehensive error handling with typed error codes and automatic HTTP/gRPC status mapping, plus request validation with automatic error conversion.

---

## Goals

- Provide type-safe error codes with multi-layer wrapping
- Automatic conversion between error codes and HTTP/gRPC status
- Rich error context with custom fields
- Request validation with clear error messages
- Seamless integration with hyperweb and hypergrpc

---

## User Stories

### Story 3.1: Error Handling (hypererror)

**As a** framework user
**I want** type-safe error handling with context
**So that** I can provide meaningful error responses across HTTP and gRPC

**Acceptance Criteria**:
- [ ] Can create typed errors with predefined codes
- [ ] Can multi-layer wrap errors with context
- [ ] Can extract error chain for debugging
- [ ] Can auto-convert to HTTP/gRPC responses
- [ ] Error fields are preserved through wrapping

**Tasks**:
- [ ] Define `Code` struct (code, HTTP status, gRPC code)
- [ ] Define predefined error code constants
- [ ] Define `Error` struct with code, message, cause, fields
- [ ] Implement `New()`, `Wrap()` constructors
- [ ] Implement convenient constructors (`NotFound()`, `BadRequest()`, etc.)
- [ ] Implement `WithField()`, `WithFields()` methods
- [ ] Implement `Error()`, `Unwrap()` (standard error interface)
- [ ] Implement `Chain()`, `Cause()` for error inspection
- [ ] Implement utility functions (`Is()`, `As()`, `HasCode()`)
- [ ] Implement `GetHTTPStatus()`, `GetGRPCCode()`
- [ ] Implement `ToResponse()` for JSON serialization
- [ ] Write unit tests (>80% coverage)
- [ ] Write godoc documentation

**Technical Details**:
```go
// Typed error code
type Code struct {
    Code       string     // "USER_NOT_FOUND"
    HTTPStatus int        // 404
    GRPCCode   codes.Code // codes.NotFound
}

// Error with context
err := hypererror.Wrap(
    hypererror.CodeInternal,
    "failed to create user",
    dbErr,
).WithField("email", email).WithField("tenant_id", tenantID)

// Multi-layer inspection
chain := err.Chain() // [layer1, layer2, layer3]
cause := err.Cause() // Original error
```

**Estimated**: 3 days

---

### Story 3.2: Request Validation (hypervalidator)

**As a** framework user
**I want** declarative request validation
**So that** I can validate input with struct tags and get clear error messages

**Acceptance Criteria**:
- [ ] Can validate structs using tags
- [ ] Validation errors automatically converted to hypererror
- [ ] Field-level error details preserved
- [ ] Support for custom validators
- [ ] Localization support for error messages

**Tasks**:
- [ ] Define `Validator` interface
- [ ] Implement based on `go-playground/validator`
- [ ] Implement error conversion (validator.ValidationErrors -> hypererror)
- [ ] Implement field path extraction
- [ ] Implement custom validator registration
- [ ] Write unit tests (>80% coverage)
- [ ] Write godoc documentation

**Technical Details**:
```go
// Declarative validation
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=32"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"gte=18,lte=120"`
}

// Automatic error conversion
if err := validator.Struct(&req); err != nil {
    // Returns hypererror with field details
    return nil, err
}
```

**Estimated**: 2 days

---

## Milestone

**Deliverable**: Production-ready error handling and validation system

**Demo Scenario**:
```go
// Service layer
func (s *UserService) CreateUser(ctx hyperctx.Context, req *CreateUserRequest) error {
    // Validate input
    if err := s.validator.Struct(req); err != nil {
        return err // Already hypererror with field details
    }

    // Business logic
    if err := s.userRepo.Create(ctx, user); err != nil {
        if errors.Is(err, gorm.ErrDuplicateKey) {
            return hypererror.Conflict("user already exists").
                WithField("email", req.Email)
        }
        return hypererror.InternalWrap("database error", err)
    }

    return nil
}

// Handler layer
func (h *UserHandler) CreateUser(c *gin.Context) {
    ctx := c.MustGet("hyperctx").(hyperctx.Context)

    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    if err := h.userService.CreateUser(ctx, &req); err != nil {
        // Automatic status code and response conversion
        status := hypererror.GetHTTPStatus(err)
        if hyperErr, ok := hypererror.As(err); ok {
            c.JSON(status, hyperErr.ToResponse())
        } else {
            c.JSON(500, gin.H{"error": "internal error"})
        }
        return
    }

    c.JSON(201, gin.H{"message": "user created"})
}
```

**Error Response Example**:
```json
{
  "code": "VALIDATION_FAILED",
  "message": "request validation failed",
  "fields": {
    "username": "must be at least 3 characters",
    "email": "must be a valid email"
  },
  "trace_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

## Technical Notes

### Architecture Decisions

- **Typed Error Codes**: Prevent string-based error code typos
- **Multi-Layer Wrapping**: Preserve error chain for debugging
- **Auto-Conversion**: Seamless HTTP/gRPC response generation

### Error Code Reference

| Code | HTTP | gRPC | Use Case |
|------|------|------|----------|
| `BAD_REQUEST` | 400 | InvalidArgument | Invalid input |
| `UNAUTHORIZED` | 401 | Unauthenticated | Auth required |
| `FORBIDDEN` | 403 | PermissionDenied | Insufficient permissions |
| `NOT_FOUND` | 404 | NotFound | Resource not found |
| `CONFLICT` | 409 | AlreadyExists | Duplicate resource |
| `VALIDATION_FAILED` | 400 | InvalidArgument | Validation failed |
| `INTERNAL_ERROR` | 500 | Internal | Server error |
| `SERVICE_UNAVAILABLE` | 503 | Unavailable | Service down |

### Dependencies

- Standard library `errors` package
- `go-playground/validator/v10` - Struct validation
- `google.golang.org/grpc/codes` - gRPC status codes

### Testing Strategy

- **Unit Tests**:
  - Error wrapping and unwrapping
  - Code mapping (HTTP/gRPC)
  - Field preservation
- **Integration Tests**:
  - Full request/response cycle
  - Error serialization

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Error info leakage in production | High | Sanitize sensitive fields |
| Complex error chains hard to debug | Medium | Implement `Chain()` method |
| Validation rule complexity | Low | Comprehensive documentation |

---

## Related Documentation

- [Architecture - hypererror](../architecture.md#52-hypererror---error-handling)
- [Architecture - hypervalidator](../architecture.md#57-hypervalidator---request-validation)
- [Error Code Reference](../architecture.md#appendix-b-error-code-reference)

---

**Last Updated**: 2025-01-XX
