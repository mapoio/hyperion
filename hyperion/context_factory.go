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
//	}
type ContextFactory interface {
	// New creates a new Hyperion context from a standard context.
	// The returned context will have Logger, Tracer, and DB injected.
	New(ctx context.Context) Context
}

// contextFactory is the default implementation of ContextFactory.
type contextFactory struct {
	logger Logger
	tracer Tracer
	db     Database

	// Decorators to apply when creating contexts
	loggerDecorators   []LoggerDecorator
	tracerDecorators   []TracerDecorator
	executorDecorators []ExecutorDecorator
}

// NewContextFactory creates a new ContextFactory with the given dependencies.
// Dependencies are typically injected via fx.Provide.
//
// Example:
//
//	var ContextModule = fx.Module("hyperion.context",
//	    fx.Provide(NewContextFactory),
//	)
func NewContextFactory(logger Logger, tracer Tracer, db Database, opts ...FactoryOption) ContextFactory {
	f := &contextFactory{
		logger: logger,
		tracer: tracer,
		db:     db,
	}

	// Apply options (decorators)
	for _, opt := range opts {
		opt(f)
	}

	return f
}

// New creates a new Hyperion context with injected dependencies.
func (f *contextFactory) New(ctx context.Context) Context {
	// Apply logger decorators
	logger := f.logger
	for _, decorator := range f.loggerDecorators {
		logger = decorator(logger)
	}

	// Apply tracer decorators
	tracer := f.tracer
	for _, decorator := range f.tracerDecorators {
		tracer = decorator(tracer)
	}

	// Apply executor decorators to Database.Executor()
	executor := f.db.Executor()
	for _, decorator := range f.executorDecorators {
		executor = decorator(executor)
	}

	return &hyperionContext{
		Context: ctx,
		logger:  logger,
		tracer:  tracer,
		db:      executor,
	}
}

// FactoryOption is a function that configures a ContextFactory.
type FactoryOption func(*contextFactory)

// WithLoggerDecorator adds logger decorators to the factory.
// Decorators are applied in the order they are added.
//
// Example:
//
//	factory := NewContextFactory(logger, tracer, db,
//	    WithLoggerDecorator(AddPrefixDecorator("[APP]")),
//	)
func WithLoggerDecorator(decorators ...LoggerDecorator) FactoryOption {
	return func(f *contextFactory) {
		f.loggerDecorators = append(f.loggerDecorators, decorators...)
	}
}

// WithTracerDecorator adds tracer decorators to the factory.
// Decorators are applied in the order they are added.
func WithTracerDecorator(decorators ...TracerDecorator) FactoryOption {
	return func(f *contextFactory) {
		f.tracerDecorators = append(f.tracerDecorators, decorators...)
	}
}

// WithExecutorDecorator adds executor decorators to the factory.
// Decorators are applied in the order they are added.
//
// Example:
//
//	factory := NewContextFactory(logger, tracer, db,
//	    WithExecutorDecorator(QueryLoggingDecorator(logger)),
//	)
func WithExecutorDecorator(decorators ...ExecutorDecorator) FactoryOption {
	return func(f *contextFactory) {
		f.executorDecorators = append(f.executorDecorators, decorators...)
	}
}
