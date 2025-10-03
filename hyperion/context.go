// Package hyperion provides a production-ready context implementation with
// factory pattern and interceptor infrastructure for building scalable Go
// applications with clean AOP-style cross-cutting concerns.
//
// # Context Factory Pattern
//
// The ContextFactory enables clean dependency injection with fx:
//
//	type Handler struct {
//	    factory hyperion.ContextFactory
//	}
//
//	func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	    ctx := h.factory.New(r.Context())
//	    ctx.Logger().Info("handling request")
//	}
//
// # Interceptor Pattern for AOP
//
// Interceptors provide explicit control over cross-cutting concerns using
// a 3-line pattern:
//
//	func (s *UserService) GetUser(ctx hyperion.Context, id string) (err error) {
//	    ctx, end := ctx.UseIntercept("UserService", "GetUser")
//	    defer end(&err)
//
//	    // Business logic here
//	    return nil
//	}
//
// Built-in interceptors (optional modules):
//   - TracingInterceptorModule: OpenTelemetry distributed tracing
//   - LoggingInterceptorModule: Structured method logging
//
// Custom interceptors can be registered via fx groups:
//
//	fx.Provide(
//	    fx.Annotate(
//	        NewMyInterceptor,
//	        fx.ResultTags(`group:"hyperion.interceptors"`),
//	    ),
//	)
//
// # Immutability
//
// All helper functions (WithLogger, WithTracer, WithDB) return new contexts:
//
//	requestLogger := logger.With("requestID", requestID)
//	requestCtx := hyperion.WithLogger(ctx, requestLogger)
//	// Original ctx unchanged, requestCtx has new logger
package hyperion

import (
	"context"
	"time"
)

// Context is the type-safe context for Hyperion applications.
// It provides access to core dependencies (Logger, DB, Tracer, Meter) and
// extends the standard context.Context interface.
type Context interface {
	context.Context

	// Logger returns the logger associated with this context.
	// When using OpenTelemetry adapters, logs are automatically correlated
	// with the current trace via trace ID and span ID.
	Logger() Logger

	// DB returns the database executor associated with this context.
	// When inside a transaction (via UnitOfWork.WithTransaction),
	// this returns the transaction executor.
	DB() Executor

	// Tracer returns the tracer associated with this context.
	Tracer() Tracer

	// Meter returns the meter for recording metrics.
	// When using OpenTelemetry adapters, metrics are automatically correlated
	// with traces via exemplars, enabling metrics → traces navigation.
	Meter() Meter

	// Span returns the current span from this context.
	// If no span is active, returns a no-op span.
	// This enables accessing span operations without re-calling tracer.Start().
	Span() Span

	// WithTimeout returns a copy of the context with the specified timeout.
	WithTimeout(timeout time.Duration) (Context, context.CancelFunc)

	// WithCancel returns a copy of the context that can be canceled.
	WithCancel() (Context, context.CancelFunc)

	// WithDeadline returns a copy of the context with the specified deadline.
	WithDeadline(deadline time.Time) (Context, context.CancelFunc)

	// UseIntercept applies registered interceptors for a method call.
	//
	// Parameters:
	//   - parts: Path segments that will be joined with "." to form full path.
	//            String parts are joined, InterceptOption parts configure the call.
	//
	// Example usage:
	//
	//   // Basic usage (applies all registered global interceptors)
	//   ctx, end := ctx.UseIntercept("UserService", "GetUser")
	//   defer end(&err)
	//
	//   // With namespace
	//   ctx, end := ctx.UseIntercept("Service", "User", "GetUser")
	//   defer end(&err)
	//
	//   // Only apply tracing
	//   ctx, end := ctx.UseIntercept("UserService", "GetUser",
	//       hyperion.WithOnly("tracing"))
	//   defer end(&err)
	//
	//   // Exclude logging (for high-frequency calls)
	//   ctx, end := ctx.UseIntercept("UserService", "GetUser",
	//       hyperion.WithExclude("logging"))
	//   defer end(&err)
	//
	//   // Add custom interceptor
	//   ctx, end := ctx.UseIntercept("UserService", "GetUser",
	//       hyperion.WithAdditional(myCustomInterceptor))
	//   defer end(&err)
	UseIntercept(parts ...any) (ctx Context, end func(err *error))
}

