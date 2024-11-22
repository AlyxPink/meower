package handlers

import (
	"context"
	"fmt"

	pb "github.com/AlyxPink/meower/api/proto"
	"github.com/AlyxPink/meower/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type meower struct {
	pb.UnimplementedMeowerServer
	db *pgxpool.Pool
}

func NewMeowerServer(db *pgxpool.Pool) pb.MeowerServer {
	return &meower{db: db}
}

func (s *meower) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	meow, err := db.New(s.db).CreateMeow(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &pb.CreateResponse{
		Meow: &pb.Meow{
			Id:        fmt.Sprintf("%x", meow.ID.Bytes),
			Name:      meow.Name,
			CreatedAt: meow.CreatedAt.Time.String(),
		},
	}, nil
}

func (s *meower) Index(ctx context.Context, req *pb.IndexRequest) (*pb.IndexResponse, error) {
	meows, err := db.New(s.db).IndexMeows(ctx)
	if err != nil {
		return nil, err
	}

	var resp []*pb.Meow
	for _, meow := range meows {
		resp = append(resp, &pb.Meow{
			Id:        fmt.Sprintf("%x", meow.ID.Bytes),
			Name:      meow.Name,
			CreatedAt: meow.CreatedAt.Time.String(),
		})
	}

	return &pb.IndexResponse{Meows: resp}, nil
}
