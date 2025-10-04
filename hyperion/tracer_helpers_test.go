package hyperion

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartSpan_BasicUsage(t *testing.T) {
	// Create a capture tracer to verify calls
	tracer := &captureTracer{}
	logger := NewNoOpLogger()

	// Create hyperion.Context with interceptors
	ctx := New(context.Background(), logger, nil, tracer, nil)

	// Note: To enable interceptors, we'd need to set them on the context
	// For this test, we'll just verify the basic API works
	var err error
	newCtx, end := StartSpan(ctx, "test-operation")
	defer end(&err)

	// Verify context was returned
	assert.NotNil(t, newCtx)

	// Verify it's still a hyperion.Context
	assert.NotNil(t, newCtx.Logger())
	assert.NotNil(t, newCtx.Tracer())
}

type testContextKey string

const testKey testContextKey = "test_key"

func TestStartSpan_PreservesContext(t *testing.T) {
	tracer := NewNoOpTracer()
	logger := NewNoOpLogger()

	ctx := New(context.Background(), logger, nil, tracer, nil)

	// Add a value to the context
	stdCtx := context.WithValue(ctx, testKey, "test_value")
	hctx := WithContext(ctx, stdCtx)

	// Call StartSpan
	var err error
	newCtx, end := StartSpan(hctx, "test-operation")
	defer end(&err)

	// Verify the value is still accessible
	value := newCtx.Value(testKey)
	assert.Equal(t, "test_value", value)
}

func TestStartSpan_WithTracingInterceptor(t *testing.T) {
	// Create a capture tracer to verify span creation
	tracer := &captureTracer{}
	logger := NewNoOpLogger()

	// Create hyperion.Context
	baseCtx := New(context.Background(), logger, nil, tracer, nil)

	// Create a TracingInterceptor
	interceptor := NewTracingInterceptor(tracer)

	// Manually set interceptors (simulating what fx would do)
	hctx := baseCtx.(*hyperionContext)
	hctx.interceptors = []Interceptor{interceptor}

	// Call StartSpan
	var err error
	newCtx, end := StartSpan(hctx, "test-operation")
	defer end(&err)

	// Verify interceptor was called
	assert.NotNil(t, newCtx)
	assert.Len(t, tracer.startCalls, 1)
	assert.Equal(t, "test-operation", tracer.startCalls[0].spanName)
}
