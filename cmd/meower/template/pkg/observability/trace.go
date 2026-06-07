package observability

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TraceIDFromContext extracts the active trace ID for log correlation, or "" if
// there is no valid span in the context.
func TraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// SpanIDFromContext extracts the active span ID for log correlation, or "" if
// there is no valid span in the context.
func SpanIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// SpanFromContext returns the current span from the context.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// SetSpanAttr attaches a string attribute to the current span. It is safe to
// call when the span is not recording or when the value is empty (both are
// no-ops). Use this to enrich spans with request-scoped identifiers — see the
// Attr* constants for common keys.
func SetSpanAttr(ctx context.Context, key, value string) {
	if value == "" {
		return
	}
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	span.SetAttributes(attribute.String(key, value))
}
