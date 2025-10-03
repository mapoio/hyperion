package hyperion

import (
	"context"
	"errors"
	"testing"
)

func TestInterceptorOrdering(t *testing.T) {
	interceptors := []Interceptor{
		&mockInterceptor{name: "c", order: 300},
		&mockInterceptor{name: "a", order: 100},
		&mockInterceptor{name: "b", order: 200},
	}

	sorted := sortInterceptors(interceptors)

	if len(sorted) != 3 {
		t.Fatalf("expected 3 interceptors, got %d", len(sorted))
	}

	if sorted[0].Name() != "a" {
		t.Errorf("expected first interceptor to be 'a', got %s", sorted[0].Name())
	}
	if sorted[1].Name() != "b" {
		t.Errorf("expected second interceptor to be 'b', got %s", sorted[1].Name())
	}
	if sorted[2].Name() != "c" {
		t.Errorf("expected third interceptor to be 'c', got %s", sorted[2].Name())
	}
}

func TestInterceptConfig_shouldApply(t *testing.T) {
	tests := []struct {
		name     string
		config   InterceptConfig
		intName  string
		expected bool
	}{
		{
			name:     "no restrictions - should apply",
			config:   InterceptConfig{},
			intName:  "tracing",
			expected: true,
		},
		{
			name: "only specified - should apply",
			config: InterceptConfig{
				Only: []string{"tracing", "logging"},
			},
			intName:  "tracing",
			expected: true,
		},
		{
			name: "only specified - should not apply",
			config: InterceptConfig{
				Only: []string{"tracing"},
			},
			intName:  "logging",
			expected: false,
		},
		{
			name: "exclude specified - should not apply",
			config: InterceptConfig{
				Exclude: []string{"logging"},
			},
			intName:  "logging",
			expected: false,
		},
		{
			name: "exclude takes precedence over only",
			config: InterceptConfig{
				Only:    []string{"tracing", "logging"},
				Exclude: []string{"logging"},
			},
			intName:  "logging",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.shouldApply(tt.intName)
			if result != tt.expected {
				t.Errorf("shouldApply(%s) = %v, want %v", tt.intName, result, tt.expected)
			}
		})
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		name         string
		parts        []any
		expectedPath string
		expectedOpts int
	}{
		{
			name:         "simple path",
			parts:        []any{"UserService", "GetUser"},
			expectedPath: "UserService.GetUser",
			expectedOpts: 0,
		},
		{
			name:         "namespaced path",
			parts:        []any{"Service", "User", "GetUser"},
			expectedPath: "Service.User.GetUser",
			expectedOpts: 0,
		},
		{
			name:         "path with options",
			parts:        []any{"UserService", "GetUser", WithOnly("tracing")},
			expectedPath: "UserService.GetUser",
			expectedOpts: 1,
		},
		{
			name:         "path with multiple options",
			parts:        []any{"UserService", "GetUser", WithOnly("tracing"), WithExclude("logging")},
			expectedPath: "UserService.GetUser",
			expectedOpts: 2,
		},
		{
			name:         "empty strings filtered",
			parts:        []any{"", "UserService", "", "GetUser", ""},
			expectedPath: "UserService.GetUser",
			expectedOpts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, opts := JoinPath(tt.parts...)
			if path != tt.expectedPath {
				t.Errorf("JoinPath() path = %v, want %v", path, tt.expectedPath)
			}
			if len(opts) != tt.expectedOpts {
				t.Errorf("JoinPath() options count = %v, want %v", len(opts), tt.expectedOpts)
			}
		})
	}
}

func TestUseIntercept_NoInterceptors(t *testing.T) {
	logger := &noopLogger{}
	tracer := &noopTracer{}
	db := &noopExecutor{}

	ctx := &hyperionContext{
		Context:      context.Background(),
		logger:       logger,
		tracer:       tracer,
		db:           db,
		interceptors: []Interceptor{}, // No interceptors
	}

	newCtx, end := ctx.UseIntercept("Test", "Method")

	if newCtx != ctx {
		t.Error("expected same context when no interceptors")
	}

	// Should not panic
	var err error
	end(&err)
}