// New creates a new Hyperion context.
func New(
	ctx context.Context,
	logger Logger,
	db Executor,
	tracer Tracer,
	meter Meter,
) Context {
	return &hyperionContext{
		Context: ctx,
		logger:  logger,
		db:      db,
		tracer:  tracer,
		meter:   meter,
	}
}

// hyperionContext is the default implementation of Context.
type hyperionContext struct {
	context.Context
	logger       Logger
	db           Executor
	tracer       Tracer
	meter        Meter
	span         Span          // Current active span (nil if no span)
	interceptors []Interceptor // Global interceptors from fx (can be empty)
}

func (c *hyperionContext) Logger() Logger {
	// If logger implements ContextAwareLogger, bind it to current context
	// for automatic trace context injection
	if contextAware, ok := c.logger.(ContextAwareLogger); ok {
		return contextAware.WithContext(c.Context)
	}
	return c.logger
}

func (c *hyperionContext) DB() Executor {
	return c.db
}

func (c *hyperionContext) Tracer() Tracer {
	return c.tracer
}

func (c *hyperionContext) Meter() Meter {
	return c.meter
}

func (c *hyperionContext) Span() Span {
	if c.span != nil {
		return c.span
	}
	// Return no-op span if no span is active
	return &noopSpan{}
}

// withContext is a helper method to create a new hyperionContext with a different underlying context.
// It preserves all the other fields (logger, db, tracer, meter, span, interceptors) from the current context.
func (c *hyperionContext) withContext(ctx context.Context) *hyperionContext {
	return &hyperionContext{
		Context:      ctx,
		logger:       c.logger,
		db:           c.db,
		tracer:       c.tracer,
		meter:        c.meter,
		span:         c.span,
		interceptors: c.interceptors,
	}
}

func (c *hyperionContext) WithTimeout(timeout time.Duration) (Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Context, timeout)
	return c.withContext(ctx), cancel
}

func (c *hyperionContext) WithCancel() (Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c.Context)
	return c.withContext(ctx), cancel
}

func (c *hyperionContext) WithDeadline(deadline time.Time) (Context, context.CancelFunc) {
	ctx, cancel := context.WithDeadline(c.Context, deadline)
	return c.withContext(ctx), cancel
}

// WithDB returns a new Context with the specified database executor.
// This creates an immutable copy with the DB replaced.
// This is used internally by UnitOfWork to inject transaction executors.
//
// Example:
//
//	txCtx := hyperion.WithDB(ctx, transactionExecutor)
//	// txCtx has new DB, but same logger, tracer, and meter
func WithDB(ctx Context, db Executor) Context {
	hctx, ok := ctx.(*hyperionContext)
	if !ok {
		// Fallback: create new context
		return New(ctx, ctx.Logger(), db, ctx.Tracer(), ctx.Meter())
	}

	return &hyperionContext{
		Context:      hctx.Context,
		logger:       hctx.logger,
		db:           db, // Replace DB
		tracer:       hctx.tracer,
		meter:        hctx.meter,
		span:         hctx.span,
		interceptors: hctx.interceptors,
	}
}

// WithLogger returns a new Context with the specified logger.
// This creates an immutable copy with the Logger replaced.
//
// Example:
//
//	requestLogger := logger.With("requestID", requestID)
//	requestCtx := hyperion.WithLogger(ctx, requestLogger)
func WithLogger(ctx Context, logger Logger) Context {
	hctx, ok := ctx.(*hyperionContext)
	if !ok {
		// Fallback: create new context
		return New(ctx, logger, ctx.DB(), ctx.Tracer(), ctx.Meter())
	}

	return &hyperionContext{
		Context:      hctx.Context,
		logger:       logger, // Replace Logger
		db:           hctx.db,
		tracer:       hctx.tracer,
		meter:        hctx.meter,
		span:         hctx.span,
		interceptors: hctx.interceptors,
	}
}

// WithTracer returns a new Context with the specified tracer.
// This creates an immutable copy with the Tracer replaced.
//
// Example:
//
//	customTracer := NewCustomTracer()
//	tracedCtx := hyperion.WithTracer(ctx, customTracer)
func WithTracer(ctx Context, tracer Tracer) Context {
	hctx, ok := ctx.(*hyperionContext)
	if !ok {
		// Fallback: create new context
		return New(ctx, ctx.Logger(), ctx.DB(), tracer, ctx.Meter())
	}

	return &hyperionContext{
		Context:      hctx.Context,
		logger:       hctx.logger,
		db:           hctx.db,
		tracer:       tracer, // Replace Tracer
		meter:        hctx.meter,
		span:         hctx.span,
		interceptors: hctx.interceptors,
	}
}

