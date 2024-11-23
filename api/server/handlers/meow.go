package handlers

import (
	"context"
	"fmt"

	pb "github.com/AlyxPink/meower/api/proto"
	"github.com/AlyxPink/meower/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type meowServiceServer struct {
	pb.UnimplementedMeowServiceServer
	db *pgxpool.Pool
}

func NewMeowerServer(db *pgxpool.Pool) pb.MeowServiceServer {
	return &meowServiceServer{db: db}
}

func (s *meowServiceServer) Create(ctx context.Context, req *pb.CreateMeowRequest) (*pb.CreateMeowResponse, error) {
	meow, err := db.New(s.db).CreateMeow(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &pb.CreateMeowResponse{
		Meow: &pb.Meow{
			Id:        fmt.Sprintf("%x", meow.ID.Bytes),
			Name:      meow.Name,
			CreatedAt: meow.CreatedAt.Time.String(),
		},
	}, nil
}

func (s *meowServiceServer) Index(ctx context.Context, req *pb.IndexMeowRequest) (*pb.IndexMeowResponse, error) {
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

	return &pb.IndexMeowResponse{Meows: resp}, nil
}
