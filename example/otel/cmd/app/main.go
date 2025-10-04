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
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		// ============================================================
		// STEP 1: Initialize OpenTelemetry SDK FIRST (Application Layer)
		// ============================================================
		// CRITICAL: SDK must be initialized FIRST to provide TracerProvider and MeterProvider
		telemetry.Module,                    // Provides *sdktrace.TracerProvider & *sdkmetric.MeterProvider
		telemetry.RuntimeMetricsModule,      // Automatic Go runtime metrics (CPU, Memory, GC)
		telemetry.HTTPInstrumentationModule, // Automatic HTTP tracing

		// ============================================================
		// STEP 2: Config and Adapters (provides Logger, Tracer, Meter, Database)
		// ============================================================
		viper.Module,     // Provides Config & ConfigWatcher
		zap.Module,       // Provides Logger
		hyperotel.Module, // Provides Tracer & Meter (wraps SDK providers)

		// Provide NoOp implementations for unused interfaces
		fx.Provide(hyperion.NewNoOpDatabase),

		// ============================================================
		// STEP 3: Register Interceptors (depends on Tracer from hyperotel.Module)
		// ============================================================
		// CRITICAL: Must be BEFORE CoreModule so ContextFactory can collect them
		hyperion.TracingInterceptorModule, // Enable OpenTelemetry tracing

		// ============================================================
		// STEP 4: Core Framework Infrastructure
		// ============================================================
		// CoreModule provides ContextFactory which depends on:
		// - Logger, Tracer, Database, Meter (from adapters)
		// - Interceptors from group (TracingInterceptor registered above)
		hyperion.CoreModule,

		// ============================================================
		// STEP 5: Business Logic
		// ============================================================
		services.Module,

		// ============================================================
		// STEP 5: HTTP Server
		// ============================================================
		fx.Provide(NewHTTPServer),
		fx.Invoke(RegisterRoutes),
		fx.Invoke(StartServer),
	)

	app.Run()
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
	externalAPIService *services.ExternalAPIService,
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

	// External API endpoints - demonstrate HTTP client tracing
	router.GET("/api/external/user/:id", func(c *gin.Context) {
		hctx := factory.New(c.Request.Context())

		// Create root span for the HTTP request
		hctx, rootSpan := hctx.Tracer().Start(hctx, "GET /api/external/user/:id")
		defer rootSpan.End()

		// Parse user ID
		userID := 1
		if id := c.Param("id"); id != "" {
			_, _ = fmt.Sscanf(id, "%d", &userID)
		}

		hctx.Logger().Info("fetching external user", "user_id", userID)

		// Call external API - this will create a child span for HTTP client call
		user, err := externalAPIService.GetRandomUser(hctx, userID)
		if err != nil {
			hctx.Logger().Error("failed to fetch user", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
	})

	router.GET("/api/external/post/:id", func(c *gin.Context) {
		hctx := factory.New(c.Request.Context())

		// Create root span for the HTTP request
		hctx, rootSpan := hctx.Tracer().Start(hctx, "GET /api/external/post/:id")
		defer rootSpan.End()

		// Parse post ID
		postID := 1
		if id := c.Param("id"); id != "" {
			_, _ = fmt.Sscanf(id, "%d", &postID)
		}

		hctx.Logger().Info("fetching external post", "post_id", postID)

		// Call external API - this will create a child span for HTTP client call
		post, err := externalAPIService.GetRandomPost(hctx, postID)
		if err != nil {
			hctx.Logger().Error("failed to fetch post", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, post)
	})

	router.GET("/api/external/ip", func(c *gin.Context) {
		hctx := factory.New(c.Request.Context())

		// Create root span for the HTTP request
		hctx, rootSpan := hctx.Tracer().Start(hctx, "GET /api/external/ip")
		defer rootSpan.End()

		hctx.Logger().Info("fetching IP geolocation")

		// Call external API - this will create a child span for HTTP client call
		info, err := externalAPIService.GetIPInfo(hctx)
		if err != nil {
			hctx.Logger().Error("failed to fetch IP info", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, info)
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