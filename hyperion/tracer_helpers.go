package hyperion

// StartSpan is a convenience helper for creating spans with interceptors.
//
// This function applies all registered interceptors (tracing, logging, etc.)
// and returns both the updated context and an end function that must be called
// to properly finish the span.
//
// Recommended usage with named return for automatic error recording:
//
//	func (s *Service) Method(ctx hyperion.Context, arg string) (result Type, err error) {
//	    ctx, end := hyperion.StartSpan(ctx, "Service.Method")
//	    defer end(&err)  // Automatically records error on span
//	    // Business logic...
//	    return result, nil
//	}
//
// Alternative: Direct use of UseIntercept for component-based naming:
//
//	func (s *Service) Method(ctx hyperion.Context, arg string) (result Type, err error) {
//	    ctx, end := ctx.UseIntercept("Service", "Method")
//	    defer end(&err)  // Automatically records error on span
//	    // Business logic...
//	    return result, nil
//	}
//
// Parameters:
//   - ctx: The current hyperion.Context
//   - spanName: Name of the span (typically "Service.Method")
//
// Returns:
//   - Updated context with all interceptors applied
//   - End function that must be called to finish the span
func StartSpan(ctx Context, spanName string) (Context, func(*error)) {
	return ctx.UseIntercept(spanName)
}
