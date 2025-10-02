# Epic 5: Microservices (v2.4)

**Version**: 2.4
**Status**: 🔜 **PLANNED** (June 2026)
**Duration**: 6 weeks
**Priority**: ⭐⭐⭐⭐

---

## Overview

Implement production-ready **microservices adapters** including gRPC server/client, HTTP client, service discovery, and API gateway, enabling scalable distributed systems while maintaining zero lock-in and observability.

---

## Goals

### Primary Goals
1. gRPC server adapter with full hyperion.Context integration
2. gRPC client with automatic tracing and load balancing
3. Production-ready HTTP client with retry, timeout, circuit breaker
4. Service discovery integration (Consul, Etcd, Kubernetes)
5. API Gateway pattern implementation
6. Complete microservices example application

### Success Criteria
- [ ] gRPC adapter supports unary, streaming, and bidirectional RPCs
- [ ] Distributed tracing spans across all service calls
- [ ] HTTP client handles failures gracefully (retry, circuit breaker)
- [ ] Service discovery auto-registers and discovers services
- [ ] API Gateway routes requests with proper load balancing
- [ ] Example app demonstrates production microservices patterns
- [ ] Performance: gRPC throughput >50,000 req/s, <5ms p99 latency

---

## Deliverables

### 1. gRPC Server Adapter 🔜

**Package**: `adapter/grpc/`

**Scope**:
- gRPC server with hyperion.Context integration
- Unary, server streaming, client streaming, bidirectional streaming
- Automatic interceptors (logging, recovery, tracing, metrics)
- Health check protocol (grpc.health.v1)
- Reflection API support
- TLS/mTLS support

**Core Implementation**:
```go
type GrpcServer struct {
    server *grpc.Server
    config ServerConfig
}

// Unary interceptor - injects hyperion.Context
func UnaryInterceptor(
    logger hyperion.Logger,
    tracer hyperion.Tracer,
    db hyperion.Database,
) grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        // Create hyperion.Context
        hctx := hyperion.NewContext(ctx, logger, tracer, db)

        // Start trace span
        hctx, span := tracer.Start(hctx, info.FullMethod)
        defer span.End()

        // Log request
        logger.Info("grpc request", "method", info.FullMethod)

        // Execute handler
        resp, err := handler(hctx, req)

        // Record error in span
        if err != nil {
            span.RecordError(err)
            logger.Error("grpc request failed", "method", info.FullMethod, "error", err)
        }

        return resp, err
    }
}

// Stream interceptor
func StreamInterceptor(
    logger hyperion.Logger,
    tracer hyperion.Tracer,
    db hyperion.Database,
) grpc.StreamServerInterceptor {
    return func(
        srv interface{},
        ss grpc.ServerStream,
        info *grpc.StreamServerInfo,
        handler grpc.StreamHandler,
    ) error {
        // Create hyperion.Context
        hctx := hyperion.NewContext(ss.Context(), logger, tracer, db)

        // Wrap server stream with hyperion.Context
        wrapped := &hyperionServerStream{
            ServerStream: ss,
            ctx:          hctx,
        }

        return handler(srv, wrapped)
    }
}

type hyperionServerStream struct {
    grpc.ServerStream
    ctx hyperion.Context
}

func (s *hyperionServerStream) Context() context.Context {
    return s.ctx
}
```

**Module Export**:
```go
var Module = fx.Module("hyperion.adapter.grpc",
    fx.Provide(
        NewGrpcServer,
        NewUnaryInterceptor,
        NewStreamInterceptor,
    ),
    fx.Invoke(RegisterHealthService),
    fx.Invoke(RegisterReflectionService),
    fx.Invoke(RegisterShutdownHook),
)
```

**Configuration Example**:
```yaml
grpc:
  host: 0.0.0.0
  port: 9090
  max_recv_msg_size: 4194304    # 4MB
  max_send_msg_size: 4194304    # 4MB
  connection_timeout: 120s
  enable_reflection: true        # Enable gRPC reflection
  enable_health_check: true      # Enable health check protocol

  tls:
    enabled: true
    cert_file: /etc/certs/server.crt
    key_file: /etc/certs/server.key
    client_ca_file: /etc/certs/ca.crt  # For mTLS
```

