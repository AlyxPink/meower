package server

import (
	"context"

	"TEMPLATE_MODULE_PATH/api/server/workers"

	"github.com/jackc/pgx/v5/pgxpool"
)

// startWorkers builds the worker manager, registers the application's
// background workers, and starts them. They run until ctx is cancelled.
//
// Register your own workers here. This file is part of the worker scaffold and
// is omitted when the project is generated with --no-workers (replaced by a
// no-op stub).
func startWorkers(ctx context.Context, db *pgxpool.Pool) {
	mgr := workers.NewManager(db)

	// Example worker — replace with your own (cleanup jobs, digest emailers, …).
	mgr.Register(&workers.HeartbeatWorker{})

	mgr.Start(ctx)
}
