// Package hyperion provides a production-ready context implementation with
// factory pattern, decorator support, and middleware infrastructure for
// building scalable Go applications with clean AOP-style cross-cutting concerns.
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
// # Decorator Pattern for AOP
//
// Decorators enable cross-cutting concerns without modifying core logic.
// The generic Decorator[T] type works with any component:
//
//	// Built-in components
//	factory := hyperion.NewContextFactory(logger, tracer, db,
//	    hyperion.WithLoggerDecorator(AddPrefixDecorator("[APP]")),
//	    hyperion.WithExecutorDecorator(QueryLoggingDecorator(logger)),
//	)
//
//	// User-defined components
//	type MyCache interface {
//	    Get(key string) ([]byte, error)
//	}
//
//	cache = hyperion.Chain[MyCache](
//	    MetricsDecorator(metrics),
//	    LoggingDecorator(logger),
//	)(cache)
//
// # Middleware for Service Layer
//
// Middleware wraps service methods for logging, tracing, transactions:
//
//	middleware := hyperion.ChainMiddleware(
//	    LoggingMiddleware,
//	    TracingMiddleware,
//	)
//
//	err := middleware(ctx, func(ctx hyperion.Context) error {
//	    return service.DoWork(ctx)
//	})
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
// It provides access to core dependencies (Logger, DB, Tracer) and
// extends the standard context.Context interface.
type Context interface {
	context.Context

	// Logger returns the logger associated with this context.
	Logger() Logger

	// DB returns the database executor associated with this context.
	// When inside a transaction (via UnitOfWork.WithTransaction),
	// this returns the transaction executor.
	DB() Executor

	// Tracer returns the tracer associated with this context.
	Tracer() Tracer

	// WithTimeout returns a copy of the context with the specified timeout.
	WithTimeout(timeout time.Duration) (Context, context.CancelFunc)

	// WithCancel returns a copy of the context that can be canceled.
	WithCancel() (Context, context.CancelFunc)

	// WithDeadline returns a copy of the context with the specified deadline.
	WithDeadline(deadline time.Time) (Context, context.CancelFunc)
}

// New creates a new Hyperion context.
func New(
	ctx context.Context,
	logger Logger,
	db Executor,
	tracer Tracer,
) Context {
	return &hyperionContext{
		Context: ctx,
		logger:  logger,
		db:      db,
		tracer:  tracer,
	}
}

// hyperionContext is the default implementation of Context.
type hyperionContext struct {
	context.Context
	logger Logger
	db     Executor
	tracer Tracer
}

func (c *hyperionContext) Logger() Logger {
	return c.logger
}

func (c *hyperionContext) DB() Executor {
	return c.db
}

func (c *hyperionContext) Tracer() Tracer {
	return c.tracer
}

// withContext is a helper method to create a new hyperionContext with a different underlying context.
// It preserves all the other fields (logger, db, tracer) from the current context.
func (c *hyperionContext) withContext(ctx context.Context) *hyperionContext {
	return &hyperionContext{
		Context: ctx,
		logger:  c.logger,
		db:      c.db,
		tracer:  c.tracer,
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
//	// txCtx has new DB, but same logger and tracer
func WithDB(ctx Context, db Executor) Context {
	hctx, ok := ctx.(*hyperionContext)
	if !ok {
		// Fallback: create new context
		return New(ctx, ctx.Logger(), db, ctx.Tracer())
	}

	return &hyperionContext{
		Context: hctx.Context,
		logger:  hctx.logger,
		db:      db, // Replace DB
		tracer:  hctx.tracer,
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
		return New(ctx, logger, ctx.DB(), ctx.Tracer())
	}

	return &hyperionContext{
		Context: hctx.Context,
		logger:  logger, // Replace Logger
		db:      hctx.db,
		tracer:  hctx.tracer,
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
		return New(ctx, ctx.Logger(), ctx.DB(), tracer)
	}

	return &hyperionContext{
		Context: hctx.Context,
		logger:  hctx.logger,
		db:      hctx.db,
		tracer:  tracer, // Replace Tracer
	}
}
