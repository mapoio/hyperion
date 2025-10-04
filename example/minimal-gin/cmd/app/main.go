package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/mapoio/hyperion"
	"github.com/mapoio/hyperion/adapter/viper"
	"github.com/mapoio/hyperion/adapter/zap"
)

func main() {
	fx.New(
		// Core provides interface definitions and NoOp defaults
		hyperion.CoreModule,

		// Adapters provide real implementations
		viper.Module, // Config from files/env
		zap.Module,   // Structured logging

		// HTTP server
		fx.Provide(NewGinServer),
		fx.Invoke(RegisterRoutes),
		fx.Invoke(StartServer),
	).Run()
}

// NewGinServer creates a configured Gin engine
func NewGinServer(logger hyperion.Logger) *gin.Engine {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Create engine with recovery middleware
	engine := gin.New()
	engine.Use(gin.Recovery())

	return engine
}

// HyperionMiddleware creates hyperion.Context from Gin request and stores it
func HyperionMiddleware(factory hyperion.ContextFactory) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create hyperion.Context from request context
		hctx := factory.New(c.Request.Context())

		// Store in Gin context for downstream handlers
		c.Set("hctx", hctx)

		c.Next()
	}
}

// GetHyperionContext extracts hyperion.Context from Gin context
func GetHyperionContext(c *gin.Context) hyperion.Context {
	hctx, exists := c.Get("hctx")
	if !exists {
		panic("hyperion context not found - did you forget HyperionMiddleware?")
	}
	return hctx.(hyperion.Context)
}

// RegisterRoutes registers HTTP routes
func RegisterRoutes(
	engine *gin.Engine,
	factory hyperion.ContextFactory,
	config hyperion.Config,
	logger hyperion.Logger,
) {
	// Add hyperion middleware
	engine.Use(HyperionMiddleware(factory))

	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Hello endpoint with logging
	engine.GET("/hello", func(c *gin.Context) {
		hctx := GetHyperionContext(c)

		name := c.DefaultQuery("name", "World")

		// Log with context (automatically includes trace_id if tracing is enabled)
		hctx.Logger().Info("handling hello request", "name", name)

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Hello, %s!", name),
		})
	})

	// Demo endpoint showing UseIntercept pattern
	engine.GET("/demo", func(c *gin.Context) {
		hctx := GetHyperionContext(c)

		// Call a "service method" with interceptor
		result, err := demoServiceCall(hctx, c.Query("input"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": result})
	})

	// Info endpoint showing config access
	engine.GET("/info", func(c *gin.Context) {
		appName := config.GetString("app.name")
		env := config.GetString("app.env")

		c.JSON(http.StatusOK, gin.H{
			"app":         appName,
			"environment": env,
		})
	})
}

// demoServiceCall demonstrates the 3-line interceptor pattern
func demoServiceCall(hctx hyperion.Context, input string) (result string, err error) {
	// 3-line pattern: automatic tracing, logging, and metrics
	hctx, end := hctx.UseIntercept("DemoService", "Process")
	defer end(&err)

	hctx.Logger().Info("processing input", "input", input)

	// Simulate processing
	time.Sleep(50 * time.Millisecond)

	if input == "" {
		return "", fmt.Errorf("input cannot be empty")
	}

	return fmt.Sprintf("Processed: %s", input), nil
}

// StartServer starts the HTTP server with graceful shutdown
func StartServer(
	lc fx.Lifecycle,
	engine *gin.Engine,
	config hyperion.Config,
	logger hyperion.Logger,
) {
	host := config.GetString("server.host")
	if host == "" {
		host = "localhost"
	}

	port := config.GetInt("server.port")
	if port == 0 {
		port = 8080
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	server := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting HTTP server", "address", addr)

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
