package hyperion

const tracingInterceptorName = "tracing"

// TracingInterceptor provides OpenTelemetry tracing for method calls.
// It creates a span for each intercepted method and automatically records errors.
type TracingInterceptor struct {
	tracer Tracer
}

// NewTracingInterceptor creates a new tracing interceptor.
func NewTracingInterceptor(tracer Tracer) *TracingInterceptor {
	return &TracingInterceptor{tracer: tracer}
}

// Name implements Interceptor.Name.
func (ti *TracingInterceptor) Name() string {
	return tracingInterceptorName
}

// Intercept implements Interceptor.Intercept.
// It creates an OpenTelemetry span for the method call.
func (ti *TracingInterceptor) Intercept(
	ctx Context,
	fullPath string,
) (Context, func(err *error), error) {
	// Start a new span - tracer.Start already updates the context with span context
	// and returns a properly configured hyperion.Context
	newHctx, span := ti.tracer.Start(ctx, fullPath)

	// Create end function that records errors and ends the span
	end := func(errPtr *error) {
		if errPtr != nil && *errPtr != nil {
			span.RecordError(*errPtr)
		}
		span.End()
	}

	return newHctx, end, nil
}

// Order implements Interceptor.Order.
// Tracing should be the outer-most interceptor (lowest order value).
func (ti *TracingInterceptor) Order() int {
	return 100
}
