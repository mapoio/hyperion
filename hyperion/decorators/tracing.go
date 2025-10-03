package decorators

import (
	"fmt"

	"github.com/mapoio/hyperion"
)

// TracingInterceptor provides method-level tracing for any service.
// It wraps method calls with OpenTelemetry spans.
type TracingInterceptor struct {
	tracer        hyperion.Tracer
	componentName string
}

// NewTracingInterceptor creates a new tracing interceptor.
//
// Example:
//
//	interceptor := NewTracingInterceptor(tracer, "UserRepository")
func NewTracingInterceptor(tracer hyperion.Tracer, componentName string) *TracingInterceptor {
	return &TracingInterceptor{
		tracer:        tracer,
		componentName: componentName,
	}
}

// Trace0 wraps a method with no return values except error.
// Use this for methods like: MethodName(ctx Context) error
//
// Example:
//
//	tracer := NewTracingInterceptor(t, "UserRepository")
//	func (p *proxy) Delete(ctx hyperion.Context) error {
//	    return Trace0(tracer, "Delete", p.target.Delete)(ctx)
//	}
func Trace0(
	interceptor *TracingInterceptor,
	methodName string,
	fn func(hyperion.Context) error,
) func(hyperion.Context) error {
	return func(ctx hyperion.Context) error {
		spanName := fmt.Sprintf("%s.%s", interceptor.componentName, methodName)
		_, span := ctx.Tracer().Start(ctx, spanName)
		defer span.End()

		err := fn(ctx)
		if err != nil {
			span.RecordError(err)
		}
		return err
	}
}

// Trace1 wraps a method with 1 argument, 1 result value and error.
// Use this for methods like: MethodName(ctx Context, arg1) (R1, error)
//
// Example:
//
//	tracer := NewTracingInterceptor(t, "UserRepository")
//	func (p *proxy) GetUser(ctx hyperion.Context, id int64) (*User, error) {
//	    return Trace1(tracer, "GetUser", p.target.GetUser)(ctx, id)
//	}
func Trace1[A1, R1 any](
	interceptor *TracingInterceptor,
	methodName string,
	fn func(hyperion.Context, A1) (R1, error),
) func(hyperion.Context, A1) (R1, error) {
	return func(ctx hyperion.Context, a1 A1) (R1, error) {
		spanName := fmt.Sprintf("%s.%s", interceptor.componentName, methodName)
		_, span := ctx.Tracer().Start(ctx, spanName)
		defer span.End()

		result, err := fn(ctx, a1)
		if err != nil {
			span.RecordError(err)
		}
		return result, err
	}
}

// Trace2 wraps a method with 2 arguments, 1 result value and error.
// Use this for methods like: MethodName(ctx Context, arg1, arg2) (R1, error)
//
// Example:
//
//	tracer := NewTracingInterceptor(t, "UserRepository")
//	func (p *proxy) UpdateUser(ctx hyperion.Context, id int64, user *User) (*User, error) {
//	    return Trace2(tracer, "UpdateUser", p.target.UpdateUser)(ctx, id, user)
//	}
func Trace2[A1, A2, R1 any](
	interceptor *TracingInterceptor,
	methodName string,
	fn func(hyperion.Context, A1, A2) (R1, error),
) func(hyperion.Context, A1, A2) (R1, error) {
	return func(ctx hyperion.Context, a1 A1, a2 A2) (R1, error) {
		spanName := fmt.Sprintf("%s.%s", interceptor.componentName, methodName)
		_, span := ctx.Tracer().Start(ctx, spanName)
		defer span.End()

		result, err := fn(ctx, a1, a2)
		if err != nil {
			span.RecordError(err)
		}
		return result, err
	}
}

// Trace3 wraps a method with 3 arguments, 1 result value and error.
//
// Example:
//
//	tracer := NewTracingInterceptor(t, "UserRepository")
//	func (p *proxy) ComplexMethod(ctx hyperion.Context, a1 T1, a2 T2, a3 T3) (*Result, error) {
//	    return Trace3(tracer, "ComplexMethod", p.target.ComplexMethod)(ctx, a1, a2, a3)
//	}
func Trace3[A1, A2, A3, R1 any](
	interceptor *TracingInterceptor,
	methodName string,
	fn func(hyperion.Context, A1, A2, A3) (R1, error),
) func(hyperion.Context, A1, A2, A3) (R1, error) {
	return func(ctx hyperion.Context, a1 A1, a2 A2, a3 A3) (R1, error) {
		spanName := fmt.Sprintf("%s.%s", interceptor.componentName, methodName)
		_, span := ctx.Tracer().Start(ctx, spanName)
		defer span.End()

		result, err := fn(ctx, a1, a2, a3)
		if err != nil {
			span.RecordError(err)
		}
		return result, err
	}
}

