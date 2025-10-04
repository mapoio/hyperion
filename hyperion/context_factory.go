package hyperion

import (
	"context"
)

// ContextFactory creates new Hyperion contexts with injected dependencies.
// It is designed to be used with fx dependency injection.
//
// Example usage with fx:
//
//	type Handler struct {
//	    factory ContextFactory
//	}
//
//	func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	    ctx := h.factory.New(r.Context())
//	    ctx.Logger().Info("handling request")
//	    ctx.Meter().Counter("requests").Add(ctx, 1)
//	}
type ContextFactory interface {
	// New creates a new Hyperion context from a standard context.
	// The returned context will have Logger, Tracer, DB, and Meter injected.
	// Interceptors are dynamically fetched from the InterceptorRegistry.
	New(ctx context.Context) Context
}

// contextFactory is the default implementation of ContextFactory.
type contextFactory struct {
	logger   Logger
	tracer   Tracer
	db       Database
	meter    Meter
	registry InterceptorRegistry // Registry to dynamically fetch interceptors
}

// NewContextFactory creates a new ContextFactory with the given dependencies.
// Dependencies are typically injected via fx.Provide.
//
// Example:
//
//	var ContextModule = fx.Module("hyperion.context",
//	    fx.Provide(NewContextFactory),
//	)
func NewContextFactory(logger Logger, tracer Tracer, db Database, meter Meter, opts ...FactoryOption) ContextFactory {
	f := &contextFactory{
		logger: logger,
		tracer: tracer,
		db:     db,
		meter:  meter,
	}

	// Apply options
	for _, opt := range opts {
		opt(f)
	}

	return f
}

// New creates a new Hyperion context with injected dependencies.
// Interceptors are dynamically fetched from the registry at context creation time.
func (f *contextFactory) New(ctx context.Context) Context {
	// Dynamically fetch interceptors from registry
	var interceptors []Interceptor
	if f.registry != nil {
		interceptors = f.registry.GetAll()
	}

	return &hyperionContext{
		Context:      ctx,
		logger:       f.logger,
		tracer:       f.tracer,
		db:           f.db.Executor(),
		meter:        f.meter,
		interceptors: interceptors, // Inject interceptors from registry
	}
}

// FactoryOption is a function that configures a ContextFactory.
type FactoryOption func(*contextFactory)

// WithRegistry sets the interceptor registry to use for context creation.
// The registry is used to dynamically fetch interceptors at context creation time.
//
// Example with fx:
//
//	fx.Provide(func(registry InterceptorRegistry, ...) ContextFactory {
//	    return NewContextFactory(..., WithRegistry(registry))
//	})
func WithRegistry(registry InterceptorRegistry) FactoryOption {
	return func(f *contextFactory) {
		f.registry = registry
	}
}
