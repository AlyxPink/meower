// Package middleware provides Fiber middleware for the web server: trace
// enrichment and trace-ID response headers. (Auth/session middleware lives in
// the handlers package and the auth scaffold.)
package middleware

import (
	"TEMPLATE_MODULE_PATH/pkg/urls"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Span attribute keys set by EnrichTraceWithContext.
const (
	attrHTMX = "request.htmx"
	attrSSE  = "request.sse"
)

// sseStreamPath is the SSE endpoint path, resolved once at init from the urls
// kit so this and the route registration can't drift.
var sseStreamPath = urls.SSEStream{}.URL()

// EnrichTraceWithContext returns a Fiber middleware that annotates the active
// OTel span with request-shape attributes (HTMX request? SSE stream?). It runs
// the rest of the stack first, then writes attributes while the span is still
// open (otelfiber closes it after the whole middleware stack returns).
//
// Extension point: once the auth scaffold is enabled, read the session here and
// add an `enduser.id` attribute so traces can be filtered by user — see the
// auth scaffold for the session key.
func EnrichTraceWithContext() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		span := trace.SpanFromContext(c.UserContext())
		if !span.IsRecording() {
			return err
		}

		if c.Get("HX-Request") == "true" {
			span.SetAttributes(attribute.Bool(attrHTMX, true))
		}
		if c.Path() == sseStreamPath {
			span.SetAttributes(attribute.Bool(attrSSE, true))
		}

		return err
	}
}

// SetTraceIDHeader returns a Fiber middleware that writes the current trace ID
// as an X-Trace-Id response header, so a browser request can be correlated with
// its Grafana Tempo trace. No header is written when there is no active trace.
func SetTraceIDHeader() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		sc := trace.SpanFromContext(c.UserContext()).SpanContext()
		if sc.IsValid() {
			c.Set("X-Trace-Id", sc.TraceID().String())
		}

		return err
	}
}
