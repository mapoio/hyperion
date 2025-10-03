package otel

import (
	"fmt"
	"time"

	"github.com/mapoio/hyperion"
)

// TracingConfig defines the configuration for OpenTelemetry tracing.
type TracingConfig struct {
	// Enabled indicates whether tracing is enabled.
	Enabled bool `mapstructure:"enabled"`

	// ServiceName is the name of the service for tracing.
	ServiceName string `mapstructure:"service_name"`

	// Exporter specifies the trace exporter type ("jaeger" or "otlp").
	Exporter string `mapstructure:"exporter"`

	// Endpoint is the exporter endpoint (e.g., "localhost:14268" for Jaeger).
	Endpoint string `mapstructure:"endpoint"`

	// SampleRate controls trace sampling (0.0 - 1.0, where 1.0 = 100%).
	SampleRate float64 `mapstructure:"sample_rate"`

	// Attributes are global span attributes applied to all traces.
	Attributes map[string]string `mapstructure:"attributes"`
}

// MetricsConfig defines the configuration for OpenTelemetry metrics.
type MetricsConfig struct {
	// Enabled indicates whether metrics collection is enabled.
	Enabled bool `mapstructure:"enabled"`

	// ServiceName is the name of the service for metrics.
	ServiceName string `mapstructure:"service_name"`

	// Exporter specifies the metrics exporter type ("prometheus" or "otlp").
	Exporter string `mapstructure:"exporter"`

	// Endpoint is the exporter endpoint (only used for OTLP).
	Endpoint string `mapstructure:"endpoint"`

	// Interval is the metrics collection interval.
	Interval time.Duration `mapstructure:"interval"`

	// Attributes are global metric attributes applied to all metrics.
	Attributes map[string]string `mapstructure:"attributes"`
}

// LoadTracingConfig loads tracing configuration from the provided config source.
func LoadTracingConfig(config hyperion.Config) (TracingConfig, error) {
	var cfg TracingConfig

	// Set defaults
	cfg.Enabled = true
	cfg.SampleRate = 1.0
	cfg.Exporter = "jaeger"

	// Load from config
	if err := config.Unmarshal("tracing", &cfg); err != nil {
		return TracingConfig{}, fmt.Errorf("failed to unmarshal tracing config: %w", err)
	}

	// Validate
	if err := validateTracingConfig(cfg); err != nil {
		return TracingConfig{}, err
	}

	return cfg, nil
}

// LoadMetricsConfig loads metrics configuration from the provided config source.
func LoadMetricsConfig(config hyperion.Config) (MetricsConfig, error) {
	var cfg MetricsConfig

	// Set defaults
	cfg.Enabled = true
	cfg.Interval = 10 * time.Second
	cfg.Exporter = "prometheus"

	// Load from config
	if err := config.Unmarshal("metrics", &cfg); err != nil {
		return MetricsConfig{}, fmt.Errorf("failed to unmarshal metrics config: %w", err)
	}

	// Validate
	if err := validateMetricsConfig(cfg); err != nil {
		return MetricsConfig{}, err
	}

	return cfg, nil
}

// validateTracingConfig validates the tracing configuration.
func validateTracingConfig(cfg TracingConfig) error {
	if !cfg.Enabled {
		return nil // No validation needed if disabled
	}

	if cfg.ServiceName == "" {
		return fmt.Errorf("tracing.service_name is required when tracing is enabled")
	}

	if cfg.Exporter != "jaeger" && cfg.Exporter != "otlp" {
		return fmt.Errorf("tracing.exporter must be 'jaeger' or 'otlp', got %q", cfg.Exporter)
	}

	if cfg.Endpoint == "" {
		return fmt.Errorf("tracing.endpoint is required when tracing is enabled")
	}

	if cfg.SampleRate < 0.0 || cfg.SampleRate > 1.0 {
		return fmt.Errorf("tracing.sample_rate must be between 0.0 and 1.0, got %f", cfg.SampleRate)
	}

	return nil
}

// validateMetricsConfig validates the metrics configuration.
func validateMetricsConfig(cfg MetricsConfig) error {
	if !cfg.Enabled {
		return nil // No validation needed if disabled
	}

	if cfg.ServiceName == "" {
		return fmt.Errorf("metrics.service_name is required when metrics are enabled")
	}

	if cfg.Exporter != "prometheus" && cfg.Exporter != "otlp" {
		return fmt.Errorf("metrics.exporter must be 'prometheus' or 'otlp', got %q", cfg.Exporter)
	}

	if cfg.Exporter == "otlp" && cfg.Endpoint == "" {
		return fmt.Errorf("metrics.endpoint is required when using OTLP exporter")
	}

	if cfg.Interval <= 0 {
		return fmt.Errorf("metrics.interval must be positive, got %v", cfg.Interval)
	}

	return nil
}
