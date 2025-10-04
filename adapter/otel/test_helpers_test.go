package otel

import (
	"context"

	"github.com/mapoio/hyperion"
)

// wrapContext is a helper function to wrap standard context into hyperion.Context for testing.
func wrapContext(stdCtx context.Context) hyperion.Context {
	return hyperion.New(stdCtx, hyperion.NewNoOpLogger(), nil, hyperion.NewNoOpTracer(), nil)
}