**Service Implementation Example**:
```go
// Generated protobuf service interface
type UserServiceServer interface {
    GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error)
    CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error)
    ListUsers(req *ListUsersRequest, stream UserService_ListUsersServer) error
}

// Implementation with hyperion.Context
type userServiceImpl struct {
    userService *service.UserService
}

func (s *userServiceImpl) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
    // ctx is actually hyperion.Context (injected by interceptor)
    hctx := ctx.(hyperion.Context)

    // Automatic tracing
    hctx, span := hctx.Tracer().Start(hctx, "UserService.GetUser")
    defer span.End()

    // Business logic
    user, err := s.userService.GetUser(hctx, req.UserId)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
    }

    return &GetUserResponse{
        User: toProtoUser(user),
    }, nil
}

func (s *userServiceImpl) ListUsers(req *ListUsersRequest, stream UserService_ListUsersServer) error {
    // Stream context is hyperion.Context
    hctx := stream.Context().(hyperion.Context)

    users, err := s.userService.ListUsers(hctx, int(req.PageSize), int(req.Page))
    if err != nil {
        return status.Errorf(codes.Internal, "failed to list users: %v", err)
    }

    for _, user := range users {
        if err := stream.Send(&UserResponse{User: toProtoUser(user)}); err != nil {
            return err
        }
    }

    return nil
}
```

