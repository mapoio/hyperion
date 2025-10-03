# Hyperion Deep Call Chain Example

This example demonstrates **distributed tracing across 10 levels of nested service calls** using Hyperion's elegant observability features.

## 🎯 What This Example Demonstrates

- **10-Level Deep Call Chain**: Realistic e-commerce order flow with 16 interconnected services
- **Automatic Trace Propagation**: Single `trace_id` flows through all 10 levels
- **Elegant API**: Using `hyperion.StartSpan()` for zero-boilerplate span creation
- **Context-Aware Logging**: Every log automatically includes trace context
- **Dependency Injection**: Uber FX manages all 16 services automatically
- **Production Patterns**: Real-world service architecture with proper error handling

## 🏗️ Architecture

### Service Call Chain (10 Levels Deep)

```
Level 1: OrderService (Entry Point)
    │
    ├──▶ Level 2: PaymentService
    │       │
    │       └──▶ Level 3: FraudDetectionService
    │               │
    │               └──▶ Level 4: RiskAnalysisService
    │                       │
    │                       └──▶ Level 5: UserProfileService
    │                               │
    │                               └──▶ Level 6: CacheService
    │                                       │
    │                                       └──▶ Level 7: StorageService
    │                                               │
    │                                               └──▶ Level 8: ReplicationService
    │                                                       │
    │                                                       └──▶ Level 9: HealthCheckService
    │                                                               │
    │                                                               └──▶ Level 10: MonitoringService ⭐ DEEPEST
    │
    ├──▶ Level 2: InventoryService
    │       │
    │       └──▶ Level 3: WarehouseService
    │               │
    │               └──▶ Level 4: ShippingService
    │                       │
    │                       └──▶ Level 5: CarrierService
    │                               │
    │                               └──▶ Level 6: RouteService
    │                                       │
    │                                       └──▶ Level 7: GeoService
    │                                               │
    │                                               └──▶ Level 8: CoordinateService
    │                                                       │
    │                                                       └──▶ Level 9: MappingService
    │                                                               │
    │                                                               └──▶ Level 10: MonitoringService ⭐ DEEPEST
    │
    └──▶ Level 3: NotificationService
            │
            ├──▶ Level 4: EmailService
            │       │
            │       └──▶ Level 5: TemplateService
            │               │
            │               └──▶ Level 7: StorageService
            │                       │
            │                       └──▶ Level 8: ReplicationService
            │                               │
            │                               └──▶ Level 9: HealthCheckService
            │                                       │
            │                                       └──▶ Level 10: MonitoringService ⭐ DEEPEST
            │
            └──▶ Level 4: SMSService
                    │
                    └──▶ Level 5: TemplateService
                            │
                            └──▶ Level 7: StorageService
                                    │
                                    └──▶ Level 8: ReplicationService
                                            │
                                            └──▶ Level 9: HealthCheckService
                                                    │
                                                    └──▶ Level 10: MonitoringService ⭐ DEEPEST
```

### E-commerce Order Flow

When you create an order, this is what happens:

1. **Level 1: OrderService** - Orchestrates the entire order creation
2. **Level 2: InventoryService** - Checks and reserves stock
3. **Level 3: WarehouseService** - Queries warehouse stock levels
4. **Level 4: ShippingService** - Verifies shipping availability
5. **Level 5: CarrierService** - Checks carrier capacity
6. **Level 6: RouteService** - Finds optimal delivery route
7. **Level 7: GeoService** - Retrieves geolocation data
8. **Level 8: CoordinateService** - Gets GPS coordinates
9. **Level 9: MappingService** - Accesses mapping database
10. **Level 10: MonitoringService** - Collects system health metrics

All happening in **one HTTP request** with **one shared trace_id**!

## 📊 Complete Service Catalog (16 Services)

| Level | Service | Responsibility | Dependencies |
|-------|---------|----------------|--------------|
| 1 | **OrderService** | Order orchestration | InventoryService, PaymentService |
| 2 | **PaymentService** | Payment processing | FraudDetectionService, NotificationService |
| 2 | **InventoryService** | Stock management | WarehouseService |
| 3 | **FraudDetectionService** | Fraud analysis | RiskAnalysisService |
| 3 | **NotificationService** | Notification coordination | EmailService, SMSService |
| 3 | **WarehouseService** | Warehouse operations | ShippingService |
| 4 | **RiskAnalysisService** | Risk scoring | UserProfileService |
| 4 | **EmailService** | Email delivery | TemplateService |
| 4 | **SMSService** | SMS delivery | TemplateService |
| 4 | **ShippingService** | Shipping logistics | CarrierService |
| 5 | **UserProfileService** | User data | CacheService |
| 5 | **TemplateService** | Template rendering | StorageService |
| 5 | **CarrierService** | Carrier management | RouteService |
| 6 | **CacheService** | Distributed caching | StorageService |
| 6 | **RouteService** | Route optimization | GeoService |
| 7 | **GeoService** | Geolocation | CoordinateService |
| 7 | **StorageService** | Object storage | ReplicationService |
| 8 | **ReplicationService** | Data replication | HealthCheckService |
| 8 | **CoordinateService** | GPS coordinates | MappingService |
| 9 | **HealthCheckService** | Health monitoring | MonitoringService |
| 9 | **MappingService** | Map data | MonitoringService |
| 10 | **MonitoringService** | System metrics | *(Deepest service)* |

