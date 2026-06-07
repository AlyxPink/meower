package observability

import (
	"sync"

	"github.com/charmbracelet/log"
	"go.opentelemetry.io/otel"
)

// quietErrorHandler is the global OpenTelemetry error handler. The SDK routes
// every failed span export through it, which — with a batch processor retrying
// on a timer — means a collector that is down (e.g. the monitoring stack isn't
// running) produces an identical "connection refused" line every few seconds.
//
// This handler collapses that spam: the first time a distinct error message is
// seen it logs a single warning; repeats of the same message drop to debug, so
// a genuine misconfiguration is still discoverable without flooding the logs.
type quietErrorHandler struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

// Handle implements otel.ErrorHandler.
func (h *quietErrorHandler) Handle(err error) {
	if err == nil {
		return
	}

	msg := err.Error()

	h.mu.Lock()
	_, repeat := h.seen[msg]
	if !repeat {
		h.seen[msg] = struct{}{}
	}
	h.mu.Unlock()

	if repeat {
		// Already warned about this exact failure; keep subsequent occurrences
		// at debug level so logs stay quiet when no collector is reachable.
		log.Debug("otel error (repeated)", "error", msg)
		return
	}

	log.Warn(
		"otel error; traces may not be exported (is a collector reachable at OTEL_EXPORTER_OTLP_ENDPOINT?)",
		"error", msg,
	)
}

// SetQuietErrorHandler installs the deduplicating OpenTelemetry error handler as
// the global handler. Call it once during startup, before initializing the
// trace provider, so export failures are reported at most once rather than on
// every batch tick. This keeps a plain `docker compose up` (no monitoring
// stack) from spamming the logs while still surfacing a real misconfiguration.
func SetQuietErrorHandler() {
	otel.SetErrorHandler(&quietErrorHandler{seen: make(map[string]struct{})})
}
