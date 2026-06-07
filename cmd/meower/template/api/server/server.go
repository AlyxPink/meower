package server

import (
	"context"
	"fmt"
	"net"
	"os"

	pbMeowV1 "TEMPLATE_MODULE_PATH/api/proto/meow/v1"
	pbUserV1 "TEMPLATE_MODULE_PATH/api/proto/user/v1"
	"TEMPLATE_MODULE_PATH/api/observability"
	"TEMPLATE_MODULE_PATH/api/server/handlers"

	"github.com/charmbracelet/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func Serve() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configure structured logging (JSON in production for Loki).
	observability.ConfigureLogging(os.Getenv("ENVIRONMENT"), os.Getenv("LOG_LEVEL"))

	// Initialize OpenTelemetry tracing. Spans export over OTLP to the endpoint
	// in OTEL_EXPORTER_OTLP_ENDPOINT (defaults to localhost:4317).
	telemetry, err := observability.InitTelemetry(ctx, observability.DefaultConfig())
	if err != nil {
		log.Warn("Failed to initialize telemetry; continuing without tracing", "error", err)
	} else {
		defer func() {
			if err := telemetry.Shutdown(context.Background()); err != nil {
				log.Error("Failed to shut down telemetry", "error", err)
			}
		}()
	}

	// Start the Prometheus metrics endpoint on :9091 (/metrics, /health).
	metricsServer, err := observability.NewMetricsServer(":9091")
	if err != nil {
		log.Warn("Failed to create metrics server; continuing without metrics", "error", err)
	} else {
		metricsServer.Start()
		defer func() {
			if err := metricsServer.Shutdown(context.Background()); err != nil {
				log.Error("Failed to shut down metrics server", "error", err)
			}
		}()
	}

	// Create a listener on TCP port for gRPC server.
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	// gRPC server with OpenTelemetry instrumentation so every RPC is a span.
	g := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	defer g.GracefulStop()

	// Register reflection service.
	reflection.Register(g)

	// Register health check service.
	grpc_health_v1.RegisterHealthServer(g, health.NewServer())

	// Create a traced PostgreSQL connection pool. Every query becomes a span
	// nested under its gRPC request span.
	db, err := observability.NewTracedPool(ctx, os.Getenv("DATABASE_URL"), observability.DefaultDBTracerConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	}

	// Register V1 services.
	pbMeowV1.RegisterMeowServiceServer(g, handlers.NewMeowerServer(db))
	pbUserV1.RegisterUserServiceServer(g, handlers.NewUserServer(db))

	// Serve the gRPC server.
	log.Info("API server listening", "addr", lis.Addr().String())
	if err := g.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
