package hyperion

// ContextMiddleware wraps a context-using function to add cross-cutting behavior.
// This enables AOP-style service layer interceptors for logging, tracing, transactions, etc.
//
// The middleware receives the context and a "next" function to call.
// It can execute code before and after calling next, modify the context, or handle errors.
//
// Example: Logging middleware
//
//	func LoggingMiddleware(ctx Context, next func(Context) error) error {
//	    ctx.Logger().Info("operation started")
//	    err := next(ctx)
//	    if err != nil {
//	        ctx.Logger().Error("operation failed", "error", err)
//	    }
//	    return err
//	}
type ContextMiddleware func(ctx Context, next func(Context) error) error

// ChainMiddleware composes multiple ContextMiddleware into a single middleware.
// Middleware are executed in the order they are provided (left to right).
//
// Example:
//
//	middleware := ChainMiddleware(
//	    LoggingMiddleware,
//	    TracingMiddleware,
//	    TransactionMiddleware,
//	)
//
//	err := middleware(ctx, func(ctx Context) error {
//	    return service.DoWork(ctx)
//	})
//
// Execution order:
//  1. LoggingMiddleware (before)
//  2. TracingMiddleware (before)
//  3. TransactionMiddleware (before)
//  4. service.DoWork(ctx)
//  5. TransactionMiddleware (after)
//  6. TracingMiddleware (after)
//  7. LoggingMiddleware (after)
func ChainMiddleware(middlewares ...ContextMiddleware) ContextMiddleware {
	return func(ctx Context, final func(Context) error) error {
		// Build middleware chain from right to left
		handler := final
		for i := len(middlewares) - 1; i >= 0; i-- {
			// Capture current middleware in closure
			mw := middlewares[i]
			next := handler
			handler = func(c Context) error {
				return mw(c, next)
			}
		}
		return handler(ctx)
	}
}

// ApplyMiddleware wraps a service method with middleware.
// This is a convenience function for single middleware application.
//
// Example:
//
//	func (s *Service) CreateUser(ctx Context, req CreateUserRequest) error {
//	    return ApplyMiddleware(LoggingMiddleware, ctx, func(ctx Context) error {
//	        return s.createUserImpl(ctx, req)
//	    })
//	}
func ApplyMiddleware(middleware ContextMiddleware, ctx Context, fn func(Context) error) error {
	return middleware(ctx, fn)
}

// MiddlewareFunc is an adapter to allow ordinary functions to be used as middleware.
// The function receives the context before and after the next handler.
//
// Example:
//
//	beforeAfter := MiddlewareFunc(func(ctx Context, next func(Context) error) error {
//	    ctx.Logger().Debug("before")
//	    defer ctx.Logger().Debug("after")
//	    return next(ctx)
//	})
type MiddlewareFunc func(Context, func(Context) error) error

// Middleware returns itself to satisfy the ContextMiddleware interface.
func (f MiddlewareFunc) Middleware() ContextMiddleware {
	return ContextMiddleware(f)
}
