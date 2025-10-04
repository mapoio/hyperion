package hyperion

import "context"

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
	New(ctx context.Context) Context
}

// contextFactory is the default implementation of ContextFactory.
type contextFactory struct {
	logger Logger
	tracer Tracer
	db     Database
	meter  Meter

	// Interceptors to inject into contexts (from fx group)
	interceptors []Interceptor
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

	// DEBUG: Log final interceptor count after options applied
	logger.Info("🔍 [DEBUG] NewContextFactory created",
		"final_interceptors", len(f.interceptors),
	)

	return f
}

// New creates a new Hyperion context with injected dependencies.
func (f *contextFactory) New(ctx context.Context) Context {
	return &hyperionContext{
		Context:      ctx,
		logger:       f.logger,
		tracer:       f.tracer,
		db:           f.db.Executor(),
		meter:        f.meter,
		interceptors: f.interceptors, // Inject interceptors from fx group
	}
}

// FactoryOption is a function that configures a ContextFactory.
type FactoryOption func(*contextFactory)

// WithInterceptors sets the interceptors to inject into contexts.
// This is typically used with fx group injection.
//
// Example with fx:
//
//	type FactoryParams struct {
//	    fx.In
//	    Logger       Logger
//	    Tracer       Tracer
//	    DB           Database
//	    Meter        Meter
//	    Interceptors []Interceptor `group:"hyperion.interceptors"`
//	}
//
//	fx.Provide(func(p FactoryParams) ContextFactory {
//	    return NewContextFactory(p.Logger, p.Tracer, p.DB, p.Meter,
//	        WithInterceptors(p.Interceptors...))
//	})
func WithInterceptors(interceptors ...Interceptor) FactoryOption {
	return func(f *contextFactory) {
		f.interceptors = sortInterceptors(interceptors)
	}
}
