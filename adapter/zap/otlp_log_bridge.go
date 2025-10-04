//nolint:gocyclo,errcheck // OTLP bridge is optional, errors handled gracefully
package zap

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

// otlpLogBridge bridges Zap logs to OpenTelemetry logs.
// It implements zapcore.WriteSyncer to capture log entries and forward them to OTLP.
type otlpLogBridge struct {
	provider  *sdklog.LoggerProvider
	processor *jsonLogProcessor
}

// newOtlpLogBridge creates a new OTLP log bridge.
func newOtlpLogBridge(provider *sdklog.LoggerProvider, serviceName string) *otlpLogBridge {
	processor := &jsonLogProcessor{
		serviceName: serviceName,
	}
	return &otlpLogBridge{
		provider:  provider,
		processor: processor,
	}
}

// jsonLogProcessor is a custom processor that parses JSON logs and emits them.
type jsonLogProcessor struct {
	serviceName string
}

// Write implements zapcore.WriteSyncer.
// It parses the JSON log entry and emits it as an OpenTelemetry log record.
func (b *otlpLogBridge) Write(p []byte) (n int, err error) {
	// Parse JSON log entry
	var entry map[string]interface{}
	if err := json.Unmarshal(p, &entry); err != nil {
		// If parsing fails, just return success to avoid breaking the logger
		return len(p), nil
	}

	// Extract standard fields
	msg, _ := entry["msg"].(string)
	levelStr, _ := entry["level"].(string)
	tsStr, _ := entry["ts"].(string)
	caller, _ := entry["caller"].(string)
	traceIDStr, _ := entry["trace_id"].(string)
	spanIDStr, _ := entry["span_id"].(string)

	// Parse timestamp
	var timestamp time.Time
	if tsStr != "" {
		if parsed, err := time.Parse(time.RFC3339, tsStr); err == nil {
			timestamp = parsed
		}
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	// Map Zap level to OTel severity
	severity := mapLevelToSeverity(levelStr)

	// Create a proper SDK log record using the provider
	// We need to emit through the provider's logger to get proper processing
	logger := b.provider.Logger(b.processor.serviceName)

	// Create API log record
	var record log.Record
	record.SetTimestamp(timestamp)
	record.SetBody(log.StringValue(msg))
	record.SetSeverity(severity)
	record.SetSeverityText(levelStr)

	// Add attributes from log entry
	attrs := make([]log.KeyValue, 0, len(entry))
	for k, v := range entry {
		// Skip standard fields that are already set
		if k == "msg" || k == "level" || k == "ts" || k == "trace_id" || k == "span_id" {
			continue
		}

		// Add caller as attribute
		if k == "caller" {
			attrs = append(attrs, log.String("caller", caller))
			continue
		}

		// Convert other fields to attributes
		attrs = append(attrs, convertToKeyValue(k, v))
	}

	record.AddAttributes(attrs...)

	// Create context with trace information if available
	ctx := context.Background()
	if traceIDStr != "" && spanIDStr != "" {
		if traceID, err := parseTraceID(traceIDStr); err == nil {
			if spanID, err := parseSpanID(spanIDStr); err == nil {
				// Create a span context and inject it into the context
				spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
					TraceID:    traceID,
					SpanID:     spanID,
					TraceFlags: trace.FlagsSampled,
				})
				ctx = trace.ContextWithSpanContext(ctx, spanCtx)
			}
		}
	}

	// Emit the log record with trace context
	// The SDK will automatically extract trace info from context and set it on the SDK Record
	logger.Emit(ctx, record)

	return len(p), nil
}

// Sync implements zapcore.WriteSyncer.
func (b *otlpLogBridge) Sync() error {
	if b.provider != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// Use ForceFlush instead of Shutdown to ensure logs are exported
		// without closing the provider. Shutdown should only be called during app termination.
		return b.provider.ForceFlush(ctx)
	}
	return nil
}

// mapLevelToSeverity maps Zap log level to OpenTelemetry severity.
func mapLevelToSeverity(level string) log.Severity {
	switch level {
	case "debug":
		return log.SeverityDebug
	case "info":
		return log.SeverityInfo
	case "warn":
		return log.SeverityWarn
	case "error":
		return log.SeverityError
	case "fatal":
		return log.SeverityFatal
	default:
		return log.SeverityInfo
	}
}

// convertToKeyValue converts a Go value to an OpenTelemetry KeyValue.
func convertToKeyValue(key string, value interface{}) log.KeyValue {
	switch v := value.(type) {
	case string:
		return log.String(key, v)
	case int:
		return log.Int64(key, int64(v))
	case int64:
		return log.Int64(key, v)
	case float64:
		return log.Float64(key, v)
	case bool:
		return log.Bool(key, v)
	default:
		// For complex types, convert to string
		return log.String(key, toString(v))
	}
}

// toString converts a value to string representation.
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// parseTraceID parses a hex string into a trace.TraceID.
func parseTraceID(s string) (trace.TraceID, error) {
	var traceID trace.TraceID
	_, err := hex.Decode(traceID[:], []byte(s))
	return traceID, err
}

// parseSpanID parses a hex string into a trace.SpanID.
func parseSpanID(s string) (trace.SpanID, error) {
	var spanID trace.SpanID
	_, err := hex.Decode(spanID[:], []byte(s))
	return spanID, err
}
