package workers

// Replace this file with your own workers — e.g. a cleanup job that deletes
// expired sessions, a digest emailer that sends daily summaries, or a metrics
// refresher that pre-computes expensive aggregates. Each worker is its own
// struct that implements the Worker interface; register it in server.Serve with
// manager.Register(NewHeartbeatWorker()).

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
)

// HeartbeatWorker is a minimal do-nothing worker that demonstrates the
// interface. It logs a debug line on every tick so you can confirm the harness
// is running without any side-effects. Delete or replace it once you have real
// workers.
type HeartbeatWorker struct{}

// NewHeartbeatWorker returns a ready-to-register HeartbeatWorker.
func NewHeartbeatWorker() *HeartbeatWorker {
	return &HeartbeatWorker{}
}

// Name implements Worker.
func (h *HeartbeatWorker) Name() string { return "heartbeat" }

// Interval implements Worker. One minute is a sensible default for a liveness
// signal; adjust to taste.
func (h *HeartbeatWorker) Interval() time.Duration { return time.Minute }

// Run implements Worker. It emits a single debug log line so that the worker
// appears in structured logs without cluttering them.
func (h *HeartbeatWorker) Run(_ context.Context) error {
	log.Debug("heartbeat worker tick")
	return nil
}
