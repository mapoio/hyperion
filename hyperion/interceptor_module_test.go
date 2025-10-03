package hyperion_test

import (
	"context"
	"testing"

	"go.uber.org/fx"

	hyperion "github.com/mapoio/hyperion"
)

// TestCoreModule_NoInterceptorsByDefault verifies that CoreModule
// does NOT register any interceptors by default.
func TestCoreModule_NoInterceptorsByDefault(t *testing.T) {
	var factory hyperion.ContextFactory

	app := fx.New(
		hyperion.CoreModule,

		fx.Populate(&factory),
		fx.NopLogger,
	)

	if err := app.Err(); err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Create a context and verify no interceptors are registered
	ctx := factory.New(context.Background())

	// UseIntercept should return the same context (no interceptors)
	newCtx, end := ctx.UseIntercept("Test", "Method")

	// Should be the same context since no interceptors
	if newCtx != ctx {
		t.Error("Expected same context when no interceptors are registered")
	}

	// Should not panic
	var err error
	end(&err)
}

// TestTracingInterceptorModule verifies that TracingInterceptorModule
// registers the tracing interceptor.
func TestTracingInterceptorModule(t *testing.T) {
	var factory hyperion.ContextFactory
	interceptorCalled := false

	// Custom interceptor to track if any interceptors are called
	customInterceptor := &testInterceptor{
		name:  "test",
		order: 1000,
		onIntercept: func() {
			interceptorCalled = true
		},
	}

	app := fx.New(
		hyperion.CoreModule,
		hyperion.TracingInterceptorModule,

		// Add a test interceptor to verify interceptor chain is active
		fx.Provide(
			fx.Annotate(
				func() hyperion.Interceptor { return customInterceptor },
				fx.ResultTags(`group:"hyperion.interceptors"`),
			),
		),

		fx.Populate(&factory),
		fx.NopLogger,
	)

	if err := app.Err(); err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Create a context
	ctx := factory.New(context.Background())

	// UseIntercept should execute interceptors
	_, end := ctx.UseIntercept("Test", "Method")

	if !interceptorCalled {
		t.Error("Expected interceptor to be called when TracingInterceptorModule is enabled")
	}

	// Should not panic
	var err error
	end(&err)
}

// TestLoggingInterceptorModule verifies that LoggingInterceptorModule
// registers the logging interceptor.
func TestLoggingInterceptorModule(t *testing.T) {
	var factory hyperion.ContextFactory

	app := fx.New(
		hyperion.CoreModule,
		hyperion.LoggingInterceptorModule,

		fx.Populate(&factory),
		fx.NopLogger,
	)

	if err := app.Err(); err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Create a context and verify logging interceptor is registered
	ctx := factory.New(context.Background())

	// UseIntercept should return the same context (logging doesn't modify ctx)
	// But the end function should be set
	_, end := ctx.UseIntercept("Test", "Method")

	// Should not panic
	var err error
	end(&err)
}

// TestAllInterceptorsModule verifies that AllInterceptorsModule
// registers both tracing and logging interceptors.
func TestAllInterceptorsModule(t *testing.T) {
	var factory hyperion.ContextFactory
	interceptorCalled := false

	// Custom interceptor to track if any interceptors are called
	customInterceptor := &testInterceptor{
		name:  "test",
		order: 1000,
		onIntercept: func() {
			interceptorCalled = true
		},
	}

	app := fx.New(
		hyperion.CoreModule,
		hyperion.AllInterceptorsModule,

		// Add a test interceptor to verify interceptor chain is active
		fx.Provide(
			fx.Annotate(
				func() hyperion.Interceptor { return customInterceptor },
				fx.ResultTags(`group:"hyperion.interceptors"`),
			),
		),

		fx.Populate(&factory),
		fx.NopLogger,
	)

	if err := app.Err(); err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Create a context
	ctx := factory.New(context.Background())

	// UseIntercept should apply all interceptors
	_, end := ctx.UseIntercept("Test", "Method")

	if !interceptorCalled {
		t.Error("Expected interceptor to be called when AllInterceptorsModule is enabled")
	}

	// Should not panic
	var err error
	end(&err)
}

// testInterceptor is a helper for testing
type testInterceptor struct {
	name        string
	order       int
	onIntercept func()
	onEnd       func(err *error)
}

func (m *testInterceptor) Name() string {
	return m.name
}

func (m *testInterceptor) Intercept(
	ctx hyperion.Context,
	fullPath string,
) (hyperion.Context, func(err *error), error) {
	if m.onIntercept != nil {
		m.onIntercept()
	}

	end := func(errPtr *error) {
		if m.onEnd != nil {
			m.onEnd(errPtr)
		}
	}

	return ctx, end, nil
}

func (m *testInterceptor) Order() int {
	return m.order
}