## 🚀 Quick Start

### 1. Start HyperDX (Optional - for viewing traces)

```bash
cd /Users/mapo/code/hyperion/example/otel
make hyperdx-up
```

**HyperDX UI**: http://localhost:8080

### 2. Run the Application

```bash
cd /Users/mapo/code/hyperion/example/otel
make run
```

Server starts on: http://localhost:8090

### 3. Create an Order (Triggers 10-Level Chain)

```bash
curl -s -X POST http://localhost:8090/api/orders \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": "user-123",
    "product_id": "product-456",
    "amount": 99.99
  }' | jq .
```

**Expected Response**:
```json
{
  "order_id": "ORD-1759512986488613000",
  "user_id": "user-123",
  "product_id": "product-456",
  "amount": 99.99,
  "status": "created"
}
```

### 4. View the Trace Waterfall

Check your application logs to see the complete 10-level call chain:

```
Level 1: OrderService.CreateOrder         (span_id: 81b24e96b4b39940)
Level 2: InventoryService.CheckStock      (span_id: 8a99e99a26c8f18b)
Level 3: WarehouseService.GetStockLevel   (span_id: dfa1f304d499f43d)
Level 4: ShippingService.CheckAvailability(span_id: d307753a4ba1ec3d)
Level 5: CarrierService.CheckCapacity     (span_id: 16a29bd2f20a0af6)
Level 6: RouteService.FindOptimalRoute    (span_id: 7a84bf03fbeb50ad)
Level 7: GeoService.GetLocationData       (span_id: 0e735afab96d94bb)
Level 8: CoordinateService.GetCoordinates (span_id: b7120adad0e82196)
Level 9: MappingService.GetMappingData    (span_id: 3916e4d23deceef4)
Level 10: MonitoringService.CollectMetrics(span_id: 7e3238523306baf2) ⭐ DEEPEST
```

All logs share the **same trace_id**: `7ee23d5d632b5fc6983a98390c9b196d`

## 🎨 Code Patterns

### Using `hyperion.StartSpan()` (The Elegant Way)

Every service uses this pattern for zero-boilerplate observability:

```go
func (s *OrderService) CreateOrder(ctx context.Context, userID, productID string, amount float64) (string, error) {
    // Creates span, updates context, and logger in one call
    ctx, span, logger := hyperion.StartSpan(ctx, s.tracer, s.logger, "OrderService.CreateOrder",
        hyperion.WithAttributes(
            hyperion.String("user.id", userID),
            hyperion.String("product.id", productID),
            hyperion.Float64("amount", amount),
        ),
    )
    defer span.End()

    logger.Info("creating order", "user_id", userID)

    // Call next level services...
    available, err := s.inventoryService.CheckStock(ctx, productID)

    return orderID, nil
}
```

**What happens under the hood**:
1. Creates a new span with tracer
2. Updates context with new span
3. Updates logger to include new `span_id` and `trace_id`
4. Returns all three in one call

**Before vs After**:

```go
// ❌ Old way (7 lines of boilerplate)
ctx, span := tracer.Start(ctx, "operation")
defer span.End()
span.SetAttributes(...)
logger = logger.WithContext(ctx)
logger.Info("message")

// ✅ New way (1 line)
ctx, span, logger := hyperion.StartSpan(ctx, tracer, logger, "operation", opts...)
defer span.End()
logger.Info("message")  // Automatically includes trace context
```

### Dependency Injection with Uber FX

All 16 services are wired together automatically:

```go
// internal/services/module.go
var Module = fx.Module("services",
    // Level 10 (deepest)
    fx.Provide(NewMonitoringService),

    // Level 9
    fx.Provide(NewHealthCheckService),
    fx.Provide(NewMappingService),

    // ... all other levels ...

    // Level 1 (entry point)
    fx.Provide(NewOrderService),
)
```

In `main.go`:

