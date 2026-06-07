package observability

import (
	"context"
	"os"

	"github.com/amirsalarsafaei/sqlc-pgx-monitoring/dbtracer"
	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
)

// DBTracerConfig holds configuration for database tracing.
type DBTracerConfig struct {
	DatabaseName      string
	LogArgs           bool // Whether to log query arguments
	LogArgsLenLimit   int  // Maximum length of logged arguments
	IncludeSQLText    bool // Whether to include SQL text in traces
	IncludeSpanSuffix bool // Whether to add query name to span name
	LogAllQueries     bool // If false, only log queries with errors
}

// DefaultDBTracerConfig returns sensible defaults for database tracing. Argument
// logging and full-query logging are enabled only in development; production
// logs queries only on error to keep log volume down.
func DefaultDBTracerConfig() DBTracerConfig {
	env := os.Getenv("ENVIRONMENT")
	isDev := env == "" || env == "development"

	return DBTracerConfig{
		DatabaseName:      "TEMPLATE_PROJECT_NAME",
		LogArgs:           isDev,
		LogArgsLenLimit:   500,
		IncludeSQLText:    true,
		IncludeSpanSuffix: true,
		LogAllQueries:     isDev && os.Getenv("LOG_ALL_QUERIES") == "true",
	}
}

// NewTracedPool creates a pgxpool.Pool with OpenTelemetry tracing enabled. It
// wraps every database query in a span carrying the query name (from SQLC),
// SQL text, duration, and any error. Those spans nest under the gRPC request
// span, giving a complete trace: gRPC method → DB queries.
func NewTracedPool(ctx context.Context, databaseURL string, cfg DBTracerConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	tracer, err := dbtracer.NewDBTracer(
		cfg.DatabaseName,
		dbtracer.WithTraceProvider(otel.GetTracerProvider()),
		dbtracer.WithMeterProvider(otel.GetMeterProvider()),
		dbtracer.WithLogArgs(cfg.LogArgs),
		dbtracer.WithLogArgsLenLimit(cfg.LogArgsLenLimit),
		dbtracer.WithIncludeSQLText(cfg.IncludeSQLText),
		dbtracer.WithIncludeSpanNameSuffix(cfg.IncludeSpanSuffix),
		dbtracer.WithShouldLog(func(err error) bool {
			if cfg.LogAllQueries {
				return true
			}
			return err != nil // Only log errors in production.
		}),
	)
	if err != nil {
		return nil, err
	}

	poolConfig.ConnConfig.Tracer = tracer

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	log.Info(
		"Database pool created with tracing enabled",
		"database", cfg.DatabaseName,
		"log_args", cfg.LogArgs,
		"include_sql", cfg.IncludeSQLText,
	)

	return pool, nil
}