// Trace1NoErr wraps a method with 1 argument and 1 result, no error.
// Use this for methods like: MethodName(ctx Context, arg1) R1
//
// Example:
//
//	tracer := NewTracingInterceptor(t, "UserRepository")
//	func (p *proxy) Count(ctx hyperion.Context, filter string) int64 {
//	    return Trace1NoErr(tracer, "Count", p.target.Count)(ctx, filter)
//	}
func Trace1NoErr[A1, R1 any](
	interceptor *TracingInterceptor,
	methodName string,
	fn func(hyperion.Context, A1) R1,
) func(hyperion.Context, A1) R1 {
	return func(ctx hyperion.Context, a1 A1) R1 {
		spanName := fmt.Sprintf("%s.%s", interceptor.componentName, methodName)
		_, span := ctx.Tracer().Start(ctx, spanName)
		defer span.End()

		return fn(ctx, a1)
	}
}

// Trace2Results wraps a method with 1 argument, 2 result values and error.
// Use this for methods like: MethodName(ctx Context, arg1) (R1, R2, error)
//
// Example:
//
//	tracer := NewTracingInterceptor(t, "UserRepository")
//	func (p *proxy) GetUserWithStats(ctx hyperion.Context, id int64) (*User, *Stats, error) {
//	    return Trace2Results(tracer, "GetUserWithStats", p.target.GetUserWithStats)(ctx, id)
//	}
func Trace2Results[A1, R1, R2 any](
	interceptor *TracingInterceptor,
	methodName string,
	fn func(hyperion.Context, A1) (R1, R2, error),
) func(hyperion.Context, A1) (R1, R2, error) {
	return func(ctx hyperion.Context, a1 A1) (R1, R2, error) {
		spanName := fmt.Sprintf("%s.%s", interceptor.componentName, methodName)
		_, span := ctx.Tracer().Start(ctx, spanName)
		defer span.End()

		r1, r2, err := fn(ctx, a1)
		if err != nil {
			span.RecordError(err)
		}
		return r1, r2, err
	}
}

// AutoTracing returns a decorator that should be applied via fx.Decorate.
//
// This function does NOT use reflection. Users must manually implement
// the proxy type that wraps their target interface.
//
// Recommended Pattern: Embedding + Selective Override
//
//	// Step 1: Define proxy with embedded interface (simplest approach)
//	type userRepositoryProxy struct {
//	    UserRepository  // Embed original interface - all methods pass through by default
//	    tracer *TracingInterceptor
//	}
//
//	func NewUserRepositoryWithTracing(base UserRepository, tracer hyperion.Tracer) UserRepository {
//	    return &userRepositoryProxy{
//	        UserRepository: base,
//	        tracer: NewTracingInterceptor(tracer, "UserRepository"),
//	    }
//	}
//
//	// Step 2: Override only methods you want to trace (one-liner each)
//	func (p *userRepositoryProxy) GetUser(ctx hyperion.Context, id int64) (*User, error) {
//	    return Trace1(p.tracer, "GetUser", p.UserRepository.GetUser)(ctx, id)
//	}
//
//	func (p *userRepositoryProxy) UpdateUser(ctx hyperion.Context, id int64, user *User) (*User, error) {
//	    return Trace2(p.tracer, "UpdateUser", p.UserRepository.UpdateUser)(ctx, id, user)
//	}
//
//	func (p *userRepositoryProxy) DeleteUser(ctx hyperion.Context, id int64) error {
//	    return Trace1(p.tracer, "DeleteUser", p.UserRepository.DeleteUser)(ctx, id)
//	}
//
//	// Step 3: Use in fx module
//	var Module = fx.Module("repository",
//	    fx.Provide(NewUserRepositoryImpl),
//	    fx.Decorate(NewUserRepositoryWithTracing),
//	)
//
// Alternative: If you don't want embedding, implement all methods explicitly.
//
// Future: We may provide a code generator to automate Step 2.
func AutoTracing[T any](tracer hyperion.Tracer, componentName string) func(T) T {
	panic(fmt.Sprintf(
		"AutoTracing requires manual proxy implementation for %s. "+
			"See decorators.NewTracingInterceptor for helper methods.",
		componentName,
	))
}
