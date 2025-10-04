package hyperion_test

import (
	"context"
	"testing"

	"go.uber.org/fx"

	"github.com/mapoio/hyperion"
)

// TestCoreModule tests that CoreModule provides all default implementations
func TestCoreModule(t *testing.T) {
	app := fx.New(
		hyperion.CoreModule,
		fx.Invoke(func(
			logger hyperion.Logger,
			tracer hyperion.Tracer,
			db hyperion.Database,
			cfg hyperion.Config,
			cache hyperion.Cache,
		) {
			// Verify all dependencies are provided
			if logger == nil {
				t.Error("Logger should not be nil")
			}
			if tracer == nil {
				t.Error("Tracer should not be nil")
			}
			if db == nil {
				t.Error("Database should not be nil")
			}
			if cfg == nil {
				t.Error("Config should not be nil")
			}
			if cache == nil {
				t.Error("Cache should not be nil")
			}

			// Test Logger
			logger.Info("test message", "key", "value")

			// Test Tracer
			hctx := hyperion.New(context.Background(), logger, nil, tracer, nil)
			newCtx, span := tracer.Start(hctx, "test-span")
			span.SetAttributes(hyperion.String("test", "value"))
			span.End()
			_ = newCtx

			// Test Config
			value := cfg.GetString("nonexistent")
			if value != "" {
				t.Errorf("Expected empty string, got %s", value)
			}

			t.Log("All default implementations working correctly")
		}),
		fx.NopLogger,
	)

	if err := app.Start(context.Background()); err != nil {
		t.Fatalf("Failed to start app: %v", err)
	}
	if err := app.Stop(context.Background()); err != nil {
		t.Errorf("Failed to stop app: %v", err)
	}
}

// TestCoreWithoutDefaultsModule tests that CoreWithoutDefaultsModule fails without adapters
func TestCoreWithoutDefaultsModule(t *testing.T) {
	app := fx.New(
		hyperion.CoreWithoutDefaultsModule,
		fx.Invoke(func(logger hyperion.Logger) {
			t.Error("Should not reach here - missing Logger dependency")
		}),
		fx.NopLogger,
	)

	err := app.Start(context.Background())
	if err == nil {
		t.Fatal("Expected error due to missing Logger, but app started successfully")
	}
	t.Logf("Expected error occurred: %v", err)
}
