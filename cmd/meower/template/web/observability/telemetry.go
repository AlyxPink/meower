// Package observability wires the web server to the shared observability kit.
// The web server exports traces over OTLP/HTTP (port 4318), whereas the api
// server uses OTLP/gRPC (port 4317) — both land in the same collector.
package observability

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	sharedobs "TEMPLATE_MODULE_PATH/pkg/observability"
)

// TelemetryConfig holds configuration for OpenTelemetry setup.
type TelemetryConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string // HTTP endpoint, e.g. "tempo:4318"
}

// DefaultConfig returns sensible defaults for the web server. The OTLP endpoint
// comes from OTEL_EXPORTER_OTLP_ENDPOINT and defaults to localhost:4318.
func DefaultConfig() TelemetryConfig {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4318"
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "dev"
	}

	return TelemetryConfig{
		ServiceName:    sharedobs.ServiceNameWeb,
		ServiceVersion: version,
		Environment:    env,
		OTLPEndpoint:   endpoint,
	}
}

// Telemetry holds the OpenTelemetry providers for graceful shutdown.
type Telemetry struct {
	tracerProvider *sdktrace.TracerProvider
}

// InitTelemetry initializes OpenTelemetry with a trace provider exporting over
// OTLP/HTTP. Call Shutdown on server exit to flush buffered spans.
func InitTelemetry(ctx context.Context, cfg TelemetryConfig) (*Telemetry, error) {
	// Install the deduplicating error handler before the trace provider starts so
	// a down collector (e.g. the monitoring stack isn't running) logs a single
	// warning instead of a "connection refused" line on every batch tick.
	sharedobs.SetQuietErrorHandler()

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	traceExporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(
			traceExporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Telemetry{tracerProvider: tracerProvider}, nil
}

// Shutdown gracefully shuts down the telemetry providers.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t.tracerProvider != nil {
		return t.tracerProvider.Shutdown(ctx)
	}
	return nil
}
