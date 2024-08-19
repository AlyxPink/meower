package services

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/grpc"

	meowerv1 "github.com/AlyxPink/meower/internal/pkg/api/meower/v1"
	"github.com/AlyxPink/meower/internal/pkg/services/meower/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterV1(ctx context.Context, s *grpc.Server) {
	// Create a new PostgreSQL connection pool
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	}

	meowerv1.RegisterMeowerSvcServer(s, meower.Meower(db))
}
