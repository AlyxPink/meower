package api

import (
	"net"
	"os"

	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	meowerv1 "github.com/AlyxPink/meower/api/implementation/meower/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	apiEndpoint = "localhost:50051"
)

func NewServer() *grpc.Server {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a listener on TCP port for gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}
	defer lis.Close()

	// Create a new gRPC server
	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()

	// Register reflection service
	reflection.Register(grpcServer)

	// Register health check service
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

	// Register V1 services
	RegisterV1(ctx, grpcServer)

	// Serve the gRPC server
	log.Printf("API server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	return grpcServer
}

func RegisterV1(ctx context.Context, s *grpc.Server) {
	// Create a new PostgreSQL connection pool
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	}

	meowerv1.RegisterMeowerSvcServer(s, meowerv1.Meower(db))
}

func getApiEndpoint() string {
	if os.Getenv("API_ENDPOINT") != "" {
		return os.Getenv("API_ENDPOINT")
	}
	return apiEndpoint
}
