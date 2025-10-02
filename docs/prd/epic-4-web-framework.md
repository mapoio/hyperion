# Epic 4: Web Framework (v2.3)

**Version**: 2.3
**Status**: 🔜 **PLANNED** (April 2026)
**Duration**: 5 weeks
**Priority**: ⭐⭐⭐⭐⭐

---

## Overview

Implement production-ready **Gin web adapter** with comprehensive middleware suite, enabling rapid development of RESTful APIs and full-stack web applications while maintaining zero lock-in and framework flexibility.

---

## Goals

### Primary Goals
1. Gin web framework adapter with full `hyperion.Context` integration
2. Production middleware suite (logging, recovery, tracing, metrics)
3. Request/Response validation and binding
4. Graceful shutdown and health checks
5. Full-stack example application with frontend integration

### Success Criteria
- [ ] Gin adapter seamlessly integrates with hyperion.Context
- [ ] All middleware properly propagate context and observability
- [ ] Request validation covers 100% of common use cases
- [ ] Health check endpoints follow Kubernetes standards
- [ ] Example app demonstrates production-ready patterns
- [ ] Performance: handle 10,000+ req/s with <10ms p99 latency

---

## Deliverables

### 1. Gin Web Adapter 🔜

**Package**: `adapter/gin/`

**Scope**:
- Gin router wrapped with hyperion abstractions
- hyperion.Context integration in handlers
- Automatic request/response binding
- Router group support
- Static file serving
- Template rendering support

**Core Implementation**:
```go
type GinServer struct {
    engine *gin.Engine
    config ServerConfig
}

// Handler type that uses hyperion.Context
type HandlerFunc func(ctx hyperion.Context) error

// Wrapper to convert Gin context to hyperion.Context
func (s *GinServer) Handle(method, path string, handler HandlerFunc) {
    s.engine.Handle(method, path, func(c *gin.Context) {
        // Create hyperion.Context from gin.Context
        hctx := s.createHyperionContext(c)

        // Execute handler
        if err := handler(hctx); err != nil {
            s.handleError(c, err)
            return
        }

        // Response already written by handler via hctx.JSON(), etc.
    })
}

// Response helpers integrated into hyperion.Context
type GinContext interface {
    hyperion.Context

    // Response methods
    JSON(code int, obj any)
    XML(code int, obj any)
    String(code int, format string, values ...any)
    HTML(code int, name string, data any)

    // Request methods
    Bind(obj any) error
    ShouldBind(obj any) error
    Param(key string) string
    Query(key string) string
    PostForm(key string) string

    // File handling
    SaveUploadedFile(file *multipart.FileHeader, dst string) error
    FormFile(name string) (*multipart.FileHeader, error)
}
```

**Module Export**:
```go
var Module = fx.Module("hyperion.adapter.gin",
    fx.Provide(
        NewGinServer,
        NewRouterGroup,
    ),
    fx.Invoke(RegisterShutdownHook),
)
```

**Configuration Example**:
```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: release           # debug, release, test
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  max_header_bytes: 1048576
  enable_pprof: false     # Enable pprof endpoints
  trusted_proxies:
    - 10.0.0.0/8
```

**Usage Example**:
```go
type UserHandler struct {
    userService *service.UserService
}

func (h *UserHandler) GetUser(ctx hyperion.Context) error {
    // Automatic tracing via hyperion.Context
    ctx, span := ctx.Tracer().Start(ctx, "UserHandler.GetUser")
    defer span.End()

    // Extract path parameter
    userID := ctx.Param("id")

    // Call service layer
    user, err := h.userService.GetUser(ctx, userID)
    if err != nil {
        ctx.Logger().Error("failed to get user", "error", err)
        return err  // Automatic error handling
    }

    // Return JSON response
    ctx.JSON(200, user)
    return nil
}

// Router registration
func RegisterRoutes(server *gin.GinServer) {
    userHandler := &UserHandler{userService: ...}

    api := server.Group("/api/v1")
    api.GET("/users/:id", userHandler.GetUser)
    api.POST("/users", userHandler.CreateUser)
    api.PUT("/users/:id", userHandler.UpdateUser)
    api.DELETE("/users/:id", userHandler.DeleteUser)
}
```

