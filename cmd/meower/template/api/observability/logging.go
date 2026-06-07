// Package observability wires the api server to the shared observability kit:
// OpenTelemetry tracing (telemetry.go), a traced database pool (dbtracer.go),
// a Prometheus metrics endpoint (metrics.go), and logging configuration. The
// cross-component primitives live in pkg/observability; this package adds the
// api-server-specific setup on top.
package observability

import (
	sharedobs "TEMPLATE_MODULE_PATH/pkg/observability"
)

// ConfigureLogging sets up logging for the API server based on environment.
func ConfigureLogging(environment string, level string) {
	sharedobs.ConfigureLogging(environment, level)
}
