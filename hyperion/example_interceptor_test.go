package hyperion_test

import (
	"context"
	"errors"

	"go.uber.org/fx"

	hyperion "github.com/mapoio/hyperion"
)

// Example_interceptorBasicUsage demonstrates basic interceptor usage
// with the UseIntercept pattern.
func Example_interceptorBasicUsage() {
	// Create an fx app with CoreModule and enable built-in interceptors
	app := fx.New(
		hyperion.CoreModule,               // Core infrastructure (no interceptors)
		hyperion.TracingInterceptorModule, // Enable tracing interceptor
		hyperion.LoggingInterceptorModule, // Enable logging interceptor

		fx.Provide(NewUserService),
		fx.Invoke(func(service *UserService, factory hyperion.ContextFactory) {
			ctx := factory.New(context.Background())

			// Use the service
			_ = service.GetUser(ctx, "user123")
		}),

		fx.NopLogger, // Suppress fx startup logs in example
	)

	_ = app.Err()
	// Output:
	// [Hyperion] Using no-op Logger
	// [Hyperion] Using no-op Tracer
	// [Hyperion] Using no-op Database
	// [Hyperion] Using no-op Meter
}

// Example_interceptorSelectiveApplication shows how to selectively apply interceptors.
func Example_interceptorSelectiveApplication() {
	app := fx.New(
		hyperion.CoreModule,
		hyperion.AllInterceptorsModule, // Enable all built-in interceptors

		fx.Provide(NewUserService),
		fx.Invoke(func(service *UserService, factory hyperion.ContextFactory) {
			ctx := factory.New(context.Background())

			// Only apply tracing (exclude logging)
			_ = service.GetUserWithTracingOnly(ctx, "user123")

			// Only apply logging (exclude tracing)
			_ = service.GetUserWithLoggingOnly(ctx, "user123")
		}),

		fx.NopLogger,
	)

	_ = app.Err()
	// Output:
	// [Hyperion] Using no-op Logger
	// [Hyperion] Using no-op Tracer
	// [Hyperion] Using no-op Database
	// [Hyperion] Using no-op Meter
}

// Example_interceptorCustomInterceptor demonstrates adding a custom interceptor.
func Example_interceptorCustomInterceptor() {
	app := fx.New(
		hyperion.CoreModule,
		hyperion.AllInterceptorsModule, // Enable built-in interceptors

		// Register custom interceptor via fx group
		fx.Provide(
			fx.Annotate(
				NewMetricsInterceptor,
				fx.ResultTags(`group:"hyperion.interceptors"`),
			),
		),

		fx.Provide(NewUserService),
		fx.Invoke(func(service *UserService, factory hyperion.ContextFactory) {
			ctx := factory.New(context.Background())

			// All interceptors (tracing, logging, metrics) will be applied
			_ = service.GetUser(ctx, "user123")
		}),

		fx.NopLogger,
	)

	_ = app.Err()
	// Output:
	// [Hyperion] Using no-op Logger
	// [Hyperion] Using no-op Tracer
	// [Hyperion] Using no-op Database
	// [Hyperion] Using no-op Meter
}

// UserService demonstrates service-layer interceptor usage.
type UserService struct {
	factory hyperion.ContextFactory
}

func NewUserService(factory hyperion.ContextFactory) *UserService {
	return &UserService{factory: factory}
}

// GetUser demonstrates the 3-line interceptor pattern.
func (s *UserService) GetUser(ctx hyperion.Context, userID string) (err error) {
	// Apply all registered interceptors
	ctx, end := ctx.UseIntercept("UserService", "GetUser")
	defer end(&err)

	// Business logic
	if userID == "" {
		return errors.New("user ID is required")
	}

	ctx.Logger().Info("fetching user", "userID", userID)

	// Simulate database call
	_ = ctx.DB()

	return nil
}

// GetUserWithTracingOnly demonstrates selective interceptor application.
func (s *UserService) GetUserWithTracingOnly(ctx hyperion.Context, userID string) (err error) {
	// Only apply tracing interceptor
	ctx, end := ctx.UseIntercept("UserService", "GetUserWithTracingOnly",
		hyperion.WithOnly("tracing"))
	defer end(&err)

	ctx.Logger().Info("fetching user with tracing only", "userID", userID)
	return nil
}

// GetUserWithLoggingOnly demonstrates selective interceptor application.
func (s *UserService) GetUserWithLoggingOnly(ctx hyperion.Context, userID string) (err error) {
	// Only apply logging interceptor
	ctx, end := ctx.UseIntercept("UserService", "GetUserWithLoggingOnly",
		hyperion.WithOnly("logging"))
	defer end(&err)

	ctx.Logger().Info("fetching user with logging only", "userID", userID)
	return nil
}

// MetricsInterceptor is a custom interceptor for demonstration.
type MetricsInterceptor struct{}

func NewMetricsInterceptor() hyperion.Interceptor {
	return &MetricsInterceptor{}
}

func (m *MetricsInterceptor) Name() string {
	return "metrics"
}

func (m *MetricsInterceptor) Intercept(
	ctx hyperion.Context,
	fullPath string,
) (hyperion.Context, func(err *error), error) {
	// Record metrics here
	end := func(errPtr *error) {
		// Record completion metrics
	}
	return ctx, end, nil
}

func (m *MetricsInterceptor) Order() int {
	return 300 // After tracing and logging
}
