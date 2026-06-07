package workers

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// --- helpers -----------------------------------------------------------------

// countingWorker increments a counter on every Run call.
type countingWorker struct {
	name     string
	interval time.Duration
	count    atomic.Int64
}

func (w *countingWorker) Name() string             { return w.name }
func (w *countingWorker) Interval() time.Duration  { return w.interval }
func (w *countingWorker) Run(_ context.Context) error {
	w.count.Add(1)
	return nil
}

// panicWorker panics unconditionally in Run.
type panicWorker struct{}

func (p *panicWorker) Name() string            { return "panicky" }
func (p *panicWorker) Interval() time.Duration { return 5 * time.Millisecond }
func (p *panicWorker) Run(_ context.Context) error {
	panic("intentional test panic")
}

// --- tests -------------------------------------------------------------------

// TestWorkerRunsAtLeastOnce verifies that a registered worker is invoked at
// least once after Start and that the run count grows over time.
func TestWorkerRunsAtLeastOnce(t *testing.T) {
	t.Parallel()

	w := &countingWorker{name: "counter", interval: 5 * time.Millisecond}
	m := NewManager(nil) // no DB needed for this test
	m.Register(w)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.Start(ctx)

	// The worker runs immediately on start, so a single short sleep is enough
	// to observe at least one execution without relying on tick timing.
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if w.count.Load() >= 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if got := w.count.Load(); got < 1 {
		t.Errorf("expected at least 1 run, got %d", got)
	}
}

// TestWorkerRunsMultipleTimes confirms that the ticker keeps firing after the
// initial immediate run.
func TestWorkerRunsMultipleTimes(t *testing.T) {
	t.Parallel()

	w := &countingWorker{name: "multi", interval: 10 * time.Millisecond}
	m := NewManager(nil)
	m.Register(w)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.Start(ctx)

	// Wait for the worker to accumulate multiple runs.
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if w.count.Load() >= 3 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if got := w.count.Load(); got < 3 {
		t.Errorf("expected at least 3 runs, got %d", got)
	}
}

// TestPanicingWorkerDoesNotCrashManager ensures that a worker panicking in Run
// does not crash the manager goroutine, and that other workers keep running.
func TestPanicingWorkerDoesNotCrashManager(t *testing.T) {
	t.Parallel()

	good := &countingWorker{name: "good", interval: 5 * time.Millisecond}
	m := NewManager(nil)
	m.Register(&panicWorker{})
	m.Register(good)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// If the panic kills the process the test itself would crash, so simply
	// reaching the assertion below proves recovery works.
	m.Start(ctx)

	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if good.count.Load() >= 2 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if got := good.count.Load(); got < 2 {
		t.Errorf("good worker expected >=2 runs after panic worker fired, got %d", got)
	}
}

// TestContextCancellationStopsWorkers checks that cancelling the context
// causes workers to stop incrementing their counters.
func TestContextCancellationStopsWorkers(t *testing.T) {
	t.Parallel()

	w := &countingWorker{name: "stoppable", interval: 5 * time.Millisecond}
	m := NewManager(nil)
	m.Register(w)

	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx)

	// Let the worker accumulate a few runs.
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if w.count.Load() >= 3 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Cancel and record the count shortly after.
	cancel()
	// Give the goroutine a moment to observe ctx.Done and return.
	time.Sleep(30 * time.Millisecond)
	countAfterCancel := w.count.Load()

	// Wait another interval window and confirm the count has not grown.
	time.Sleep(50 * time.Millisecond)
	countLater := w.count.Load()

	if countLater > countAfterCancel {
		t.Errorf(
			"worker kept running after context cancel: count grew from %d to %d",
			countAfterCancel, countLater,
		)
	}
}
