package otel

import (
	"context"
	"testing"
	"time"

	"github.com/mapoio/hyperion"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestNewOtelTracer(t *testing.T) {
	tests := []struct {
		name    string
		config  TracingConfig
		wantErr bool
	}{
		{
			name: "valid otlp config",
			config: TracingConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "otlp",
				Endpoint:    "localhost:4317",
				SampleRate:  1.0,
			},
			wantErr: false,
		},
		{
			name: "disabled tracing",
			config: TracingConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing service name",
			config: TracingConfig{
				Enabled:    true,
				Exporter:   "otlp",
				Endpoint:   "localhost:4317",
				SampleRate: 1.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockConfig{
				data: map[string]any{
					"tracing": tt.config,
				},
			}

			tracer, err := NewOtelTracer(mock)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOtelTracer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.config.Enabled {
				if tracer == nil {
					t.Fatal("expected non-nil tracer")
				}

				// Test basic functionality
				ctx := context.Background()
				_, span := tracer.Start(wrapContext(ctx), "test-span")
				if span == nil {
					t.Error("expected non-nil span")
				}
				span.End()

				// Shutdown via type assertion
				if ot, ok := tracer.(*OtelTracer); ok {
					if err := ot.Shutdown(ctx); err != nil {
						t.Errorf("failed to shutdown tracer: %v", err)
					}
				}
			}
		})
	}
}

func TestNewOtelMeter(t *testing.T) {
	// Reset provider before tests to ensure independence
	resetProviderForTesting()

	tests := []struct {
		name    string
		config  MetricsConfig
		wantErr bool
	}{
		{
			name: "valid prometheus config",
			config: MetricsConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "prometheus",
				Interval:    10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "otlp config",
			config: MetricsConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "otlp",
				Endpoint:    "localhost:4317",
				Interval:    10 * time.Second,
			},
			wantErr: false, // OTLP metrics now implemented
		},
		{
			name: "disabled metrics",
			config: MetricsConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing service name",
			config: MetricsConfig{
				Enabled:  true,
				Exporter: "prometheus",
				Interval: 10 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset provider for each sub-test to ensure independence
			resetProviderForTesting()

			mock := &mockConfig{
				data: map[string]any{
					"metrics": tt.config,
				},
			}

			meter, err := NewOtelMeter(mock)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOtelMeter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.config.Enabled {
				if meter == nil {
					t.Fatal("expected non-nil meter")
				}

				// Test basic functionality
				counter := meter.Counter("test-counter")
				if counter == nil {
					t.Error("expected non-nil counter")
				}
			}
		})
	}
}

func TestRegisterShutdownHook(t *testing.T) {
	t.Run("registers shutdown hook successfully", func(t *testing.T) {
		resetProviderForTesting()

		app := fxtest.New(t,
			fx.Provide(func() hyperion.Config {
				return &mockConfig{
					data: map[string]any{
						"tracing": TracingConfig{
							Enabled:     true,
							ServiceName: "test-service",
							Exporter:    "otlp",
							Endpoint:    "localhost:4317",
							SampleRate:  1.0,
						},
					},
				}
			}),
			fx.Provide(func(cfg hyperion.Config) (*sdktrace.TracerProvider, error) {
				tracer, err := NewOtelTracer(cfg)
				if err != nil {
					return nil, err
				}
				return tracer.(*OtelTracer).provider.(*sdktrace.TracerProvider), nil
			}),
			TracerModule,
			fx.Invoke(RegisterShutdownHook),
		)

		app.RequireStart()
		app.RequireStop()
	})
}

func TestTracerModule(t *testing.T) {
	t.Run("tracer module provides tracer", func(t *testing.T) {
		resetProviderForTesting()

		var tracer hyperion.Tracer

		app := fxtest.New(t,
			fx.Provide(func() hyperion.Config {
				return &mockConfig{
					data: map[string]any{
						"tracing": TracingConfig{
							Enabled:     true,
							ServiceName: "test-service",
							Exporter:    "otlp",
							Endpoint:    "localhost:4317",
							SampleRate:  1.0,
						},
					},
				}
			}),
			fx.Provide(func(cfg hyperion.Config) (*sdktrace.TracerProvider, error) {
				tracer, err := NewOtelTracer(cfg)
				if err != nil {
					return nil, err
				}
				return tracer.(*OtelTracer).provider.(*sdktrace.TracerProvider), nil
			}),
			TracerModule,
			fx.Populate(&tracer),
		)

		app.RequireStart()
		defer app.RequireStop()

		if tracer == nil {
			t.Fatal("expected tracer to be populated")
		}

		// Test basic functionality
		ctx := context.Background()
		_, span := tracer.Start(wrapContext(ctx), "test-span")
		if span == nil {
			t.Error("expected non-nil span")
		}
		span.End()
	})
}

func TestMeterModule(t *testing.T) {
	t.Run("meter module provides meter", func(t *testing.T) {
		resetProviderForTesting()

		var meter hyperion.Meter

		app := fxtest.New(t,
			fx.Provide(func() hyperion.Config {
				return &mockConfig{
					data: map[string]any{
						"metrics": MetricsConfig{
							Enabled:     true,
							ServiceName: "test-service",
							Exporter:    "prometheus",
							Interval:    10 * time.Second,
						},
					},
				}
			}),
			fx.Provide(func(cfg hyperion.Config) (*sdkmetric.MeterProvider, error) {
				meter, err := NewOtelMeter(cfg)
				if err != nil {
					return nil, err
				}
				return meter.(*OtelMeter).provider.(*sdkmetric.MeterProvider), nil
			}),
			MeterModule,
			fx.Populate(&meter),
		)

		app.RequireStart()
		defer app.RequireStop()

		if meter == nil {
			t.Fatal("expected meter to be populated")
		}

		// Test basic functionality
		counter := meter.Counter("test-counter")
		if counter == nil {
			t.Error("expected non-nil counter")
		}
	})
}
