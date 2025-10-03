package hyperion

// Interceptor defines the interface for method interceptors.
//
// Interceptors are chained together and executed in order determined by Order().
// They can modify the context, record metrics, trace calls, log events, etc.
//
// Example implementation:
//
//	type TracingInterceptor struct {
//	    tracer Tracer
//	}
//
//	func (ti *TracingInterceptor) Name() string {
//	    return "tracing"
//	}
//
//	func (ti *TracingInterceptor) Intercept(ctx Context, fullPath string) (Context, func(err *error), error) {
//	    newCtx, span := ti.tracer.Start(ctx, fullPath)
//	    end := func(errPtr *error) {
//	        if errPtr != nil && *errPtr != nil {
//	            span.RecordError(*errPtr)
//	        }
//	        span.End()
//	    }
//	    return newCtx, end, nil
//	}
//
//	func (ti *TracingInterceptor) Order() int {
//	    return 100
//	}
type Interceptor interface {
	// Name returns the unique name of this interceptor.
	// Used for filtering interceptors via InterceptConfig options.
	//
	// Examples: "tracing", "logging", "metrics", "retry"
	Name() string

	// Intercept intercepts a method call.
	//
	// Parameters:
	//   - ctx: The current hyperion.Context
	//   - fullPath: Full method path (e.g., "Service.User.GetUser" or "UserRepository.GetUser")
	//
	// Returns:
	//   - newCtx: Updated context (can be the same as input if no modification needed)
	//   - end: Function to be called when method execution completes (via defer)
	//          The end function receives a pointer to the error for inspection
	//   - ierr: Error during interceptor setup (rare, logs warning and continues with other interceptors)
	Intercept(
		ctx Context,
		fullPath string,
	) (newCtx Context, end func(err *error), ierr error)

	// Order returns the execution order of this interceptor.
	// Interceptors are executed from lowest to highest order value.
	//
	// Recommended values:
	//   - 0-99:   Infrastructure (panic recovery, rate limiting)
	//   - 100-199: Tracing (OpenTelemetry spans)
	//   - 200-299: Logging (structured logs)
	//   - 300-399: Metrics (Prometheus, StatsD)
	//   - 400-499: Custom business logic
	//   - 500+:    Post-processing
	Order() int
}
