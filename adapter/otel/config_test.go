package otel

import (
	"testing"
	"time"
)

type mockConfig struct {
	data map[string]any
}

func (m *mockConfig) Unmarshal(key string, rawVal any) error {
	if val, ok := m.data[key]; ok {
		// Simple type assertion for testing
		switch v := rawVal.(type) {
		case *TracingConfig:
			if cfg, ok := val.(TracingConfig); ok {
				*v = cfg
			}
		case *MetricsConfig:
			if cfg, ok := val.(MetricsConfig); ok {
				*v = cfg
			}
		}
	}
	return nil
}

func (m *mockConfig) Get(key string) any {
	return m.data[key]
}

func (m *mockConfig) GetString(key string) string {
	if val, ok := m.data[key].(string); ok {
		return val
	}
	return ""
}

func (m *mockConfig) GetInt(key string) int {
	if val, ok := m.data[key].(int); ok {
		return val
	}
	return 0
}

func (m *mockConfig) GetInt64(key string) int64 {
	if val, ok := m.data[key].(int64); ok {
		return val
	}
	return 0
}

func (m *mockConfig) GetBool(key string) bool {
	if val, ok := m.data[key].(bool); ok {
		return val
	}
	return false
}

func (m *mockConfig) GetFloat64(key string) float64 {
	if val, ok := m.data[key].(float64); ok {
		return val
	}
	return 0
}

func (m *mockConfig) GetStringSlice(key string) []string {
	if val, ok := m.data[key].([]string); ok {
		return val
	}
	return nil
}

func (m *mockConfig) IsSet(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *mockConfig) AllKeys() []string {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func TestLoadTracingConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  TracingConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: TracingConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "jaeger",
				Endpoint:    "localhost:14268",
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
			name: "otlp exporter",
			config: TracingConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "otlp",
				Endpoint:    "localhost:4317",
				SampleRate:  0.5,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockConfig{
				data: map[string]any{
					"tracing": tt.config,
				},
			}

			cfg, err := LoadTracingConfig(mock)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTracingConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && cfg.Enabled {
				if cfg.ServiceName != tt.config.ServiceName {
					t.Errorf("ServiceName = %v, want %v", cfg.ServiceName, tt.config.ServiceName)
				}
				if cfg.Exporter != tt.config.Exporter {
					t.Errorf("Exporter = %v, want %v", cfg.Exporter, tt.config.Exporter)
				}
			}
		})
	}
}

func TestValidateTracingConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  TracingConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: TracingConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "jaeger",
				Endpoint:    "localhost:14268",
				SampleRate:  1.0,
			},
			wantErr: false,
		},
		{
			name: "missing service name",
			config: TracingConfig{
				Enabled:    true,
				Exporter:   "jaeger",
				Endpoint:   "localhost:14268",
				SampleRate: 1.0,
			},
			wantErr: true,
		},
		{
			name: "invalid exporter",
			config: TracingConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "invalid",
				Endpoint:    "localhost:14268",
				SampleRate:  1.0,
			},
			wantErr: true,
		},
		{
			name: "missing endpoint",
			config: TracingConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "jaeger",
				SampleRate:  1.0,
			},
			wantErr: true,
		},
		{
			name: "invalid sample rate - too high",
			config: TracingConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "jaeger",
				Endpoint:    "localhost:14268",
				SampleRate:  1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid sample rate - negative",
			config: TracingConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "jaeger",
				Endpoint:    "localhost:14268",
				SampleRate:  -0.1,
			},
			wantErr: true,
		},
		{
			name: "disabled tracing skips validation",
			config: TracingConfig{
				Enabled: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTracingConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTracingConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadMetricsConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  MetricsConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: MetricsConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "prometheus",
				Interval:    10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "disabled metrics",
			config: MetricsConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "otlp exporter",
			config: MetricsConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "otlp",
				Endpoint:    "localhost:4317",
				Interval:    5 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockConfig{
				data: map[string]any{
					"metrics": tt.config,
				},
			}

			cfg, err := LoadMetricsConfig(mock)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadMetricsConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && cfg.Enabled {
				if cfg.ServiceName != tt.config.ServiceName {
					t.Errorf("ServiceName = %v, want %v", cfg.ServiceName, tt.config.ServiceName)
				}
				if cfg.Exporter != tt.config.Exporter {
					t.Errorf("Exporter = %v, want %v", cfg.Exporter, tt.config.Exporter)
				}
			}
		})
	}
}

func TestValidateMetricsConfig(t *testing.T) {
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
			name: "valid otlp config",
			config: MetricsConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "otlp",
				Endpoint:    "localhost:4317",
				Interval:    10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing service name",
			config: MetricsConfig{
				Enabled:  true,
				Exporter: "prometheus",
				Interval: 10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid exporter",
			config: MetricsConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "invalid",
				Interval:    10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "otlp without endpoint",
			config: MetricsConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "otlp",
				Interval:    10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid interval",
			config: MetricsConfig{
				Enabled:     true,
				ServiceName: "test-service",
				Exporter:    "prometheus",
				Interval:    0,
			},
			wantErr: true,
		},
		{
			name: "disabled metrics skips validation",
			config: MetricsConfig{
				Enabled: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMetricsConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMetricsConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
