# Epic 4: Web Service Layer

**Priority**: ⭐⭐⭐⭐⭐ (Highest)
**Estimated Duration**: 1.5 weeks
**Status**: Not Started
**Dependencies**: Epic 1 (Core Foundation), Epic 2 (Data Layer), Epic 3 (Error Handling)

---

## Overview

Implement HTTP (Gin) and gRPC servers with automatic tracing, context injection, and middleware/interceptor support.

---

## Goals

- Provide production-ready HTTP server with Gin
- Provide gRPC server with interceptors
- Automatic trace context extraction and injection
- Unified error handling for both protocols
- Graceful shutdown support

---

## User Stories

### Story 4.1: Web Server (hyperweb)

**As a** framework user
**I want** a production-ready HTTP server
**So that** I can build REST APIs with automatic tracing and error handling

**Acceptance Criteria**:
- [ ] Can start HTTP server on configured port
- [ ] Each request automatically creates `hyperctx.Context`
- [ ] Trace context automatically extracted from headers
- [ ] Automatic panic recovery with logging
- [ ] Request/response logging
- [ ] Graceful shutdown works properly
- [ ] CORS support

**Tasks**:
- [ ] Define `Server` struct
- [ ] Implement Gin engine initialization
- [ ] Implement `TraceMiddleware` (extract/inject trace context)
- [ ] Implement `RecoveryMiddleware` (panic recovery)
- [ ] Implement `LoggerMiddleware` (request/response logging)
- [ ] Implement `CORSMiddleware`
- [ ] Implement `ErrorHandlerMiddleware` (hypererror conversion)
- [ ] Implement fx lifecycle hooks (OnStart/OnStop)
- [ ] Implement graceful shutdown with timeout
- [ ] Implement health check endpoint
- [ ] Write unit tests (>80% coverage)
- [ ] Write integration tests with real HTTP server
- [ ] Write godoc documentation

**Technical Details**:
```go
// Automatic context creation
func TraceMiddleware(logger hyperlog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract trace context from headers
        ctx := otel.GetTextMapPropagator().Extract(
            c.Request.Context(),
            propagation.HeaderCarrier(c.Request.Header),
        )

        // Create hyperctx.Context
        hctx := hyperctx.New(ctx, logger, db)

        // Create root span
        hctx, end := hctx.StartSpan("handler", "HTTP", c.Request.URL.Path)
        defer end()

        // Store in Gin context
        c.Set("hyperctx", hctx)
        c.Next()
    }
}
```

**Span Hierarchy**:
```
HTTP Request /api/v1/users/123
├── handler.UserHandler.GetUser
│   └── service.UserService.GetByID
│       └── repository.UserRepository.FindByID
│           └── db.query.users
```

**Estimated**: 5 days

---

### Story 4.2: gRPC Server (hypergrpc)

**As a** framework user
**I want** a production-ready gRPC server
**So that** I can build microservices with automatic tracing

**Acceptance Criteria**:
- [ ] Can start gRPC server on configured port
- [ ] Each RPC automatically creates `hyperctx.Context`
- [ ] Trace context automatically extracted from metadata
- [ ] Automatic panic recovery
- [ ] Health check service registered
- [ ] Graceful shutdown works properly

**Tasks**:
- [ ] Define `Server` struct
- [ ] Implement gRPC server initialization
- [ ] Implement `UnaryInterceptor` (trace context + hyperctx)
- [ ] Implement `StreamInterceptor` for streaming RPCs
- [ ] Implement panic recovery interceptor
- [ ] Implement logging interceptor
- [ ] Implement error conversion (hypererror -> gRPC status)
- [ ] Implement health check service registration
- [ ] Implement fx lifecycle hooks (OnStart/OnStop)
- [ ] Implement graceful shutdown
- [ ] Write unit tests (>80% coverage)
- [ ] Write integration tests with real gRPC server
- [ ] Write godoc documentation

