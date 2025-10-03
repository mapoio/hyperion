package otel

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/exporters/prometheus"
)

func TestCreateTraceExporter(t *testing.T) {
	tests := []struct {
		name    string
		cfg     TracingConfig
		wantErr bool
	}{
		{
			name: "otlp exporter",
			cfg: TracingConfig{
				Exporter: "otlp",
				Endpoint: "localhost:4317",
			},
			wantErr: false,
		},
		{
			name: "unsupported exporter",
			cfg: TracingConfig{
				Exporter: "zipkin",
				Endpoint: "localhost:9411",
			},
			wantErr: true,
		},
		{
			name: "jaeger exporter (via OTLP)",
			cfg: TracingConfig{
				Exporter: "jaeger",
				Endpoint: "localhost:14268",
			},
			wantErr: false, // Jaeger uses OTLP internally
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter, err := createTraceExporter(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("createTraceExporter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if exporter == nil {
					t.Error("expected non-nil exporter")
				}

				// Clean up - use context.Background() instead of nil
				ctx := context.Background()
				if err := exporter.Shutdown(ctx); err != nil {
					t.Logf("failed to shutdown exporter: %v", err)
				}
			}
		})
	}
}

func TestCreateMetricsReader(t *testing.T) {
	tests := []struct {
		name     string
		cfg      MetricsConfig
		wantType string
		wantErr  bool
	}{
		{
			name: "prometheus reader",
			cfg: MetricsConfig{
				Exporter: "prometheus",
			},
			wantType: "prometheus",
			wantErr:  false,
		},
		{
			name: "otlp reader (not implemented)",
			cfg: MetricsConfig{
				Exporter: "otlp",
				Endpoint: "localhost:4317",
			},
			wantType: "otlp",
			wantErr:  true, // OTLP metrics not yet implemented
		},
		{
			name: "unsupported exporter",
			cfg: MetricsConfig{
				Exporter: "statsd",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := createMetricsReader(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("createMetricsReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if reader == nil {
					t.Error("expected non-nil reader")
				}

				// Type assertions to verify correct reader type
				if tt.wantType == "prometheus" {
					if _, ok := reader.(*prometheus.Exporter); !ok {
						t.Errorf("expected prometheus.Exporter, got %T", reader)
					}
				}

				// Clean up - readers don't have Shutdown method, they're closed via MeterProvider
			}
		})
	}
}
