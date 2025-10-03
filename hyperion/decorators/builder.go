package decorators

import (
	"github.com/mapoio/hyperion"
)

// ProxyBuilder was considered but the embedding pattern is simpler and more idiomatic.
//
// RECOMMENDED APPROACH: Use interface embedding + selective method override
//
// This is the most convenient way to add tracing to your services:
//
//	// Step 1: Define your interface
//	type UserRepository interface {
//	    GetUser(ctx hyperion.Context, id int64) (*User, error)
//	    UpdateUser(ctx hyperion.Context, id int64, user *User) (*User, error)
//	    DeleteUser(ctx hyperion.Context, id int64) error
//	    ListUsers(ctx hyperion.Context, filter string) ([]*User, error)
//	}
//
//	// Step 2: Implement the base repository
//	type userRepositoryImpl struct {
//	    db hyperion.Executor
//	}
//
//	func NewUserRepository(db hyperion.Executor) UserRepository {
//	    return &userRepositoryImpl{db: db}
//	}
//
//	// Step 3: Create tracing proxy with embedding (one-time setup)
//	type userRepositoryProxy struct {
//	    UserRepository  // Embed interface - all methods pass through by default
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
//	// Step 4: Override methods you want to trace (one-liner each)
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
//	func (p *userRepositoryProxy) ListUsers(ctx hyperion.Context, filter string) ([]*User, error) {
//	    return Trace1(p.tracer, "ListUsers", p.UserRepository.ListUsers)(ctx, filter)
//	}
//
//	// Step 5: Wire with fx
//	var Module = fx.Module("repository",
//	    fx.Provide(NewUserRepository),
//	    fx.Decorate(NewUserRepositoryWithTracing),
//	)
//
// Why this pattern is optimal:
//
//  1. Zero boilerplate: Just embed the interface
//  2. One-line per method: Each traced method is a single TraceN call
//  3. Selective tracing: Only override methods you want to trace
//  4. Type-safe: Compiler catches signature mismatches
//  5. IDE-friendly: Full autocomplete and navigation
//  6. Performance: Zero reflection at runtime
//
// Available Trace helpers:
//
//   - Trace0(name, fn)              for: (ctx) error
//   - Trace1(name, fn)              for: (ctx, arg1) (result, error)
//   - Trace2(name, fn)              for: (ctx, arg1, arg2) (result, error)
//   - Trace3(name, fn)              for: (ctx, arg1, arg2, arg3) (result, error)
//   - Trace1NoErr(name, fn)         for: (ctx, arg1) result
//   - Trace2Results(name, fn)       for: (ctx, arg1) (result1, result2, error)
//   - WrapMethod(name, fn)          for: custom wrapping (advanced)
//   - WrapMethodWithResult(name, fn) for: custom wrapping (advanced)
//
// ProxyBuilder is not needed with this pattern.
type ProxyBuilder[T any] struct {
	componentName string
	tracer        hyperion.Tracer
	interceptor   *TracingInterceptor
}

// NewProxy creates a new proxy builder.
func NewProxy[T any](componentName string, tracer hyperion.Tracer) *ProxyBuilder[T] {
	return &ProxyBuilder[T]{
		componentName: componentName,
		tracer:        tracer,
		interceptor:   NewTracingInterceptor(tracer, componentName),
	}
}

// Interceptor returns the underlying interceptor for manual wrapping.
func (b *ProxyBuilder[T]) Interceptor() *TracingInterceptor {
	return b.interceptor
}
