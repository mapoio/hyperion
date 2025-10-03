package hyperion

import "strings"

// InterceptConfig configures which interceptors to apply for a specific method call.
type InterceptConfig struct {
	// Only applies these interceptors (by name).
	// If empty, all registered interceptors are applied.
	Only []string

	// Exclude these interceptors (by name).
	// Takes precedence over Only.
	Exclude []string

	// Additional interceptors to apply for this specific call.
	// These are prepended to the global interceptor chain.
	Additional []Interceptor
}

// InterceptOption is a function that modifies InterceptConfig.
type InterceptOption func(*InterceptConfig)

// WithOnly creates an option to only apply specific interceptors.
//
// Example:
//
//	ctx, end := ctx.UseIntercept("UserService", "GetUser",
//	    hyperion.WithOnly("tracing", "logging"))
//	defer end(&err)
func WithOnly(names ...string) InterceptOption {
	return func(cfg *InterceptConfig) {
		cfg.Only = names
	}
}

// WithExclude creates an option to exclude specific interceptors.
//
// Example:
//
//	// Exclude logging for high-frequency calls
//	ctx, end := ctx.UseIntercept("UserService", "GetUser",
//	    hyperion.WithExclude("logging"))
//	defer end(&err)
func WithExclude(names ...string) InterceptOption {
	return func(cfg *InterceptConfig) {
		cfg.Exclude = names
	}
}

// WithAdditional adds custom interceptors for this specific call.
//
// Example:
//
//	metricsInterceptor := NewCustomMetrics("user.get")
//	ctx, end := ctx.UseIntercept("UserService", "GetUser",
//	    hyperion.WithAdditional(metricsInterceptor))
//	defer end(&err)
func WithAdditional(interceptors ...Interceptor) InterceptOption {
	return func(cfg *InterceptConfig) {
		cfg.Additional = append(cfg.Additional, interceptors...)
	}
}

// shouldApply checks if an interceptor should be applied based on config.
func (cfg *InterceptConfig) shouldApply(name string) bool {
	// Exclude takes precedence
	for _, exclude := range cfg.Exclude {
		if exclude == name {
			return false
		}
	}

	// If Only is specified, check if name is in the list
	if len(cfg.Only) > 0 {
		for _, only := range cfg.Only {
			if only == name {
				return true
			}
		}
		return false
	}

	// By default, apply all interceptors
	return true
}

// JoinPath joins path segments with "." separator.
// Handles string and InterceptOption types.
//
// Example:
//
//	JoinPath("Service", "User", "GetUser") => "Service.User.GetUser"
//	JoinPath("UserService", "GetUser", WithOnly("tracing")) => "UserService.GetUser", [WithOnly("tracing")]
func JoinPath(parts ...any) (path string, opts []InterceptOption) {
	segments := make([]string, 0, len(parts))
	options := make([]InterceptOption, 0)

	for _, part := range parts {
		switch v := part.(type) {
		case string:
			if v != "" {
				segments = append(segments, v)
			}
		case InterceptOption:
			options = append(options, v)
		}
	}

	return strings.Join(segments, "."), options
}
