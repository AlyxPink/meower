// Package observability provides shared logging and tracing helpers used across
// every component (api, web, and any future service). Keeping it in one
// sub-module means all components emit logs and traces in the same shape, so
// they correlate cleanly in Loki/Tempo.
package observability

import (
	"context"

	"github.com/charmbracelet/log"
)

// LogWithTrace returns a logger with trace context attached for log
// correlation. When a trace is active, it adds trace_id and span_id fields so a
// log line can be jumped to from its trace (and vice versa) in Grafana.
func LogWithTrace(ctx context.Context) *log.Logger {
	traceID := TraceIDFromContext(ctx)
	spanID := SpanIDFromContext(ctx)

	if traceID != "" {
		return log.With("trace_id", traceID, "span_id", spanID)
	}
	return log.Default()
}

// ConfigureLogging sets up structured logging based on environment and level.
// In production it switches to JSON output so Loki can parse fields; in
// development it keeps the human-readable formatter.
func ConfigureLogging(environment string, level string) {
	switch level {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	log.SetReportTimestamp(true)
	log.SetReportCaller(false)

	// Use JSON format in production for Loki parsing.
	if environment == "production" {
		log.SetFormatter(log.JSONFormatter)
	}
}
