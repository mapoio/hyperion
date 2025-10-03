package decorators

import (
	"time"

	"github.com/mapoio/hyperion"
)

// LoggingInterceptor provides method-level logging for any service.
// It logs method entry, exit, duration, and errors.
type LoggingInterceptor struct {
	logger        hyperion.Logger
	componentName string
}

// NewLoggingInterceptor creates a new logging interceptor.
//
// Example:
//
//	interceptor := NewLoggingInterceptor(logger, "UserRepository")
func NewLoggingInterceptor(logger hyperion.Logger, componentName string) *LoggingInterceptor {
	return &LoggingInterceptor{
		logger:        logger,
		componentName: componentName,
	}
}

// Log0 wraps a method with no return values except error.
// Use this for methods like: MethodName(ctx Context) error
//
// Example:
//
//	logger := NewLoggingInterceptor(l, "UserRepository")
//	func (p *proxy) Delete(ctx hyperion.Context) error {
//	    return Log0(logger, "Delete", p.target.Delete)(ctx)
//	}
func Log0(
	interceptor *LoggingInterceptor,
	methodName string,
	fn func(hyperion.Context) error,
) func(hyperion.Context) error {
	return func(ctx hyperion.Context) error {
		start := time.Now()
		logger := ctx.Logger().With("component", interceptor.componentName, "method", methodName)

		logger.Debug("method called")

		err := fn(ctx)

		duration := time.Since(start)
		if err != nil {
			logger.Error("method failed",
				"duration", duration.String(),
				"error", err.Error(),
			)
		} else {
			logger.Debug("method completed",
				"duration", duration.String(),
			)
		}

		return err
	}
}

// Log1 wraps a method with 1 argument, 1 result value and error.
// Use this for methods like: MethodName(ctx Context, arg1) (R1, error)
//
// Example:
//
//	logger := NewLoggingInterceptor(l, "UserRepository")
//	func (p *proxy) GetUser(ctx hyperion.Context, id int64) (*User, error) {
//	    return Log1(logger, "GetUser", p.target.GetUser)(ctx, id)
//	}
func Log1[A1, R1 any](
	interceptor *LoggingInterceptor,
	methodName string,
	fn func(hyperion.Context, A1) (R1, error),
) func(hyperion.Context, A1) (R1, error) {
	return func(ctx hyperion.Context, a1 A1) (R1, error) {
		start := time.Now()
		logger := ctx.Logger().With("component", interceptor.componentName, "method", methodName)

		logger.Debug("method called")

		result, err := fn(ctx, a1)

		duration := time.Since(start)
		if err != nil {
			logger.Error("method failed",
				"duration", duration.String(),
				"error", err.Error(),
			)
		} else {
			logger.Debug("method completed",
				"duration", duration.String(),
			)
		}

		return result, err
	}
}

// Log2 wraps a method with 2 arguments, 1 result value and error.
// Use this for methods like: MethodName(ctx Context, arg1, arg2) (R1, error)
//
// Example:
//
//	logger := NewLoggingInterceptor(l, "UserRepository")
//	func (p *proxy) UpdateUser(ctx hyperion.Context, id int64, user *User) (*User, error) {
//	    return Log2(logger, "UpdateUser", p.target.UpdateUser)(ctx, id, user)
//	}
func Log2[A1, A2, R1 any](
	interceptor *LoggingInterceptor,
	methodName string,
	fn func(hyperion.Context, A1, A2) (R1, error),
) func(hyperion.Context, A1, A2) (R1, error) {
	return func(ctx hyperion.Context, a1 A1, a2 A2) (R1, error) {
		start := time.Now()
		logger := ctx.Logger().With("component", interceptor.componentName, "method", methodName)

		logger.Debug("method called")

		result, err := fn(ctx, a1, a2)

		duration := time.Since(start)
		if err != nil {
			logger.Error("method failed",
				"duration", duration.String(),
				"error", err.Error(),
			)
		} else {
			logger.Debug("method completed",
				"duration", duration.String(),
			)
		}

		return result, err
	}
}

// Log3 wraps a method with 3 arguments, 1 result value and error.
//
// Example:
//
//	logger := NewLoggingInterceptor(l, "UserRepository")
//	func (p *proxy) ComplexMethod(ctx hyperion.Context, a1 T1, a2 T2, a3 T3) (*Result, error) {
//	    return Log3(logger, "ComplexMethod", p.target.ComplexMethod)(ctx, a1, a2, a3)
//	}
func Log3[A1, A2, A3, R1 any](
	interceptor *LoggingInterceptor,
	methodName string,
	fn func(hyperion.Context, A1, A2, A3) (R1, error),
) func(hyperion.Context, A1, A2, A3) (R1, error) {
	return func(ctx hyperion.Context, a1 A1, a2 A2, a3 A3) (R1, error) {
		start := time.Now()
		logger := ctx.Logger().With("component", interceptor.componentName, "method", methodName)

		logger.Debug("method called")

		result, err := fn(ctx, a1, a2, a3)

		duration := time.Since(start)
		if err != nil {
			logger.Error("method failed",
				"duration", duration.String(),
				"error", err.Error(),
			)
		} else {
			logger.Debug("method completed",
				"duration", duration.String(),
			)
		}

		return result, err
	}
}