// WithContext returns a new hyperion.Context with the underlying context.Context replaced.
// This is useful when a tracer or other component returns a new standard context
// (e.g., with trace context) and you need to wrap it back into hyperion.Context.
//
// Example:
//
//	stdCtx, span := otelTracer.Start(hctx, "operation")
//	newHctx := hyperion.WithContext(hctx, stdCtx)
func WithContext(ctx Context, stdCtx context.Context) Context {
	hctx, ok := ctx.(*hyperionContext)
	if !ok {
		// Fallback: create new context
		return New(stdCtx, ctx.Logger(), ctx.DB(), ctx.Tracer(), ctx.Meter())
	}

	return &hyperionContext{
		Context:      stdCtx, // Replace underlying context
		logger:       hctx.logger,
		db:           hctx.db,
		tracer:       hctx.tracer,
		meter:        hctx.meter,
		span:         hctx.span,
		interceptors: hctx.interceptors,
	}
}

// WithSpan returns a new Context with the specified span.
// This is used internally when tracer.Start() creates a new span.
//
// Example:
//
//	newCtx, span := tracer.Start(ctx, "operation")
//	// newCtx already has the span set via WithSpan
func WithSpan(ctx Context, span Span) Context {
	hctx, ok := ctx.(*hyperionContext)
	if !ok {
		// Fallback: create new context
		return New(ctx, ctx.Logger(), ctx.DB(), ctx.Tracer(), ctx.Meter())
	}

	return &hyperionContext{
		Context:      hctx.Context,
		logger:       hctx.logger,
		db:           hctx.db,
		tracer:       hctx.tracer,
		meter:        hctx.meter,
		span:         span, // Set new span
		interceptors: hctx.interceptors,
	}
}

// UseIntercept implements Context.UseIntercept.
// It applies registered interceptors based on the provided configuration.
func (c *hyperionContext) UseIntercept(parts ...any) (ctx Context, endFunc func(err *error)) {
	// Parse path and options
	fullPath, opts := JoinPath(parts...)

	// Build configuration
	config := &InterceptConfig{}
	for _, opt := range opts {
		opt(config)
	}

	// Select interceptors to apply
	selectedInterceptors := c.selectInterceptors(config)

	if len(selectedInterceptors) == 0 {
		// No interceptors to apply, return no-op
		return c, func(*error) {}
	}

	// Apply interceptors in order
	currentCtx := Context(c)
	endFuncs := make([]func(err *error), 0, len(selectedInterceptors))

	for _, interceptor := range selectedInterceptors {
		newCtx, end, err := interceptor.Intercept(currentCtx, fullPath)
		if err != nil {
			c.logger.Error("Interceptor error",
				"interceptor", interceptor.Name(),
				"path", fullPath,
				"error", err,
			)
			continue
		}

		currentCtx = newCtx
		endFuncs = append(endFuncs, end)
	}

	// Combined end function (calls all end functions in reverse order - LIFO)
	combinedEnd := func(errPtr *error) {
		for i := len(endFuncs) - 1; i >= 0; i-- {
			endFuncs[i](errPtr)
		}
	}

	return currentCtx, combinedEnd
}

// selectInterceptors selects which interceptors to apply based on config.
func (c *hyperionContext) selectInterceptors(config *InterceptConfig) []Interceptor {
	selected := make([]Interceptor, 0)

	// First, add additional interceptors (if any)
	// These are added first so they execute in the outer layer
	selected = append(selected, config.Additional...)

	// Then, add global interceptors (filtered by config)
	for _, interceptor := range c.interceptors {
		if config.shouldApply(interceptor.Name()) {
			selected = append(selected, interceptor)
		}
	}

	// Interceptors should already be sorted by Order() when registered
	// But we'll sort again to ensure correct order with additional interceptors
	return sortInterceptors(selected)
}

// sortInterceptors sorts interceptors by their Order() value.
func sortInterceptors(interceptors []Interceptor) []Interceptor {
	// Create a copy to avoid modifying the input slice
	sorted := make([]Interceptor, len(interceptors))
	copy(sorted, interceptors)

	// Simple insertion sort (efficient for small lists)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j].Order() < sorted[j-1].Order(); j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}

	return sorted
}