```go
fx.New(
    // ... other modules ...
    services.Module,  // Provides all 16 services

    fx.Invoke(func(orderService *services.OrderService) {
        // OrderService ready with all dependencies injected!
    }),
).Run()
```

## 📈 Observability Features

### Automatic Trace Propagation

**Single Request**: `POST /api/orders`

**Results in**:
- **1 trace_id** shared across all services
- **10+ unique span_ids** (one per service call)
- **Complete parent-child relationships** visible in HyperDX
- **Full request waterfall** showing exactly where time is spent

### Structured Logging with Context

Every log entry includes:

```json
{
  "level": "info",
  "ts": "2025-10-04T12:34:56.789Z",
  "msg": "user profile retrieved",
  "trace_id": "7ee23d5d632b5fc6983a98390c9b196d",
  "span_id": "81b24e96b4b39940",
  "user_id": "user-123",
  "verified": true,
  "trust_score": 85
}
```

### Span Attributes

Services attach rich metadata to spans:

```go
span.SetAttributes(
    hyperion.String("user.id", userID),
    hyperion.Bool("stock.available", available),
    hyperion.Int("stock.level", stockLevel),
    hyperion.Float64("payment.amount", amount),
)
```

### Span Events

Mark important milestones:

```go
span.AddEvent("cache hit")
span.AddEvent("database query started")
span.AddEvent("payment authorized")
```

### Error Recording

Errors are captured with full context:

```go
if err != nil {
    logger.Error("operation failed", "error", err)
    span.RecordError(err)
    return err
}
```

## 🔍 Testing the Deep Call Chain

### Test 1: Basic Order Creation

```bash
curl -s -X POST http://localhost:8090/api/orders \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"u1","product_id":"p1","amount":50}' | jq .
```

**What to observe**:
- All 10 levels execute successfully
- Same `trace_id` in all logs
- Unique `span_id` for each service
- Total execution time (check span duration in HyperDX)

### Test 2: View in HyperDX

1. Open http://localhost:8080
2. Navigate to **Traces** section
3. Find your trace (search by `trace_id` from logs)
4. Click to expand the trace waterfall
5. See all 10 levels with timing information

**What you'll see**:
```
OrderService.CreateOrder                    [████████████████████] 150ms
├─ InventoryService.CheckStock              [████████████████    ] 120ms
│  └─ WarehouseService.GetStockLevel        [██████████████      ] 100ms
│     └─ ShippingService.CheckAvailability  [████████████        ]  80ms
│        └─ CarrierService.CheckCapacity    [██████████          ]  60ms
│           └─ RouteService.FindOptimalRoute[████████            ]  40ms
│              └─ GeoService.GetLocationData[██████              ]  30ms
│                 └─ CoordinateService.Get...[████                ]  20ms
│                    └─ MappingService.Get...[██                  ]  10ms
│                       └─ MonitoringServ...[                     ]   5ms
└─ PaymentService.ProcessPayment            [████████            ]  50ms
   └─ FraudDetectionService.AnalyzeRisk     [██████              ]  30ms
      └─ ... (continues down to Level 10)
```

### Test 3: Load Testing

Generate multiple orders:

```bash
for i in {1..10}; do
  curl -s -X POST http://localhost:8090/api/orders \
    -H 'Content-Type: application/json' \
    -d "{\"user_id\":\"user-$i\",\"product_id\":\"product-$i\",\"amount\":$((50 + i * 10))}" \
    > /dev/null &
done
wait
```

**What to observe**:
- 10 different `trace_id` values (one per order)
- Each trace has its own 10-level call chain
- Service map in HyperDX shows all 16 services

## 🛠️ Configuration

Edit `configs/config.yaml`:

```yaml
tracing:
  enabled: true
  service_name: "hyperion-deep-chain-example"
  exporter: "otlp"
  endpoint: "localhost:4317"  # HyperDX gRPC endpoint
  sample_rate: 1.0            # 100% sampling (perfect for demo)

log:
  level: "info"
  encoding: "json"

server:
  host: "localhost"
  port: 8090
```

## 🐛 Troubleshooting

### No Traces in HyperDX?

**Check**:
1. HyperDX is running: `docker ps | grep hyperdx`
2. Application connected: Check logs for "Connected to OTLP exporter"
3. Endpoint correct: `configs/config.yaml` should have `localhost:4317`

**Fix**: Restart both services
```bash
make hyperdx-down && make hyperdx-up
# Wait 10 seconds
make run
```

### Logs Missing trace_id?

**Cause**: Logger doesn't implement `ContextAwareLogger`

**Check**: Make sure you're using Hyperion's Zap adapter:
```go
import "github.com/mapoio/hyperion/adapters/hyperlog/zap"
```