// Log1NoErr wraps a method with 1 argument and 1 result, no error.
// Use this for methods like: MethodName(ctx Context, arg1) R1
//
// Example:
//
//	logger := NewLoggingInterceptor(l, "UserRepository")
//	func (p *proxy) Count(ctx hyperion.Context, filter string) int64 {
//	    return Log1NoErr(logger, "Count", p.target.Count)(ctx, filter)
//	}
func Log1NoErr[A1, R1 any](
	interceptor *LoggingInterceptor,
	methodName string,
	fn func(hyperion.Context, A1) R1,
) func(hyperion.Context, A1) R1 {
	return func(ctx hyperion.Context, a1 A1) R1 {
		start := time.Now()
		logger := ctx.Logger().With("component", interceptor.componentName, "method", methodName)

		logger.Debug("method called")

		result := fn(ctx, a1)

		duration := time.Since(start)
		logger.Debug("method completed",
			"duration", duration.String(),
		)

		return result
	}
}

// Log2Results wraps a method with 1 argument, 2 result values and error.
// Use this for methods like: MethodName(ctx Context, arg1) (R1, R2, error)
//
// Example:
//
//	logger := NewLoggingInterceptor(l, "UserRepository")
//	func (p *proxy) GetUserWithStats(ctx hyperion.Context, id int64) (*User, *Stats, error) {
//	    return Log2Results(logger, "GetUserWithStats", p.target.GetUserWithStats)(ctx, id)
//	}
func Log2Results[A1, R1, R2 any](
	interceptor *LoggingInterceptor,
	methodName string,
	fn func(hyperion.Context, A1) (R1, R2, error),
) func(hyperion.Context, A1) (R1, R2, error) {
	return func(ctx hyperion.Context, a1 A1) (R1, R2, error) {
		start := time.Now()
		logger := ctx.Logger().With("component", interceptor.componentName, "method", methodName)

		logger.Debug("method called")

		r1, r2, err := fn(ctx, a1)

		duration := time.Since(start)
		if err != nil {
			logger.Error("method failed",
				"duration", duration.String(),
				"error", err.Error(),
			)
		} else {
			logger.Debug("method completed",
				"duration", duration.String(),
			)
		}

		return r1, r2, err
	}
}

// AutoLogging returns a decorator that should be applied via fx.Decorate.
//
// This function does NOT use reflection. Users must manually implement
// the proxy type that wraps their target interface.
//
// Recommended Pattern: Embedding + Selective Override (same as tracing)
//
//	// Step 1: Define proxy with embedded interface
//	type userRepositoryProxy struct {
//	    UserRepository  // Embed original interface
//	    logger *LoggingInterceptor
//	}
//
//	func NewUserRepositoryWithLogging(base UserRepository, logger hyperion.Logger) UserRepository {
//	    return &userRepositoryProxy{
//	        UserRepository: base,
//	        logger: NewLoggingInterceptor(logger, "UserRepository"),
//	    }
//	}
//
//	// Step 2: Override only methods you want to log (one-liner each)
//	func (p *userRepositoryProxy) GetUser(ctx hyperion.Context, id int64) (*User, error) {
//	    return Log1(p.logger, "GetUser", p.UserRepository.GetUser)(ctx, id)
//	}
//
//	func (p *userRepositoryProxy) UpdateUser(ctx hyperion.Context, id int64, user *User) (*User, error) {
//	    return Log2(p.logger, "UpdateUser", p.UserRepository.UpdateUser)(ctx, id, user)
//	}
//
//	func (p *userRepositoryProxy) DeleteUser(ctx hyperion.Context, id int64) error {
//	    return Log1(p.logger, "DeleteUser", p.UserRepository.DeleteUser)(ctx, id)
//	}
//
//	// Step 3: Use in fx module
//	var Module = fx.Module("repository",
//	    fx.Provide(NewUserRepositoryImpl),
//	    fx.Decorate(NewUserRepositoryWithLogging),
//	)
//
// You can combine multiple decorators using fx.Decorate:
//
//	var Module = fx.Module("repository",
//	    fx.Provide(NewUserRepositoryImpl),
//	    fx.Decorate(NewUserRepositoryWithLogging),   // Add logging
//	    fx.Decorate(NewUserRepositoryWithTracing),   // Add tracing on top
//	)
func AutoLogging[T any](logger hyperion.Logger, componentName string) func(T) T {
	panic("AutoLogging requires manual proxy implementation. " +
		"See decorators.NewLoggingInterceptor for helper methods.")
}