**Tasks**:
- [ ] Implement GinServer struct (2 days)
- [ ] Create hyperion.Context wrapper (2 days)
- [ ] Add request/response helpers (1 day)
- [ ] Add router group support (1 day)
- [ ] Integrate with Config (1 day)
- [ ] Add graceful shutdown (1 day)
- [ ] Write unit tests (>80% coverage) (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 7 working days

---

### 2. Middleware Suite 🔜

**Package**: `adapter/gin/middleware/`

**Scope**:
- Request logging with structured fields
- Panic recovery with stack traces
- Distributed tracing propagation
- Prometheus metrics collection
- CORS support
- Rate limiting
- Authentication/Authorization helpers

#### 2.1 Request Logging Middleware

```go
func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        hctx := getHyperionContext(c)

        start := time.Now()
        path := c.Request.URL.Path
        method := c.Request.Method

        // Process request
        c.Next()

        // Log after request
        latency := time.Since(start)
        statusCode := c.Writer.Status()

        hctx.Logger().Info("request completed",
            "method", method,
            "path", path,
            "status", statusCode,
            "latency_ms", latency.Milliseconds(),
            "client_ip", c.ClientIP(),
            "user_agent", c.Request.UserAgent(),
        )
    }
}
```

#### 2.2 Recovery Middleware

```go
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                hctx := getHyperionContext(c)

                // Log panic with stack trace
                hctx.Logger().Error("panic recovered",
                    "error", err,
                    "stack", string(debug.Stack()),
                )

                // Record in trace
                if span := trace.SpanFromContext(c.Request.Context()); span.IsRecording() {
                    span.RecordError(fmt.Errorf("panic: %v", err))
                    span.SetAttributes(attribute.Bool("panic", true))
                }

                // Return error response
                c.JSON(500, gin.H{
                    "error": "Internal Server Error",
                    "trace_id": hctx.TraceID(),
                })
                c.Abort()
            }
        }()

        c.Next()
    }
}
```

#### 2.3 Tracing Middleware

```go
func Tracing(serviceName string) gin.HandlerFunc {
    return func(c *gin.Context) {
        hctx := getHyperionContext(c)

        // Extract trace context from headers (W3C Trace Context)
        ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

        // Start span
        spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
        ctx, span := hctx.Tracer().Start(ctx, spanName)
        defer span.End()

        // Set span attributes
        span.SetAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.url", c.Request.URL.String()),
            attribute.String("http.client_ip", c.ClientIP()),
        )

        // Update request context
        c.Request = c.Request.WithContext(ctx)

        // Process request
        c.Next()

        // Record response status
        span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))

        // Inject trace context into response headers
        otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))
    }
}
```

#### 2.4 Metrics Middleware

```go
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "hyperion_http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "hyperion_http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "path"},
    )
)

func Metrics() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.FullPath()
        method := c.Request.Method

        c.Next()

        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())

        httpRequestsTotal.WithLabelValues(method, path, status).Inc()
        httpRequestDuration.WithLabelValues(method, path).Observe(duration)
    }
}
```

#### 2.5 CORS Middleware

```go
func CORS(config CORSConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")

        if config.AllowOrigin(origin) {
            c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
            c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
            c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
            c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
            c.Writer.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
        }

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

#### 2.6 Rate Limiting Middleware

```go
func RateLimit(limiter *rate.Limiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !limiter.Allow() {
            hctx := getHyperionContext(c)
            hctx.Logger().Warn("rate limit exceeded", "client_ip", c.ClientIP())

            c.JSON(429, gin.H{
                "error": "Too Many Requests",
                "retry_after": limiter.Reserve().Delay().Seconds(),
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

**Configuration Example**:
```yaml
middleware:
  logging:
    enabled: true
    skip_paths:
      - /health
      - /metrics

  cors:
    enabled: true
    allowed_origins:
      - https://example.com
      - https://app.example.com
    allowed_methods:
      - GET
      - POST
      - PUT
      - DELETE
    allowed_headers:
      - Authorization
      - Content-Type
    max_age: 3600

  rate_limit:
    enabled: true
    requests_per_second: 100
    burst: 200
```

**Tasks**:
- [ ] Implement RequestLogger middleware (1 day)
- [ ] Implement Recovery middleware (1 day)
- [ ] Implement Tracing middleware (1 day)
- [ ] Implement Metrics middleware (1 day)
- [ ] Implement CORS middleware (1 day)
- [ ] Implement RateLimit middleware (1 day)
- [ ] Write middleware tests (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 5 working days

---

### 3. Validation & Binding 🔜

**Package**: `adapter/gin/validation/`

**Scope**:
- Request validation using `go-playground/validator`
- Custom validation rules
- Automatic error response formatting
- Support for JSON, XML, Form, Query binding

**Validation Integration**:
```go
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=32,alphanum"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8,max=72"`
    Age      int    `json:"age" binding:"gte=18,lte=120"`
}

func (h *UserHandler) CreateUser(ctx hyperion.Context) error {
    var req CreateUserRequest

    // Automatic validation
    if err := ctx.Bind(&req); err != nil {
        // Returns structured validation errors
        return err
    }

    user, err := h.userService.CreateUser(ctx, req.Username, req.Email, req.Password)
    if err != nil {
        return err
    }

    ctx.JSON(201, user)
    return nil
}
```

**Custom Validators**:
```go
// Register custom validator
func init() {
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterValidation("username", validateUsername)
        v.RegisterValidation("strong_password", validateStrongPassword)
    }
}

func validateUsername(fl validator.FieldLevel) bool {
    username := fl.Field().String()
    // Custom validation logic
    return regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`).MatchString(username)
}
```

**Error Response Format**:
```json
{
  "error": "Validation Failed",
  "fields": {
    "username": "must be between 3 and 32 characters",
    "email": "must be a valid email address",
    "password": "must be at least 8 characters"
  },
  "trace_id": "abc123xyz"
}
```

**Tasks**:
- [ ] Integrate go-playground/validator (1 day)
- [ ] Implement custom validators (1 day)
- [ ] Create validation error formatter (1 day)
- [ ] Add common validation rules (1 day)
- [ ] Write validation tests (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 3 working days

---

### 4. Health Checks & Graceful Shutdown 🔜

**Package**: `adapter/gin/health/`

**Scope**:
- Liveness and readiness endpoints (Kubernetes-compatible)
- Dependency health checks (database, cache, external services)
- Graceful shutdown with connection draining
- Startup/shutdown lifecycle hooks

**Health Check Implementation**:
```go
type HealthChecker struct {
    db    hyperion.Database
    cache hyperion.Cache
}

func (h *HealthChecker) Liveness(ctx hyperion.Context) error {
    // Simple check - server is running
    ctx.JSON(200, gin.H{
        "status": "UP",
        "timestamp": time.Now().Unix(),
    })
    return nil
}

func (h *HealthChecker) Readiness(ctx hyperion.Context) error {
    checks := make(map[string]string)

    // Check database
    if err := h.db.Ping(ctx); err != nil {
        checks["database"] = "DOWN"
    } else {
        checks["database"] = "UP"
    }

    // Check cache
    if err := h.cache.Exists(ctx, "health-check"); err != nil {
        checks["cache"] = "DOWN"
    } else {
        checks["cache"] = "UP"
    }

    // Overall status
    allHealthy := true
    for _, status := range checks {
        if status == "DOWN" {
            allHealthy = false
            break
        }
    }

    statusCode := 200
    if !allHealthy {
        statusCode = 503
    }

    ctx.JSON(statusCode, gin.H{
        "status": map[bool]string{true: "UP", false: "DOWN"}[allHealthy],
        "checks": checks,
        "timestamp": time.Now().Unix(),
    })
    return nil
}
```

**Graceful Shutdown**:
```go
func (s *GinServer) Start(lc fx.Lifecycle) {
    srv := &http.Server{
        Addr:           fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
        Handler:        s.engine,
        ReadTimeout:    s.config.ReadTimeout,
        WriteTimeout:   s.config.WriteTimeout,
        IdleTimeout:    s.config.IdleTimeout,
        MaxHeaderBytes: s.config.MaxHeaderBytes,
    }

    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            go func() {
                if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                    s.logger.Error("server failed", "error", err)
                }
            }()
            s.logger.Info("server started", "addr", srv.Addr)
            return nil
        },
        OnStop: func(ctx context.Context) error {
            s.logger.Info("shutting down server gracefully...")

            // Create shutdown context with timeout
            shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
            defer cancel()

            // Graceful shutdown
            if err := srv.Shutdown(shutdownCtx); err != nil {
                s.logger.Error("server shutdown failed", "error", err)
                return err
            }

            s.logger.Info("server stopped")
            return nil
        },
    })
}
```

**Kubernetes Integration**:
```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: app
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10

        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5

        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 15"]
```

**Tasks**:
- [ ] Implement health check endpoints (1 day)
- [ ] Add dependency health checks (1 day)
- [ ] Implement graceful shutdown (1 day)
- [ ] Add lifecycle hooks (1 day)
- [ ] Write health check tests (1 day)
- [ ] Documentation and Kubernetes examples (1 day)

**Timeline**: 3 working days

---

### 5. Full-Stack Example Application 🔜

**Package**: `examples/fullstack-app/`

**Scope**:
- Production-ready RESTful API backend
- React/Vue frontend (SPA)
- Complete CRUD operations
- Authentication (JWT)
- File upload/download
- WebSocket support (optional)
- Docker deployment

**Features**:
- User authentication and authorization
- Product catalog management
- Order processing
- Real-time notifications (WebSocket)
- File upload (product images)
- Admin dashboard

**Technology Stack**:
```
Backend:
- hyperion.CoreModule
- adapter/viper (Config)
- adapter/zap (Logger)
- adapter/gorm (Database)
- adapter/otel (Tracing)
- adapter/redis (Cache)
- adapter/gin (Web)

Frontend:
- React 18 + TypeScript
- TailwindCSS
- React Query (API client)
- React Router
- Zustand (State management)
```

**Backend Architecture**:
```
examples/fullstack-app/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handler/
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── product_handler.go
│   │   └── order_handler.go
│   ├── service/
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── product_service.go
│   │   └── order_service.go
│   ├── repository/
│   │   ├── user_repository.go
│   │   ├── product_repository.go
│   │   └── order_repository.go
│   └── domain/
│       ├── user.go
│       ├── product.go
│       └── order.go
├── web/                    # Frontend
│   ├── src/
│   ├── public/
│   └── package.json
├── docker-compose.yml
├── Dockerfile
└── README.md
```

**Example Handler with Full Middleware Stack**:
```go
func NewRouter(
    server *gin.GinServer,
    authHandler *handler.AuthHandler,
    userHandler *handler.UserHandler,
    productHandler *handler.ProductHandler,
) {
    // Global middleware
    server.Use(
        middleware.Recovery(),
        middleware.RequestLogger(),
        middleware.Tracing("fullstack-app"),
        middleware.Metrics(),
        middleware.CORS(corsConfig),
    )

    // Public routes
    server.POST("/api/auth/login", authHandler.Login)
    server.POST("/api/auth/register", authHandler.Register)

    // Health checks
    server.GET("/health/live", healthChecker.Liveness)
    server.GET("/health/ready", healthChecker.Readiness)

    // Protected routes
    api := server.Group("/api", middleware.AuthRequired())
    {
        // User routes
        users := api.Group("/users")
        users.GET("/:id", userHandler.GetUser)
        users.PUT("/:id", userHandler.UpdateUser)

        // Product routes (admin only)
        products := api.Group("/products", middleware.RequireRole("admin"))
        products.POST("", productHandler.CreateProduct)
        products.PUT("/:id", productHandler.UpdateProduct)
        products.DELETE("/:id", productHandler.DeleteProduct)

        // Order routes
        orders := api.Group("/orders")
        orders.GET("", orderHandler.ListOrders)
        orders.POST("", orderHandler.CreateOrder)
        orders.GET("/:id", orderHandler.GetOrder)
    }

    // WebSocket endpoint
    server.GET("/ws", websocketHandler.HandleWebSocket)

    // Static files (frontend)
    server.Static("/", "./web/dist")
}
```

**Authentication Example**:
```go
func (h *AuthHandler) Login(ctx hyperion.Context) error {
    var req LoginRequest
    if err := ctx.Bind(&req); err != nil {
        return err
    }

    user, err := h.authService.Authenticate(ctx, req.Email, req.Password)
    if err != nil {
        return hyperion.NewError(401, "Invalid credentials")
    }

    // Generate JWT
    token, err := h.jwtService.GenerateToken(user)
    if err != nil {
        return err
    }

    ctx.JSON(200, LoginResponse{
        Token: token,
        User:  user,
    })
    return nil
}

func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        hctx := getHyperionContext(c)

        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }

        claims, err := jwtService.ValidateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        // Add user info to context
        c.Set("user_id", claims.UserID)
        c.Set("user_role", claims.Role)

        c.Next()
    }
}
```

**Docker Deployment**:
```dockerfile
# Multi-stage build
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY web/ .
RUN npm install && npm run build

FROM alpine:latest
WORKDIR /app
COPY --from=backend-builder /app/server .
COPY --from=frontend-builder /app/dist ./web/dist
COPY configs/ ./configs/

EXPOSE 8080
CMD ["./server"]
```

**Tasks**:
- [ ] Set up project structure (1 day)
- [ ] Implement authentication (JWT) (2 days)
- [ ] Implement domain logic (users, products, orders) (2 days)
- [ ] Build React frontend (3 days)
- [ ] Add WebSocket support (1 day)
- [ ] Create Docker deployment (1 day)
- [ ] Write deployment guide (1 day)
- [ ] Add comprehensive README (1 day)

**Timeline**: 8 working days

---

### 6. Integration Testing 🔜

**Scope**:
- Integration tests for Gin adapter
- Middleware integration tests
- End-to-end API tests
- Frontend integration tests (Playwright/Cypress)

**Test Coverage**:
- [ ] Gin server lifecycle (startup, shutdown)
- [ ] All middleware (logging, recovery, tracing, metrics, CORS)
- [ ] Request validation scenarios
- [ ] Health check endpoints
- [ ] Authentication flow
- [ ] CRUD operations
- [ ] WebSocket connections
- [ ] Frontend E2E tests

**Tasks**:
- [ ] Write Gin adapter integration tests (1 day)
- [ ] Write middleware integration tests (1 day)
- [ ] Write E2E API tests (1 day)
- [ ] Write frontend E2E tests (1 day)

**Timeline**: 2 working days

---

## Implementation Timeline

### Week 1-2: Gin Adapter & Middleware (2 weeks)
- Days 1-3: Core Gin adapter implementation
- Days 4-5: hyperion.Context integration
- Days 6-10: Middleware suite implementation
- Day 11: Testing and refinement

### Week 3: Validation & Health Checks (1 week)
- Days 1-3: Validation and binding
- Days 4-6: Health checks and graceful shutdown
- Day 7: Integration testing

### Week 4-5: Full-Stack Example (1.5 weeks)
- Days 1-4: Backend implementation
- Days 5-7: Frontend implementation
- Days 8-9: Docker deployment and documentation
- Day 10: Final testing and bug fixes

**Total**: 5 weeks

---

## Technical Challenges

### Challenge 1: Context Propagation
**Problem**: Seamlessly convert gin.Context to hyperion.Context
**Solution**: Wrapper pattern with context.Context inheritance
**Status**: Planned

### Challenge 2: Error Handling
**Problem**: Unified error response format across all handlers
**Solution**: Custom error middleware with structured error types
**Status**: Planned

### Challenge 3: WebSocket Integration
**Problem**: WebSocket doesn't follow HTTP request/response pattern
**Solution**: Separate WebSocket handler with hyperion.Context support
**Status**: Planned

### Challenge 4: Frontend-Backend Integration
**Problem**: CORS, authentication, API client setup
**Solution**: Pre-configured CORS middleware, JWT auth, React Query integration
**Status**: Planned

---

## Success Metrics

### Code Metrics
- Test coverage: >= 80% for all components
- Performance: 10,000+ req/s, <10ms p99 latency
- Lines of code: ~1,500 LOC (gin), ~800 LOC (middleware), ~2,000 LOC (example)

### Quality Metrics
- golangci-lint: Zero errors
- Integration tests: All passing
- Example app: Fully functional full-stack application

### Performance Benchmarks
- Request throughput: >10,000 req/s (hello world)
- P99 latency: <10ms (simple endpoints)
- Memory usage: <100MB (idle), <500MB (under load)

### Community Metrics (Target)
- Production users: 50+ by end of v2.3
- GitHub stars: 1,000+
- Community adapters: 5+
- Tutorial completions: 100+

---

## Next Epic

👉 **[Epic 5: Microservices](epic-5-microservices.md)** (v2.4 - Planned)

**Focus**: gRPC adapter, HTTP client, service discovery, microservices example

**Timeline**: June 2026

---

## Related Documentation

- [Epic 1: Core Foundation](epic-1-core-foundation.md)
- [Epic 2: Essential Adapters](epic-2-essential-adapters.md)
- [Epic 3: Observability Stack](epic-3-observability-stack.md)
- [Architecture Overview](../architecture.md)
- [Implementation Plan](../implementation-plan.md)

---

**Epic Status**: 🔜 **PLANNED** (April 2026)

**Last Updated**: October 2025
**Version**: 2.3 Planning
