package hyperion

import "time"

// LoggingInterceptor provides structured logging for method calls.
// It logs method start, completion, duration, and errors.
type LoggingInterceptor struct {
	logger Logger
}

// NewLoggingInterceptor creates a new logging interceptor.
func NewLoggingInterceptor(logger Logger) *LoggingInterceptor {
	return &LoggingInterceptor{logger: logger}
}

// Name implements Interceptor.Name.
func (li *LoggingInterceptor) Name() string {
	return "logging"
}

// Intercept implements Interceptor.Intercept.
// It logs method execution with timing and error information.
func (li *LoggingInterceptor) Intercept(
	ctx Context,
	fullPath string,
) (Context, func(err *error), error) {
	start := time.Now()

	li.logger.Debug("Method started", "path", fullPath)

	end := func(errPtr *error) {
		duration := time.Since(start)

		if errPtr != nil && *errPtr != nil {
			li.logger.Error("Method failed",
				"path", fullPath,
				"duration", duration,
				"error", *errPtr,
			)
		} else {
			li.logger.Debug("Method completed",
				"path", fullPath,
				"duration", duration,
			)
		}
	}

	// Logging doesn't modify the context, just observes execution
	return ctx, end, nil
}

// Order implements Interceptor.Order.
// Logging should run after tracing but before metrics.
func (li *LoggingInterceptor) Order() int {
	return 200
}