**Technical Details**:
```go
// Unary interceptor
func UnaryInterceptor(logger hyperlog.Logger, db *gorm.DB) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // Extract trace context from metadata
        ctx = otel.GetTextMapPropagator().Extract(ctx, ...)

        // Create hyperctx.Context
        hctx := hyperctx.New(ctx, logger, db)
        hctx, end := hctx.StartSpan("grpc", info.FullMethod, "unary")
        defer end()

        // Call handler
        resp, err := handler(hctx, req)

        // Convert hypererror to gRPC status
        if err != nil {
            if hyperErr, ok := hypererror.As(err); ok {
                return nil, status.Error(hyperErr.GRPCCode(), hyperErr.Message())
            }
        }

        return resp, err
    }
}
```

**Estimated**: 2 days

---

## Milestone

**Deliverable**: Production-ready HTTP and gRPC servers

**Demo Scenario**:

**HTTP Server**:
```go
// main.go
func main() {
    fx.New(
        hyperion.Web(),
        fx.Invoke(RegisterRoutes),
    ).Run()
}

// Handler
func RegisterRoutes(server *hyperweb.Server, handler *UserHandler) {
    engine := server.Engine()

    v1 := engine.Group("/api/v1")
    {
        v1.GET("/users/:id", handler.GetUser)
        v1.POST("/users", handler.CreateUser)
    }
}

func (h *UserHandler) GetUser(c *gin.Context) {
    ctx := c.MustGet("hyperctx").(hyperctx.Context)

    userID := c.Param("id")
    user, err := h.userService.GetByID(ctx, userID)

    if err != nil {
        status := hypererror.GetHTTPStatus(err)
        c.JSON(status, hypererror.As(err).ToResponse())
        return
    }

    c.JSON(200, user)
}
```

**gRPC Server**:
```go
func main() {
    fx.New(
        hyperion.GRPC(),
        fx.Invoke(RegisterUserService),
    ).Run()
}

func RegisterUserService(server *hypergrpc.Server, userService *service.UserService) {
    pb.RegisterUserServiceServer(server.Server(), &UserServiceImpl{
        userService: userService,
    })
}

func (s *UserServiceImpl) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    hctx := ctx.(hyperctx.Context)

    user, err := s.userService.GetByID(hctx, req.Id)
    if err != nil {
        return nil, err // Auto-converted to gRPC status
    }

    return &pb.User{Id: user.ID, Username: user.Username}, nil
}
```

---

## Technical Notes

### Architecture Decisions

- **Gin for HTTP**: High performance, mature ecosystem, extensive middleware
- **Native gRPC**: Industry standard, efficient binary protocol
- **Middleware Chain**: Trace → Recovery → Logger → CORS → Error Handler

### Middleware Order (hyperweb)

```
Request
  ↓
TraceMiddleware         # Extract trace, create hyperctx
  ↓
RecoveryMiddleware      # Panic recovery
  ↓
LoggerMiddleware        # Request logging
  ↓
CORSMiddleware          # CORS headers
  ↓
ErrorHandlerMiddleware  # Error conversion
  ↓
Handler
  ↓
Response
```

### Dependencies

- `github.com/gin-gonic/gin` - HTTP framework
- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/grpc/health/grpc_health_v1` - Health check
- `google.golang.org/grpc/reflection` - gRPC reflection

### Configuration

```yaml
web:
  host: "0.0.0.0"
  port: 8080
  mode: release  # debug, release, test
  read_timeout: 30s
  write_timeout: 30s
  max_multipart_memory: 32  # MB
  cors:
    enabled: true
    allowed_origins: ["*"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]

grpc:
  port: 9090
  max_recv_msg_size: 4  # MB
  max_send_msg_size: 4  # MB
  connection_timeout: 120s
  health_check_enabled: true
```

### Testing Strategy

- **Unit Tests**:
  - Middleware logic
  - Interceptor logic
  - Error conversion
- **Integration Tests**:
  - Full HTTP request/response cycle
  - Full gRPC request/response cycle
  - Graceful shutdown

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Middleware order dependency | Medium | Clear documentation, tests |
| Graceful shutdown timeout | Low | Configurable timeout |
| CORS misconfiguration | Medium | Safe defaults, validation |

---

## Related Documentation

- [Architecture - hyperweb](../architecture.md#59-hyperweb---web-server-gin)
- [Architecture - hypergrpc](../architecture.md#510-hypergrpc---grpc-server)
- [Quick Start Guide](../quick-start.md)

---

**Last Updated**: 2025-01-XX
