package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	pbAuthorV1 "github.com/AlyxPink/meower/api/proto/author/v1"
	pbMeowV1 "github.com/AlyxPink/meower/api/proto/meow/v1"
	"github.com/AlyxPink/meower/api/server/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const (
	apiEndpoint = "localhost:50051"
)

func Serve() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a listener on TCP port for gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}
	defer lis.Close()

	// Create a new gRPC server
	g := grpc.NewServer()
	defer g.GracefulStop()

	// Register reflection service
	reflection.Register(g)

	// Register health check service
	grpc_health_v1.RegisterHealthServer(g, health.NewServer())

	// Create a new PostgreSQL connection pool
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	}

	// Register V1 services
	pbMeowV1.RegisterMeowServiceServer(g, handlers.NewMeowerServer(db))
	pbAuthorV1.RegisterAuthorServiceServer(g, handlers.NewAuthorServer(db))

	// Serve the gRPC server
	log.Printf("API server listening at %v", lis.Addr())
	if err := g.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
