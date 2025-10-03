package hyperion_test

import (
	"context"
	"testing"

	hyperion "github.com/mapoio/hyperion"
	"github.com/stretchr/testify/assert"
)

// mockContextAwareLogger implements both Logger and ContextAwareLogger
type mockContextAwareLogger struct {
	hyperion.Logger
	ctx context.Context
}

func (m *mockContextAwareLogger) WithContext(ctx context.Context) hyperion.Logger {
	return &mockContextAwareLogger{
		Logger: m.Logger,
		ctx:    ctx,
	}
}

// mockTracer implements Tracer interface
type mockTracer struct{}

func (m *mockTracer) Start(ctx context.Context, spanName string, opts ...hyperion.SpanOption) (context.Context, hyperion.Span) {
	// Create a new context with a value to simulate trace context
	newCtx := context.WithValue(ctx, "span_name", spanName)
	return newCtx, &mockSpan{name: spanName}
}

// mockSpan implements Span interface
type mockSpan struct {
	name string
}

func (m *mockSpan) End(opts ...hyperion.SpanEndOption) {}

func (m *mockSpan) SetAttributes(attrs ...hyperion.Attribute) {}

func (m *mockSpan) RecordError(err error, opts ...hyperion.EventOption) {}

func (m *mockSpan) AddEvent(name string, opts ...hyperion.EventOption) {}

func (m *mockSpan) SpanContext() hyperion.SpanContext {
	return &mockSpanContext{}
}

// mockSpanContext implements SpanContext interface
type mockSpanContext struct{}

func (m *mockSpanContext) TraceID() string {
	return "test-trace-id"
}

func (m *mockSpanContext) SpanID() string {
	return "test-span-id"
}

func (m *mockSpanContext) IsValid() bool {
	return true
}

func TestStartSpan_WithContextAwareLogger(t *testing.T) {
	ctx := context.Background()
	tracer := &mockTracer{}
	logger := &mockContextAwareLogger{
		Logger: hyperion.NewNoOpLogger(),
		ctx:    ctx,
	}

	// Call StartSpan
	newCtx, span, newLogger := hyperion.StartSpan(ctx, tracer, logger, "test-operation")

	// Verify span was created
	assert.NotNil(t, span)
	assert.Equal(t, "test-operation", span.(*mockSpan).name)

	// Verify context was updated
	assert.NotEqual(t, ctx, newCtx)
	assert.Equal(t, "test-operation", newCtx.Value("span_name"))

	// Verify logger was updated with new context
	assert.NotEqual(t, logger, newLogger)
	contextAware, ok := newLogger.(*mockContextAwareLogger)
	assert.True(t, ok)
	assert.Equal(t, newCtx, contextAware.ctx)
}

func TestStartSpan_WithNonContextAwareLogger(t *testing.T) {
	ctx := context.Background()
	tracer := &mockTracer{}
	logger := hyperion.NewNoOpLogger()

	// Call StartSpan
	newCtx, span, newLogger := hyperion.StartSpan(ctx, tracer, logger, "test-operation")

	// Verify span was created
	assert.NotNil(t, span)

	// Verify context was updated
	assert.NotEqual(t, ctx, newCtx)

	// Verify logger was returned unchanged (since it doesn't implement ContextAwareLogger)
	assert.Equal(t, logger, newLogger)
}

func TestStartSpan_WithAttributes(t *testing.T) {
	ctx := context.Background()
	tracer := &mockTracer{}
	logger := hyperion.NewNoOpLogger()

	// Call StartSpan with attributes
	newCtx, span, _ := hyperion.StartSpan(ctx, tracer, logger, "test-operation",
		hyperion.WithAttributes(
			hyperion.String("key1", "value1"),
			hyperion.Int("key2", 42),
		),
	)

	// Verify span was created
	assert.NotNil(t, span)

	// Verify context was updated
	assert.NotEqual(t, ctx, newCtx)
}
