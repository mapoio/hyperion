package telemetry

import (
	"net/http"

	"github.com/mapoio/hyperion"
	hyperotel "github.com/mapoio/hyperion/adapter/otel"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/fx"
)

// NewInstrumentedHTTPClient creates an HTTP client with automatic tracing.
// All outgoing HTTP requests will be automatically traced.
func NewInstrumentedHTTPClient(tracer hyperion.Tracer) *http.Client {
	// Type-assert to get access to the underlying TracerProvider
	otelTracer, ok := tracer.(*hyperotel.OtelTracer)
	if !ok {
		// Fallback to default client if not using OTel tracer
		return http.DefaultClient
	}

	return &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithTracerProvider(otelTracer.TracerProvider()),
		),
	}
}

// HTTPInstrumentationModule provides automatic HTTP tracing.
var HTTPInstrumentationModule = fx.Module("http_instrumentation",
	fx.Provide(NewInstrumentedHTTPClient),
)