func TestUseIntercept_WithInterceptors(t *testing.T) {
	logger := &noopLogger{}
	tracer := &noopTracer{}
	db := &noopExecutor{}

	var executionOrder []string
	var endOrder []string

	ctx := &hyperionContext{
		Context: context.Background(),
		logger:  logger,
		tracer:  tracer,
		db:      db,
		interceptors: []Interceptor{
			&mockInterceptor{
				name:  "first",
				order: 100,
				onIntercept: func() {
					executionOrder = append(executionOrder, "first")
				},
				onEnd: func(_ *error) {
					endOrder = append(endOrder, "first")
				},
			},
			&mockInterceptor{
				name:  "second",
				order: 200,
				onIntercept: func() {
					executionOrder = append(executionOrder, "second")
				},
				onEnd: func(_ *error) {
					endOrder = append(endOrder, "second")
				},
			},
		},
	}

	newCtx, end := ctx.UseIntercept("Test", "Method")

	if newCtx == nil {
		t.Fatal("expected non-nil context")
	}

	// Check execution order
	if len(executionOrder) != 2 {
		t.Fatalf("expected 2 interceptors to execute, got %d", len(executionOrder))
	}
	if executionOrder[0] != "first" || executionOrder[1] != "second" {
		t.Errorf("wrong execution order: %v", executionOrder)
	}

	// Call end functions
	var err error
	end(&err)

	// Check end order (LIFO)
	if len(endOrder) != 2 {
		t.Fatalf("expected 2 end functions to execute, got %d", len(endOrder))
	}
	if endOrder[0] != "second" || endOrder[1] != "first" {
		t.Errorf("wrong end order (expected LIFO): %v", endOrder)
	}
}

func TestUseIntercept_WithError(t *testing.T) {
	logger := &noopLogger{}
	tracer := &noopTracer{}
	db := &noopExecutor{}

	var recordedError error

	ctx := &hyperionContext{
		Context: context.Background(),
		logger:  logger,
		tracer:  tracer,
		db:      db,
		interceptors: []Interceptor{
			&mockInterceptor{
				name:  "observer",
				order: 100,
				onEnd: func(err *error) {
					if err != nil && *err != nil {
						recordedError = *err
					}
				},
			},
		},
	}

	_, end := ctx.UseIntercept("Test", "Method")

	testErr := errors.New("test error")
	end(&testErr)

	if recordedError == nil {
		t.Error("expected interceptor to receive error")
	}
	if recordedError.Error() != "test error" {
		t.Errorf("wrong error recorded: %v", recordedError)
	}
}

func TestUseIntercept_SelectiveApplication(t *testing.T) {
	logger := &noopLogger{}
	tracer := &noopTracer{}
	db := &noopExecutor{}

	var executed []string

	ctx := &hyperionContext{
		Context: context.Background(),
		logger:  logger,
		tracer:  tracer,
		db:      db,
		interceptors: []Interceptor{
			&mockInterceptor{
				name:  "tracing",
				order: 100,
				onIntercept: func() {
					executed = append(executed, "tracing")
				},
			},
			&mockInterceptor{
				name:  "logging",
				order: 200,
				onIntercept: func() {
					executed = append(executed, "logging")
				},
			},
			&mockInterceptor{
				name:  "metrics",
				order: 300,
				onIntercept: func() {
					executed = append(executed, "metrics")
				},
			},
		},
	}

	// Test WithOnly
	executed = nil
	_, end := ctx.UseIntercept("Test", "Method", WithOnly("tracing"))
	end(nil)

	if len(executed) != 1 || executed[0] != "tracing" {
		t.Errorf("WithOnly failed: executed = %v", executed)
	}

	// Test WithExclude
	executed = nil
	_, end = ctx.UseIntercept("Test", "Method", WithExclude("logging"))
	end(nil)

	if len(executed) != 2 {
		t.Errorf("WithExclude failed: expected 2 interceptors, got %v", executed)
	}
	for _, name := range executed {
		if name == "logging" {
			t.Error("WithExclude failed: logging should be excluded")
		}
	}
}

// Mock implementations for testing

type mockInterceptor struct {
	name        string
	order       int
	onIntercept func()
	onEnd       func(err *error)
}

func (m *mockInterceptor) Name() string {
	return m.name
}

func (m *mockInterceptor) Intercept(
	ctx Context,
	fullPath string,
) (Context, func(err *error), error) {
	if m.onIntercept != nil {
		m.onIntercept()
	}

	end := func(errPtr *error) {
		if m.onEnd != nil {
			m.onEnd(errPtr)
		}
	}

	return ctx, end, nil
}

func (m *mockInterceptor) Order() int {
	return m.order
}
