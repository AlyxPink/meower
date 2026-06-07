// Package workers provides a lightweight background-worker harness for the API
// server. It defines the Worker interface and a Manager that runs each worker
// on its own goroutine, ticking at the interval the worker declares.
//
// # Adding a worker
//
//  1. Implement the Worker interface (Name, Interval, Run).
//  2. Call manager.Register(yourWorker) before manager.Start.
//  3. That's it — the manager handles scheduling, error logging, and clean
//     shutdown when the context is cancelled.
//
// Workers are started in server.Serve after the database pool is ready so they
// can safely use the pool passed to the Manager.
package workers

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Worker is the interface every background job must implement.
//
// Name returns a stable human-readable identifier used in log output.
// Interval controls how often Run is called; the manager also calls Run once
// immediately on start before the first tick.
// Run performs the job's work. It should respect ctx cancellation for any
// long-running operations. A non-nil error is logged but does not stop the
// worker — transient failures are expected.
type Worker interface {
	Name() string
	Interval() time.Duration
	Run(ctx context.Context) error
}

// Manager owns the set of registered workers and coordinates their lifecycle.
// The DB pool is available to workers that need database access; pass it
// through when constructing your worker or embed it directly in the struct.
type Manager struct {
	db      *pgxpool.Pool
	workers []Worker
	mu      sync.Mutex
}

// NewManager creates an idle Manager. Call Register to add workers and Start
// to begin executing them.
func NewManager(db *pgxpool.Pool) *Manager {
	return &Manager{db: db}
}

// DB returns the pool so worker constructors that receive *Manager can obtain
// the shared pool without it being part of the Worker interface.
func (m *Manager) DB() *pgxpool.Pool {
	return m.db
}

// Register adds w to the set of managed workers. Register must be called
// before Start; calling it after Start has no effect on already-running loops.
func (m *Manager) Register(w Worker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workers = append(m.workers, w)
}

// Start launches a goroutine for each registered worker and returns
// immediately. Each goroutine:
//
//   - Calls w.Run once right away (so work happens at startup, not after the
//     first full interval).
//   - Then waits for the ticker and calls w.Run on each tick.
//   - Logs any error returned by Run without stopping the loop (transient
//     failures should not kill a worker permanently).
//   - Recovers from panics inside Run so one buggy worker cannot crash the
//     server process; the panic value and stack are logged at error level and
//     the goroutine continues to the next tick.
//   - Returns as soon as ctx is done, so shutting down the server context
//     cleanly stops all workers.
func (m *Manager) Start(ctx context.Context) {
	m.mu.Lock()
	snapshot := make([]Worker, len(m.workers))
	copy(snapshot, m.workers)
	m.mu.Unlock()

	for _, w := range snapshot {
		w := w // capture loop variable
		go func() {
			log.Info("worker started", "worker", w.Name(), "interval", w.Interval())

			// Run once immediately before the first tick.
			runWorker(ctx, w)

			ticker := time.NewTicker(w.Interval())
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					log.Info("worker stopped", "worker", w.Name())
					return
				case <-ticker.C:
					runWorker(ctx, w)
				}
			}
		}()
	}
}

// runWorker calls w.Run inside a deferred recover so a panic cannot propagate
// out of the goroutine. Errors are logged with the worker name; panics are
// logged with a stack trace.
func runWorker(ctx context.Context, w Worker) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(
				"panic recovered in worker",
				"worker", w.Name(),
				"panic", fmt.Sprintf("%v", r),
				"stack", string(debug.Stack()),
			)
		}
	}()

	if err := w.Run(ctx); err != nil {
		log.Error("worker run failed", "worker", w.Name(), "error", err)
	}
}
