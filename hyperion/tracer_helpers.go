package hyperion

// StartSpan is a convenience helper for creating spans with interceptors.
//
// This function applies all registered interceptors (tracing, logging, etc.)
// and returns the updated context. It's particularly useful for simple cases
// where you want automatic span creation without the full UseIntercept pattern.
//
// Recommended usage with named return for automatic error recording:
//
//	func (s *Service) Method(ctx hyperion.Context, arg string) (result Type, err error) {
//	    ctx = hyperion.StartSpan(ctx, "Service.Method")
//	    defer func() {
//	        if err != nil {
//	            // Error is automatically recorded by interceptor
//	        }
//	    }()
//	    // Business logic...
//	    return result, nil
//	}
//
// For full control with error recording, prefer UseIntercept:
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
func StartSpan(ctx Context, spanName string) Context {
	newCtx, _ := ctx.UseIntercept(spanName)
	return newCtx
}