**Tasks**:
- [ ] Implement gRPC server core (2 days)
- [ ] Implement unary interceptor (1 day)
- [ ] Implement stream interceptor (1 day)
- [ ] Add health check service (1 day)
- [ ] Add reflection service (1 day)
- [ ] Implement TLS/mTLS support (1 day)
- [ ] Write unit tests (>80% coverage) (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 6 working days

---

### 2. gRPC Client Adapter 🔜

**Package**: `adapter/grpc/client/`

**Scope**:
- gRPC client with automatic tracing propagation
- Client interceptors (logging, retry, timeout)
- Load balancing (round-robin, least-conn)
- Service discovery integration
- Connection pooling and health checking

**Core Implementation**:
```go
type GrpcClient struct {
    conn *grpc.ClientConn
}

// Unary client interceptor
func UnaryClientInterceptor(tracer hyperion.Tracer) grpc.UnaryClientInterceptor {
    return func(
        ctx context.Context,
        method string,
        req, reply interface{},
        cc *grpc.ClientConn,
        invoker grpc.UnaryInvoker,
        opts ...grpc.CallOption,
    ) error {
        // Start trace span
        ctx, span := tracer.Start(ctx, fmt.Sprintf("grpc.client:%s", method))
        defer span.End()

        // Inject trace context into metadata
        ctx = otel.GetTextMapPropagator().Inject(ctx, &metadataCarrier{metadata: metadata.MD{}})

        // Execute RPC
        err := invoker(ctx, method, req, reply, cc, opts...)

        // Record error
        if err != nil {
            span.RecordError(err)
        }

        return err
    }
}

// Stream client interceptor
func StreamClientInterceptor(tracer hyperion.Tracer) grpc.StreamClientInterceptor {
    return func(
        ctx context.Context,
        desc *grpc.StreamDesc,
        cc *grpc.ClientConn,
        method string,
        streamer grpc.Streamer,
        opts ...grpc.CallOption,
    ) (grpc.ClientStream, error) {
        ctx, span := tracer.Start(ctx, fmt.Sprintf("grpc.client:%s", method))

        // Inject trace context
        ctx = otel.GetTextMapPropagator().Inject(ctx, &metadataCarrier{metadata: metadata.MD{}})

        stream, err := streamer(ctx, desc, cc, method, opts...)
        if err != nil {
            span.End()
            return nil, err
        }

        return &tracedClientStream{
            ClientStream: stream,
            span:         span,
        }, nil
    }
}

type tracedClientStream struct {
    grpc.ClientStream
    span trace.Span
}

func (s *tracedClientStream) RecvMsg(m interface{}) error {
    err := s.ClientStream.RecvMsg(m)
    if err != nil {
        s.span.RecordError(err)
        s.span.End()
    }
    return err
}
```

**Client Factory**:
```go
type ClientFactory struct {
    tracer hyperion.Tracer
    logger hyperion.Logger
}

func (f *ClientFactory) NewClient(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
    // Default options
    defaultOpts := []grpc.DialOption{
        grpc.WithUnaryInterceptor(UnaryClientInterceptor(f.tracer)),
        grpc.WithStreamInterceptor(StreamClientInterceptor(f.tracer)),
        grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
    }

    opts = append(defaultOpts, opts...)

    conn, err := grpc.DialContext(ctx, target, opts...)
    if err != nil {
        f.logger.Error("failed to create grpc client", "target", target, "error", err)
        return nil, err
    }

    f.logger.Info("grpc client created", "target", target)
    return conn, nil
}
```

**Usage Example**:
```go
func (s *OrderService) CreateOrder(ctx hyperion.Context, order *domain.Order) error {
    // Get user info from User Service via gRPC
    userClient := userpb.NewUserServiceClient(s.userGrpcClient)

    userResp, err := userClient.GetUser(ctx, &userpb.GetUserRequest{
        UserId: order.UserID,
    })
    if err != nil {
        return fmt.Errorf("failed to get user: %w", err)
    }

    // Distributed tracing automatically propagated
    order.UserEmail = userResp.User.Email

    // Continue with order creation...
    return s.orderRepo.Create(ctx, order)
}
```

**Tasks**:
- [ ] Implement client factory (1 day)
- [ ] Implement unary client interceptor (1 day)
- [ ] Implement stream client interceptor (1 day)
- [ ] Add retry logic (1 day)
- [ ] Add load balancing support (1 day)
- [ ] Write unit tests (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 4 working days

---

### 3. HTTP Client Adapter 🔜

**Package**: `adapter/httpclient/`

**Scope**:
- Production-ready HTTP client with observability
- Retry with exponential backoff
- Circuit breaker pattern
- Timeout configuration
- Request/response logging
- Automatic tracing propagation

**Core Implementation**:
```go
type HttpClient struct {
    client  *http.Client
    tracer  hyperion.Tracer
    logger  hyperion.Logger
    config  ClientConfig
}

type ClientConfig struct {
    Timeout         time.Duration
    MaxRetries      int
    RetryWaitMin    time.Duration
    RetryWaitMax    time.Duration
    CircuitBreaker  CircuitBreakerConfig
}

type CircuitBreakerConfig struct {
    Enabled          bool
    MaxRequests      uint32
    Interval         time.Duration
    Timeout          time.Duration
    FailureThreshold uint32
}

func (c *HttpClient) Do(ctx hyperion.Context, req *http.Request) (*http.Response, error) {
    // Start trace span
    ctx, span := c.tracer.Start(ctx, fmt.Sprintf("http.client:%s %s", req.Method, req.URL.Path))
    defer span.End()

    // Inject trace context into headers
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

    // Set span attributes
    span.SetAttributes(
        attribute.String("http.method", req.Method),
        attribute.String("http.url", req.URL.String()),
    )

    // Execute with retry
    resp, err := c.doWithRetry(ctx, req)
    if err != nil {
        span.RecordError(err)
        c.logger.Error("http request failed",
            "method", req.Method,
            "url", req.URL.String(),
            "error", err,
        )
        return nil, err
    }

    span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

    return resp, nil
}

func (c *HttpClient) doWithRetry(ctx hyperion.Context, req *http.Request) (*http.Response, error) {
    var resp *http.Response
    var err error

    for i := 0; i <= c.config.MaxRetries; i++ {
        if i > 0 {
            // Exponential backoff
            wait := time.Duration(math.Pow(2, float64(i-1))) * c.config.RetryWaitMin
            if wait > c.config.RetryWaitMax {
                wait = c.config.RetryWaitMax
            }

            c.logger.Debug("retrying http request",
                "attempt", i,
                "wait", wait,
                "url", req.URL.String(),
            )

            time.Sleep(wait)
        }

        resp, err = c.client.Do(req.WithContext(ctx))

        // Success or non-retryable error
        if err == nil && !c.shouldRetry(resp.StatusCode) {
            return resp, nil
        }

        // Retry on error or retryable status code
        if err != nil || c.shouldRetry(resp.StatusCode) {
            if resp != nil {
                resp.Body.Close()
            }
            continue
        }
    }

    return resp, err
}

func (c *HttpClient) shouldRetry(statusCode int) bool {
    return statusCode >= 500 || statusCode == 429
}
```

**Circuit Breaker Integration**:
```go
type circuitBreakerClient struct {
    *HttpClient
    cb *gobreaker.CircuitBreaker
}

func (c *circuitBreakerClient) Do(ctx hyperion.Context, req *http.Request) (*http.Response, error) {
    result, err := c.cb.Execute(func() (interface{}, error) {
        return c.HttpClient.Do(ctx, req)
    })

    if err != nil {
        return nil, err
    }

    return result.(*http.Response), nil
}
```

**JSON Client Helper**:
```go
type JsonClient struct {
    *HttpClient
}

func (c *JsonClient) Get(ctx hyperion.Context, url string, result interface{}) error {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return err
    }

    req.Header.Set("Accept", "application/json")

    resp, err := c.Do(ctx, req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("http error: %d", resp.StatusCode)
    }

    return json.NewDecoder(resp.Body).Decode(result)
}

func (c *JsonClient) Post(ctx hyperion.Context, url string, body, result interface{}) error {
    data, err := json.Marshal(body)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", url, bytes.NewReader(data))
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    resp, err := c.Do(ctx, req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("http error: %d", resp.StatusCode)
    }

    if result != nil {
        return json.NewDecoder(resp.Body).Decode(result)
    }

    return nil
}
```

**Configuration Example**:
```yaml
http_client:
  timeout: 30s
  max_retries: 3
  retry_wait_min: 100ms
  retry_wait_max: 5s

  circuit_breaker:
    enabled: true
    max_requests: 5
    interval: 60s
    timeout: 30s
    failure_threshold: 3
```

**Tasks**:
- [ ] Implement HTTP client core (1 day)
- [ ] Add retry with exponential backoff (1 day)
- [ ] Integrate circuit breaker (1 day)
- [ ] Add JSON client helpers (1 day)
- [ ] Add tracing propagation (1 day)
- [ ] Write unit tests (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 4 working days

---

### 4. Service Discovery Integration 🔜

**Package**: `adapter/discovery/`

**Scope**:
- Service registration and discovery
- Support for Consul, Etcd, Kubernetes
- Health check integration
- Automatic service de-registration
- Load balancing with service discovery

**Consul Adapter**:
```go
type ConsulDiscovery struct {
    client *consul.Client
    config DiscoveryConfig
}

type DiscoveryConfig struct {
    ServiceName string
    ServiceID   string
    Host        string
    Port        int
    Tags        []string
    HealthCheck HealthCheckConfig
}

type HealthCheckConfig struct {
    HTTP     string
    Interval string
    Timeout  string
}

func (d *ConsulDiscovery) Register(ctx context.Context) error {
    registration := &consul.AgentServiceRegistration{
        ID:      d.config.ServiceID,
        Name:    d.config.ServiceName,
        Address: d.config.Host,
        Port:    d.config.Port,
        Tags:    d.config.Tags,
        Check: &consul.AgentServiceCheck{
            HTTP:     d.config.HealthCheck.HTTP,
            Interval: d.config.HealthCheck.Interval,
            Timeout:  d.config.HealthCheck.Timeout,
        },
    }

    return d.client.Agent().ServiceRegister(registration)
}

func (d *ConsulDiscovery) Deregister(ctx context.Context) error {
    return d.client.Agent().ServiceDeregister(d.config.ServiceID)
}

func (d *ConsulDiscovery) Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
    services, _, err := d.client.Health().Service(serviceName, "", true, nil)
    if err != nil {
        return nil, err
    }

    instances := make([]*ServiceInstance, 0, len(services))
    for _, service := range services {
        instances = append(instances, &ServiceInstance{
            ID:      service.Service.ID,
            Address: service.Service.Address,
            Port:    service.Service.Port,
            Tags:    service.Service.Tags,
        })
    }

    return instances, nil
}
```

**Kubernetes Adapter**:
```go
type K8sDiscovery struct {
    clientset *kubernetes.Clientset
    namespace string
}

func (d *K8sDiscovery) Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
    endpoints, err := d.clientset.CoreV1().Endpoints(d.namespace).Get(ctx, serviceName, metav1.GetOptions{})
    if err != nil {
        return nil, err
    }

    var instances []*ServiceInstance
    for _, subset := range endpoints.Subsets {
        for _, addr := range subset.Addresses {
            for _, port := range subset.Ports {
                instances = append(instances, &ServiceInstance{
                    Address: addr.IP,
                    Port:    int(port.Port),
                })
            }
        }
    }

    return instances, nil
}
```

**gRPC Resolver Integration**:
```go
// Custom gRPC resolver using service discovery
type discoveryResolver struct {
    discovery Discovery
    service   string
}

func (r *discoveryResolver) ResolveNow(resolver.ResolveNowOptions) {
    instances, err := r.discovery.Discover(context.Background(), r.service)
    if err != nil {
        return
    }

    addrs := make([]resolver.Address, 0, len(instances))
    for _, inst := range instances {
        addrs = append(addrs, resolver.Address{
            Addr: fmt.Sprintf("%s:%d", inst.Address, inst.Port),
        })
    }

    r.cc.UpdateState(resolver.State{Addresses: addrs})
}

// Usage
conn, err := grpc.Dial(
    "discovery:///user-service",  // Use custom scheme
    grpc.WithResolvers(newDiscoveryResolver(consulDiscovery, "user-service")),
)
```

**Configuration Example**:
```yaml
service_discovery:
  type: consul              # consul, etcd, kubernetes

  consul:
    address: localhost:8500
    service:
      name: user-service
      id: user-service-1
      host: 192.168.1.10
      port: 9090
      tags:
        - v1
        - production
      health_check:
        http: http://192.168.1.10:8080/health/ready
        interval: 10s
        timeout: 5s

  kubernetes:
    namespace: default
    in_cluster: true
```

**Tasks**:
- [ ] Define Discovery interface (1 day)
- [ ] Implement Consul adapter (2 days)
- [ ] Implement Kubernetes adapter (2 days)
- [ ] Add gRPC resolver integration (1 day)
- [ ] Add lifecycle hooks (register/deregister) (1 day)
- [ ] Write integration tests (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 6 working days

---

### 5. API Gateway 🔜

**Package**: `examples/api-gateway/`

**Scope**:
- Central entry point for microservices
- Request routing and load balancing
- Authentication and authorization
- Rate limiting
- Response aggregation
- Protocol translation (HTTP → gRPC)

**Architecture**:
```
Client (HTTP)
    ↓
API Gateway (Port 8080)
    ├── Authentication (JWT)
    ├── Rate Limiting
    ├── Request Routing
    └── Protocol Translation
        ├─→ User Service (gRPC :9091)
        ├─→ Order Service (gRPC :9092)
        ├─→ Product Service (gRPC :9093)
        └─→ Notification Service (gRPC :9094)
```

**Gateway Implementation**:
```go
type Gateway struct {
    userClient    userpb.UserServiceClient
    orderClient   orderpb.OrderServiceClient
    productClient productpb.ProductServiceClient
}

// HTTP endpoint that calls multiple gRPC services
func (g *Gateway) GetOrderDetails(ctx hyperion.Context) error {
    orderID := ctx.Param("id")

    // Get order from Order Service
    orderResp, err := g.orderClient.GetOrder(ctx, &orderpb.GetOrderRequest{
        OrderId: orderID,
    })
    if err != nil {
        return err
    }

    // Get user from User Service (parallel)
    userResp, err := g.userClient.GetUser(ctx, &userpb.GetUserRequest{
        UserId: orderResp.Order.UserId,
    })
    if err != nil {
        return err
    }

    // Get products from Product Service (parallel)
    var products []*productpb.Product
    for _, item := range orderResp.Order.Items {
        productResp, err := g.productClient.GetProduct(ctx, &productpb.GetProductRequest{
            ProductId: item.ProductId,
        })
        if err != nil {
            continue
        }
        products = append(products, productResp.Product)
    }

    // Aggregate response
    ctx.JSON(200, OrderDetailsResponse{
        Order:    toRestOrder(orderResp.Order),
        User:     toRestUser(userResp.User),
        Products: toRestProducts(products),
    })

    return nil
}
```

**Service Routing**:
```go
func SetupGatewayRoutes(server *gin.GinServer, gw *Gateway) {
    // Global middleware
    server.Use(
        middleware.Recovery(),
        middleware.RequestLogger(),
        middleware.Tracing("api-gateway"),
        middleware.Metrics(),
        middleware.RateLimit(rate.NewLimiter(1000, 2000)),
    )

    // Public routes
    server.POST("/api/auth/login", gw.Login)

    // Protected routes
    api := server.Group("/api", middleware.AuthRequired())
    {
        // User routes → User Service
        api.GET("/users/:id", gw.GetUser)
        api.PUT("/users/:id", gw.UpdateUser)

        // Order routes → Order Service + User Service
        api.GET("/orders/:id", gw.GetOrderDetails)  // Aggregates multiple services
        api.POST("/orders", gw.CreateOrder)

        // Product routes → Product Service
        api.GET("/products", gw.ListProducts)
        api.GET("/products/:id", gw.GetProduct)
    }
}
```

**Tasks**:
- [ ] Implement gateway core (2 days)
- [ ] Add request routing (1 day)
- [ ] Add response aggregation (1 day)
- [ ] Add authentication/authorization (1 day)
- [ ] Add rate limiting (1 day)
- [ ] Documentation and examples (1 day)

**Timeline**: 4 working days

---

### 6. Microservices Example Application 🔜

**Package**: `examples/microservices-demo/`

**Scope**:
- Complete microservices architecture
- 4-5 independent services
- API Gateway
- Service discovery
- Distributed tracing
- Docker Compose deployment
- Kubernetes manifests

**Services Architecture**:
```
API Gateway (HTTP :8080)
    ├── User Service (gRPC :9091, HTTP :8081)
    │   └── PostgreSQL
    ├── Order Service (gRPC :9092, HTTP :8082)
    │   └── PostgreSQL
    ├── Product Service (gRPC :9093, HTTP :8083)
    │   └── PostgreSQL
    └── Notification Service (gRPC :9094, HTTP :8084)
        └── Redis (queue)

Infrastructure:
    ├── Consul (Service Discovery :8500)
    ├── Jaeger (Tracing :16686)
    ├── Prometheus (Metrics :9090)
    └── Grafana (Dashboard :3000)
```

**Service Communication Flow**:
```
1. Client → API Gateway (HTTP)
2. API Gateway → User Service (gRPC)
3. API Gateway → Order Service (gRPC)
4. Order Service → Product Service (gRPC)
5. Order Service → Notification Service (gRPC, async)
6. All services register with Consul
7. All operations traced in Jaeger
8. All metrics in Prometheus/Grafana
```

**Docker Compose**:
```yaml
version: '3.8'

services:
  # Infrastructure
  consul:
    image: consul:latest
    ports:
      - "8500:8500"

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"
      - "14268:14268"

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"

  # Databases
  postgres-user:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: userdb
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"

  postgres-order:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: orderdb
      POSTGRES_PASSWORD: postgres
    ports:
      - "5433:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  # Services
  api-gateway:
    build: ./api-gateway
    ports:
      - "8080:8080"
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ENDPOINT=http://jaeger:14268/api/traces
    depends_on:
      - consul
      - jaeger

  user-service:
    build: ./user-service
    ports:
      - "9091:9091"
      - "8081:8081"
    environment:
      - CONSUL_ADDR=consul:8500
      - DATABASE_DSN=postgres://postgres:postgres@postgres-user:5432/userdb
    depends_on:
      - consul
      - postgres-user

  order-service:
    build: ./order-service
    ports:
      - "9092:9092"
      - "8082:8082"
    environment:
      - CONSUL_ADDR=consul:8500
      - DATABASE_DSN=postgres://postgres:postgres@postgres-order:5432/orderdb
    depends_on:
      - consul
      - postgres-order

  product-service:
    build: ./product-service
    ports:
      - "9093:9093"
      - "8083:8083"
    environment:
      - CONSUL_ADDR=consul:8500
    depends_on:
      - consul

  notification-service:
    build: ./notification-service
    ports:
      - "9094:9094"
      - "8084:8084"
    environment:
      - CONSUL_ADDR=consul:8500
      - REDIS_ADDR=redis:6379
    depends_on:
      - consul
      - redis
```

**Kubernetes Deployment**:
```yaml
# user-service-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: user-service
        image: hyperion/user-service:v2.4
        ports:
        - containerPort: 9091
          name: grpc
        - containerPort: 8081
          name: http
        env:
        - name: DATABASE_DSN
          valueFrom:
            secretKeyRef:
              name: user-db-secret
              key: dsn
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8081
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8081

---
apiVersion: v1
kind: Service
metadata:
  name: user-service
spec:
  selector:
    app: user-service
  ports:
  - name: grpc
    port: 9091
    targetPort: 9091
  - name: http
    port: 8081
    targetPort: 8081
```

**Tasks**:
- [ ] Design service architecture (1 day)
- [ ] Implement User Service (2 days)
- [ ] Implement Order Service (2 days)
- [ ] Implement Product Service (1 day)
- [ ] Implement Notification Service (1 day)
- [ ] Implement API Gateway (2 days)
- [ ] Create Docker Compose setup (1 day)
- [ ] Create Kubernetes manifests (2 days)
- [ ] Write deployment guide (1 day)
- [ ] Add comprehensive README (1 day)

**Timeline**: 10 working days

---

### 7. Integration Testing 🔜

**Scope**:
- Integration tests for gRPC server/client
- Integration tests for HTTP client
- Service discovery integration tests
- End-to-end microservices tests

**Test Coverage**:
- [ ] gRPC unary and streaming RPCs
- [ ] gRPC client retry and load balancing
- [ ] HTTP client retry and circuit breaker
- [ ] Service registration and discovery (Consul)
- [ ] Distributed tracing across services
- [ ] API Gateway routing and aggregation
- [ ] Microservices E2E scenarios

**Tasks**:
- [ ] Write gRPC integration tests (2 days)
- [ ] Write HTTP client integration tests (1 day)
- [ ] Write service discovery tests (1 day)
- [ ] Write E2E microservices tests (2 days)

**Timeline**: 4 working days

---

## Implementation Timeline

### Week 1-2: gRPC Adapter (2 weeks)
- Days 1-3: gRPC server implementation
- Days 4-5: gRPC client implementation
- Days 6-7: Interceptors and middleware
- Days 8-10: Testing and refinement

### Week 3: HTTP Client & Discovery (1 week)
- Days 1-2: HTTP client implementation
- Days 3-4: Retry and circuit breaker
- Days 5-7: Service discovery adapters

### Week 4: API Gateway (1 week)
- Days 1-2: Gateway core implementation
- Days 3-4: Routing and aggregation
- Days 5-7: Authentication and rate limiting

### Week 5-6: Microservices Example (2 weeks)
- Days 1-5: Individual services implementation
- Days 6-7: API Gateway integration
- Days 8-10: Docker Compose and Kubernetes
- Days 11-12: Documentation and testing

**Total**: 6 weeks

---

## Technical Challenges

### Challenge 1: Distributed Transaction Management
**Problem**: Ensure consistency across multiple services
**Solution**: Saga pattern with compensation logic
**Status**: Planned

### Challenge 2: Service Mesh Integration
**Problem**: Should Hyperion provide service mesh features?
**Solution**: Focus on core features, integrate with Istio/Linkerd for advanced scenarios
**Status**: Deferred to v3.0

### Challenge 3: gRPC Load Balancing
**Problem**: Client-side load balancing with service discovery
**Solution**: Custom gRPC resolver with health checking
**Status**: Planned

### Challenge 4: Protocol Translation Overhead
**Problem**: API Gateway HTTP → gRPC conversion overhead
**Solution**: Connection pooling, protobuf optimization
**Status**: Planned

---

## Success Metrics

### Code Metrics
- Test coverage: >= 80% for all components
- Performance: gRPC >50,000 req/s, <5ms p99 latency
- Lines of code: ~2,000 LOC (grpc), ~1,000 LOC (httpclient), ~1,500 LOC (discovery), ~3,000 LOC (example)

### Quality Metrics
- golangci-lint: Zero errors
- Integration tests: All passing
- Example app: Fully functional microservices

### Performance Benchmarks
- gRPC throughput: >50,000 req/s (unary)
- gRPC latency: <5ms p99
- HTTP client: Successful retry on transient failures
- API Gateway: <10ms overhead

### Community Metrics (Target)
- Production users: 100+ by end of v2.4
- GitHub stars: 2,000+
- Community adapters: 10+
- Enterprise adoption: 5+ companies

---

## Future Work (v3.0+)

### Advanced Features (Deferred)
- Service mesh integration (Istio, Linkerd)
- Advanced resilience patterns (bulkhead, rate limiting per service)
- gRPC streaming optimizations
- GraphQL gateway
- Event-driven architecture (message queue integration)
- Distributed saga orchestration

---

## Related Documentation

- [Epic 1: Core Foundation](epic-1-core-foundation.md)
- [Epic 2: Essential Adapters](epic-2-essential-adapters.md)
- [Epic 3: Observability Stack](epic-3-observability-stack.md)
- [Epic 4: Web Framework](epic-4-web-framework.md)
- [Architecture Overview](../architecture.md)
- [Implementation Plan](../implementation-plan.md)

---

**Epic Status**: 🔜 **PLANNED** (June 2026)

**Last Updated**: October 2025
**Version**: 2.4 Planning
