package zap

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// TestOtlpLogBridge_Write tests the Write method of otlpLogBridge.
func TestOtlpLogBridge_Write(t *testing.T) {
	// Create a no-op exporter for testing
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create OTLP exporter (will fail to connect, but that's OK for this test)
	exporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint("localhost:14317"), // Non-existent endpoint
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		t.Skipf("Failed to create OTLP exporter: %v (this is OK in CI)", err)
	}

	res, _ := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName("test-service")),
	)

	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)
	defer func() {
		_ = provider.Shutdown(ctx)
	}()

	bridge := newOtlpLogBridge(provider, "test-service")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid JSON log",
			input:   `{"level":"info","ts":"2024-01-01T00:00:00Z","msg":"test message","key":"value"}`,
			wantErr: false,
		},
		{
			name:    "valid JSON with trace context",
			input:   `{"level":"info","ts":"2024-01-01T00:00:00Z","msg":"test","trace_id":"4ab3828f4f2bf47f24fe3b23b5df8d71","span_id":"1e823dd17fcdec4c"}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON (should not error)",
			input:   `not json at all`,
			wantErr: false, // Should silently succeed
		},
		{
			name:    "empty input",
			input:   ``,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := bridge.Write([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if n != len(tt.input) {
				t.Errorf("Write() returned %d bytes, want %d", n, len(tt.input))
			}
		})
	}
}

// TestOtlpLogBridge_Sync tests the Sync method.
func TestOtlpLogBridge_Sync(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint("localhost:14317"),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		t.Skipf("Failed to create OTLP exporter: %v", err)
	}

	res, _ := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName("test-service")),
	)

	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)
	defer func() {
		_ = provider.Shutdown(ctx)
	}()

	bridge := newOtlpLogBridge(provider, "test-service")

	// Sync should not error
	if err := bridge.Sync(); err != nil {
		t.Errorf("Sync() error = %v", err)
	}
}

// TestMapLevelToSeverity tests severity mapping.
func TestMapLevelToSeverity(t *testing.T) {
	tests := []struct {
		level string
		want  int // We can't import the exact type, so we use int
	}{
		{"debug", 5},   // log.SeverityDebug
		{"info", 9},    // log.SeverityInfo
		{"warn", 13},   // log.SeverityWarn
		{"error", 17},  // log.SeverityError
		{"fatal", 21},  // log.SeverityFatal
		{"unknown", 9}, // log.SeverityInfo (default)
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			got := mapLevelToSeverity(tt.level)
			if int(got) != tt.want {
				t.Errorf("mapLevelToSeverity(%s) = %d, want %d", tt.level, int(got), tt.want)
			}
		})
	}
}

// TestParseTraceID tests trace ID parsing.
func TestParseTraceID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid trace ID",
			input:   "4ab3828f4f2bf47f24fe3b23b5df8d71",
			wantErr: false,
		},
		{
			name:    "invalid trace ID (too short)",
			input:   "123",
			wantErr: true,
		},
		{
			name:    "invalid trace ID (not hex)",
			input:   "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			wantErr: true,
		},
		{
			name:    "empty trace ID",
			input:   "",
			wantErr: false, // hex.Decode on empty string returns 0 bytes decoded, no error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseTraceID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTraceID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestParseSpanID tests span ID parsing.
func TestParseSpanID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid span ID",
			input:   "1e823dd17fcdec4c",
			wantErr: false,
		},
		{
			name:    "invalid span ID (too short)",
			input:   "123",
			wantErr: true,
		},
		{
			name:    "invalid span ID (not hex)",
			input:   "zzzzzzzzzzzzzzzz",
			wantErr: true,
		},
		{
			name:    "empty span ID",
			input:   "",
			wantErr: false, // hex.Decode on empty string returns 0 bytes decoded, no error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseSpanID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSpanID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCreateOtlpLogBridge tests the createOtlpLogBridge function.
func TestCreateOtlpLogBridge(t *testing.T) {
	tests := []struct {
		name    string
		config  *OtlpLogConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "valid config",
			config: &OtlpLogConfig{
				Enabled:     true,
				Endpoint:    "localhost:14317",
				ServiceName: "test-service",
			},
			wantErr: false,
		},
		{
			name: "empty service name (should use default)",
			config: &OtlpLogConfig{
				Enabled:     true,
				Endpoint:    "localhost:14317",
				ServiceName: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge, err := createOtlpLogBridge(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("createOtlpLogBridge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && bridge == nil {
				t.Error("createOtlpLogBridge() returned nil bridge")
			}
		})
	}
}

// TestToString tests the toString conversion function.
func TestToString(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "string",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "int",
			input: 42,
			want:  "42",
		},
		{
			name:  "float",
			input: 3.14,
			want:  "3.14",
		},
		{
			name:  "bool",
			input: true,
			want:  "true",
		},
		{
			name:  "nil",
			input: nil,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toString(tt.input)
			if got != tt.want {
				t.Errorf("toString() = %v, want %v", got, tt.want)
			}
		})
	}
}
