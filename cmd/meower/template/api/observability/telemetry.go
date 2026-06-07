package observability

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sharedobs "TEMPLATE_MODULE_PATH/pkg/observability"
)

// TelemetryConfig holds configuration for OpenTelemetry setup.
type TelemetryConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string // e.g., "tempo:4317"
}

// DefaultConfig returns a configuration with sensible defaults for the API
// server. The OTLP endpoint is read from OTEL_EXPORTER_OTLP_ENDPOINT and falls
// back to localhost:4317, so a fresh project traces to the local monitoring
// stack with no code changes (and exports nowhere harmful if it isn't running).
func DefaultConfig() TelemetryConfig {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4317"
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
		ServiceName:    sharedobs.ServiceNameAPI,
		ServiceVersion: version,
		Environment:    env,
		OTLPEndpoint:   endpoint,
	}
}

// Telemetry holds the OpenTelemetry providers for graceful shutdown.
type Telemetry struct {
	tracerProvider *sdktrace.TracerProvider
}

// InitTelemetry initializes OpenTelemetry with a trace provider that exports
// spans over OTLP/gRPC. Returns a Telemetry instance whose Shutdown must be
// called on server exit to flush buffered spans.
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

	// Create OTLP trace exporter over an insecure gRPC connection (the
	// collector is reached over the local docker network, not the internet).
	conn, err := grpc.NewClient(
		cfg.OTLPEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	// Batch span processor for export efficiency.
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(
			traceExporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Sample everything; tune for production volume.
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Telemetry{tracerProvider: tracerProvider}, nil
}

// Shutdown gracefully shuts down the telemetry providers, flushing any buffered
// spans.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t.tracerProvider != nil {
		return t.tracerProvider.Shutdown(ctx)
	}
	return nil
}
