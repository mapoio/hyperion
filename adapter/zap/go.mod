module github.com/mapoio/hyperion/adapter/zap

go 1.24

require (
	github.com/mapoio/hyperion v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel/trace v1.38.0
	go.uber.org/fx v1.24.0
	go.uber.org/zap v1.27.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
)

replace github.com/mapoio/hyperion => ../../hyperion