### Service Dependencies Not Wiring?

**Cause**: Missing service in FX module

**Fix**: Check `internal/services/module.go` includes all services:
```go
var Module = fx.Module("services",
    fx.Provide(NewMonitoringService),  // All services must be listed
    fx.Provide(NewHealthCheckService),
    // ... etc
)
```

## 📂 Project Structure

```
example/otel/
├── cmd/
│   └── app/
│       └── main.go                 # Application entry point
├── internal/
│   └── services/
│       ├── module.go               # Service dependency injection
│       ├── order_service.go        # Level 1: Entry point
│       ├── payment_service.go      # Level 2
│       ├── inventory_service.go    # Level 2
│       ├── fraud_detection_service.go  # Level 3
│       ├── notification_service.go     # Level 3
│       ├── warehouse_service.go        # Level 3
│       ├── risk_analysis_service.go    # Level 4
│       ├── email_service.go            # Level 4
│       ├── sms_service.go              # Level 4
│       ├── shipping_service.go         # Level 4
│       ├── user_profile_service.go     # Level 5
│       ├── template_service.go         # Level 5
│       ├── carrier_service.go          # Level 5
│       ├── cache_service.go            # Level 6
│       ├── route_service.go            # Level 6
│       ├── geo_service.go              # Level 7
│       ├── storage_service.go          # Level 7
│       ├── replication_service.go      # Level 8
│       ├── coordinate_service.go       # Level 8
│       ├── health_check_service.go     # Level 9
│       ├── mapping_service.go          # Level 9
│       └── monitoring_service.go       # Level 10: Deepest
├── configs/
│   └── config.yaml                 # Configuration
├── docker-compose.yml              # HyperDX setup
├── Makefile                        # Convenient commands
├── go.mod                          # Go module
└── README.md                       # This file
```

## 🎓 Key Learnings

### 1. Context Propagation is Automatic

Once you use `hyperion.StartSpan()`, context flows automatically:

```go
// Level 1
ctx, span, logger := hyperion.StartSpan(ctx, tracer, logger, "Level1")
defer span.End()

// Pass ctx to Level 2 - trace_id is automatically propagated
result := level2Service.DoWork(ctx)
```

### 2. Logger Updates are Seamless

Each service gets a logger with the correct span context:

```go
// Level 1: span_id = abc123
logger.Info("at level 1")  // Logs: trace_id=xyz, span_id=abc123

// Level 2: span_id = def456 (new span)
ctx, span, logger := hyperion.StartSpan(ctx, tracer, logger, "Level2")
logger.Info("at level 2")  // Logs: trace_id=xyz, span_id=def456 (same trace, new span!)
```

### 3. Dependency Injection Scales

With 16 services and complex dependencies, manual wiring would be a nightmare. Uber FX makes it trivial:

```go
// Just provide constructors - FX figures out the order
fx.Provide(
    NewOrderService,         // Depends on: InventoryService, PaymentService
    NewInventoryService,     // Depends on: WarehouseService
    NewWarehouseService,     // Depends on: ShippingService
    // ... FX resolves the entire dependency graph automatically
)
```

### 4. Real-World Architectures Need Deep Chains

This isn't academic - real microservice architectures often have:
- **User Request** → API Gateway → Auth Service → Business Service → Cache → Database → Replication
- That's already 7 levels!
- Add monitoring, logging, feature flags, and you easily hit 10+

### 5. Observability is Essential

Without traces and logs with correlation:
- **Problem**: "Why is this request slow?"
- **Manual debugging**: Check 16 different service logs, no correlation
- **With Hyperion**: Open trace in HyperDX, see exactly which service took 200ms

## 🚀 Next Steps

1. **Add More Business Logic**: Extend services with real database calls, external APIs
2. **Add Metrics**: Use `hyperion.Meter` to record counters, histograms
3. **Add Error Scenarios**: Simulate failures at different levels, see error propagation
4. **Add Circuit Breakers**: Prevent cascading failures
5. **Deploy to Production**: Use real OTLP collector (Jaeger, Grafana Tempo, etc.)

## 📚 Resources

- **Hyperion Core Library**: `/Users/mapo/code/hyperion/hyperion/`
- **StartSpan Documentation**: `/Users/mapo/code/hyperion/docs/features/elegant-span-creation.md`
- [HyperDX Documentation](https://www.hyperdx.io/docs)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Uber FX](https://uber-go.github.io/fx/)

## 📝 License

This example is part of the Hyperion framework and follows the same license.

---

**Happy Tracing! 🔍✨**

For questions or issues, check the application logs or HyperDX UI for detailed diagnostics.
