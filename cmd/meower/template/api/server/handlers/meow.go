package handlers

import (
	"context"
	"fmt"

	"{{.ModulePath}}/api/db"
	meowV1 "{{.ModulePath}}/api/proto/meow/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type meowServiceServer struct {
	meowV1.UnimplementedMeowServiceServer
	db *pgxpool.Pool
}

func NewMeowerServer(db *pgxpool.Pool) meowV1.MeowServiceServer {
	return &meowServiceServer{db: db}
}

func (s *meowServiceServer) CreateMeow(ctx context.Context, req *meowV1.CreateMeowRequest) (*meowV1.CreateMeowResponse, error) {
	meow, err := db.New(s.db).CreateMeow(ctx, req.Content)
	if err != nil {
		return nil, err
	}
	resp := &meowV1.CreateMeowResponse{
		Meow: &meowV1.Meow{
			Id:        fmt.Sprintf("%x", meow.ID.Bytes),
			Content:   meow.Content,
			CreatedAt: timestamppb.New(meow.CreatedAt.Time),
		},
	}

	return resp, nil
}

func (s *meowServiceServer) IndexMeow(ctx context.Context, req *meowV1.IndexMeowRequest) (*meowV1.IndexMeowResponse, error) {
	meows, err := db.New(s.db).IndexMeows(ctx)
	if err != nil {
		return nil, err
	}

	var resp []*meowV1.Meow
	for _, meow := range meows {
		resp = append(resp, &meowV1.Meow{
			Id:        fmt.Sprintf("%x", meow.ID.Bytes),
			Content:   meow.Content,
			CreatedAt: timestamppb.New(meow.CreatedAt.Time),
		})
	}

	return &meowV1.IndexMeowResponse{Meows: resp}, nil
}
