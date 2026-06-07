package observability

import (
	"context"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	sharedobs "TEMPLATE_MODULE_PATH/pkg/observability"
)

// BusinessMetrics holds application gauges scraped by Prometheus. The single
// example below shows the pattern; add your own gauges/counters here and set
// them from StartMetricsUpdater's callback.
type BusinessMetrics struct {
	// UsersTotal is an example gauge. Replace it with metrics that matter to
	// your application (entity counts, queue depths, …).
	UsersTotal prometheus.Gauge
}

// MetricsServer exposes a Prometheus /metrics endpoint and owns the OTel meter
// provider, so otelgrpc instrumentation metrics are exported alongside the
// application gauges.
type MetricsServer struct {
	server          *http.Server
	businessMetrics *BusinessMetrics
	registry        *prometheus.Registry
}

// NewMetricsServer creates a metrics server listening on addr. It wires an OTel
// Prometheus exporter into the same registry that serves /metrics, so both
// runtime/instrumentation metrics and the application gauges appear at one
// endpoint.
func NewMetricsServer(addr string) (*MetricsServer, error) {
	registry := prometheus.NewRegistry()

	// Export OTel metrics (e.g. otelgrpc instrumentation) through this registry.
	exporter, err := otelprom.New(otelprom.WithRegisterer(registry))
	if err != nil {
		return nil, err
	}

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(sharedobs.ServiceNameAPI),
		),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)
	otel.SetMeterProvider(meterProvider)

	// Register the application gauges with the same registry.
	bm := &BusinessMetrics{
		UsersTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "TEMPLATE_PROJECT_NAME_users_total",
			Help: "Total number of users (example application gauge)",
		}),
	}
	registry.MustRegister(bm.UsersTotal)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return &MetricsServer{
		server:          &http.Server{Addr: addr, Handler: mux},
		businessMetrics: bm,
		registry:        registry,
	}, nil
}

// Start runs the metrics HTTP server in a goroutine.
func (m *MetricsServer) Start() {
	go func() {
		log.Info("Starting metrics server", "addr", m.server.Addr)
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Metrics server error", "error", err)
		}
	}()
}

// Shutdown gracefully shuts down the metrics server.
func (m *MetricsServer) Shutdown(ctx context.Context) error {
	return m.server.Shutdown(ctx)
}

// MetricsCallback updates the application gauges. It receives the live
// BusinessMetrics so it can Set() each gauge from freshly queried values.
type MetricsCallback func(ctx context.Context, bm *BusinessMetrics) error

// StartMetricsUpdater periodically invokes callback to refresh the application
// gauges (e.g. by querying current entity counts), until ctx is cancelled.
func (m *MetricsServer) StartMetricsUpdater(ctx context.Context, interval time.Duration, callback MetricsCallback) {
	go func() {
		if err := callback(ctx, m.businessMetrics); err != nil {
			log.Error("Failed to update business metrics", "error", err)
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := callback(ctx, m.businessMetrics); err != nil {
					log.Error("Failed to update business metrics", "error", err)
				}
			}
		}
	}()
}
