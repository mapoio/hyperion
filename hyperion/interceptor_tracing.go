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
	// Start a new span
	newStdCtx, span := ti.tracer.Start(ctx, fullPath)

	// Create a new hyperion.Context with the updated standard context
	// We need to preserve all other fields (logger, db, tracer, meter, interceptors)
	hctx, ok := ctx.(*hyperionContext)
	if !ok {
		// Fallback: create new context
		newCtx := New(newStdCtx, ctx.Logger(), ctx.DB(), ctx.Tracer(), ctx.Meter())
		end := func(errPtr *error) {
			if errPtr != nil && *errPtr != nil {
				span.RecordError(*errPtr)
			}
			span.End()
		}
		return newCtx, end, nil
	}

	// Create new context with updated standard context
	newHyperionCtx := &hyperionContext{
		Context:      newStdCtx,
		logger:       hctx.logger,
		tracer:       hctx.tracer,
		db:           hctx.db,
		meter:        hctx.meter,
		interceptors: hctx.interceptors,
	}

	// Create end function that records errors and ends the span
	end := func(errPtr *error) {
		if errPtr != nil && *errPtr != nil {
			span.RecordError(*errPtr)
		}
		span.End()
	}

	return newHyperionCtx, end, nil
}

// Order implements Interceptor.Order.
// Tracing should be the outer-most interceptor (lowest order value).
func (ti *TracingInterceptor) Order() int {
	return 100
}
