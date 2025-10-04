package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mapoio/hyperion"
	hyperotel "github.com/mapoio/hyperion/adapter/otel"
	"github.com/mapoio/hyperion/adapter/viper"
	"github.com/mapoio/hyperion/adapter/zap"
	"github.com/mapoio/hyperion/example/otel/internal/services"
	"github.com/mapoio/hyperion/example/otel/internal/telemetry"
	"go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		// ============================================================
		// STEP 1: Initialize OpenTelemetry SDK (Application Layer)
		// ============================================================
		// This is where the application has FULL CONTROL over OTel configuration.
		// The SDK initialization happens ONCE and is shared across all components.
		telemetry.Module,            // OTel SDK with TracerProvider & MeterProvider
		telemetry.RuntimeMetricsModule, // Automatic Go runtime metrics (CPU, Memory, GC)
		telemetry.HTTPInstrumentationModule, // Automatic HTTP tracing

		// ============================================================
		// STEP 2: Integrate Hyperion Adapters with OTel SDK
		// ============================================================
		// Hyperion adapters use the OTel SDK initialized in Step 1.
		// They provide hyperion.Tracer and hyperion.Meter interfaces.
		fx.Provide(
			func(tp *sdktrace.TracerProvider) hyperion.Tracer {
				return hyperotel.NewOtelTracerFromProvider(tp, "hyperion-otel-example")
			},
		),
		fx.Provide(
			func(mp *metric.MeterProvider) hyperion.Meter {
				return hyperotel.NewOtelMeterFromProvider(mp, "hyperion-otel-example")
			},
		),

		// ============================================================
		// STEP 3: Other Hyperion Dependencies
		// ============================================================
		viper.Module, // Configuration
		zap.Module,   // Logger with OTel trace context injection

		fx.Provide(hyperion.NewNoOpDatabase),
		fx.Provide(hyperion.NewNoOpCache),

		// Register TracingInterceptor
		fx.Provide(hyperion.NewTracingInterceptor),

		// Register ContextFactory
		fx.Provide(
			func(
				logger hyperion.Logger,
				tracer hyperion.Tracer,
				db hyperion.Database,
				meter hyperion.Meter,
				tracingInterceptor *hyperion.TracingInterceptor,
			) hyperion.ContextFactory {
				interceptors := []hyperion.Interceptor{tracingInterceptor}
				logger.Info("🔍 [DEBUG] ContextFactory provider called",
					"interceptors_count", len(interceptors),
				)
				return hyperion.NewContextFactory(
					logger,
					tracer,
					db,
					meter,
					hyperion.WithInterceptors(interceptors...),
				)
			},
		),

		// ============================================================
		// STEP 4: Business Logic
		// ============================================================
		services.Module,

		// ============================================================
		// STEP 5: HTTP Server
		// ============================================================
		fx.Provide(NewHTTPServer),
		fx.Invoke(RegisterRoutes),
		fx.Invoke(StartServer),
	).Run()
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// NewHTTPServer creates a new Gin HTTP server
func NewHTTPServer() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	return router
}

// RegisterRoutes registers all HTTP routes
func RegisterRoutes(
	router *gin.Engine,
	config hyperion.Config,
	factory hyperion.ContextFactory, // ⭐ ContextFactory 用于创建 hyperion.Context
	logger hyperion.Logger,
	orderService *services.OrderService,
) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Deep call chain endpoint - demonstrates 10-level service calls!
	router.POST("/api/orders", func(c *gin.Context) {
		// ⭐ 关键：使用 ContextFactory 从 gin.Context 创建 hyperion.Context
		hctx := factory.New(c.Request.Context())

		// Create root span for the HTTP request
		hctx, rootSpan := hctx.Tracer().Start(hctx, "POST /api/orders")
		defer rootSpan.End()

		// Parse request
		var req struct {
			UserID    string  `json:"user_id"`
			ProductID string  `json:"product_id"`
			Amount    float64 `json:"amount"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			hctx.Logger().Error("invalid request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		hctx.Logger().Info("creating order",
			"user_id", req.UserID,
			"product_id", req.ProductID,
			"amount", req.Amount,
		)

		// ⭐ Service call - 自动触发拦截器 (tracing + logging)
		// CreateOrder 内部使用 ctx.UseIntercept() 自动创建 span
		orderID, err := orderService.CreateOrder(hctx, req.UserID, req.ProductID, req.Amount)
		if err != nil {
			hctx.Logger().Error("order creation failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		hctx.Logger().Info("order created successfully", "order_id", orderID)

		c.JSON(http.StatusOK, gin.H{
			"order_id":   orderID,
			"user_id":    req.UserID,
			"product_id": req.ProductID,
			"amount":     req.Amount,
			"status":     "created",
		})
	})
}

// StartServer starts the HTTP server
func StartServer(
	lc fx.Lifecycle,
	router *gin.Engine,
	config hyperion.Config,
	logger hyperion.Logger,
) {
	var cfg ServerConfig
	if err := config.Unmarshal("server", &cfg); err != nil {
		cfg = ServerConfig{
			Host:         "localhost",
			Port:         8090,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting HTTP server", "address", server.Addr)

			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("HTTP server error", "error", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping HTTP server")
			return server.Shutdown(ctx)
		},
	})
}
