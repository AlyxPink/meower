package handlers

import (
	"context"
	"fmt"

	"github.com/AlyxPink/meower/api/db"
	authorV1 "github.com/AlyxPink/meower/api/proto/author/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authorServiceServer struct {
	authorV1.UnimplementedAuthorServiceServer
	db *pgxpool.Pool
}

func NewAuthorServer(db *pgxpool.Pool) authorV1.AuthorServiceServer {
	return &authorServiceServer{db: db}
}

func (s *authorServiceServer) Create(ctx context.Context, req *authorV1.CreateAuthorRequest) (*authorV1.CreateAuthorResponse, error) {
	author, err := db.New(s.db).CreateAuthor(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	resp := &authorV1.CreateAuthorResponse{
		Author: &authorV1.Author{
			Id:        fmt.Sprintf("%x", author.ID.Bytes),
			Name:      author.Name,
			CreatedAt: author.CreatedAt.Time.String(),
		},
	}

	return resp, nil
}

func (s *authorServiceServer) Index(ctx context.Context, req *authorV1.IndexAuthorRequest) (*authorV1.IndexAuthorResponse, error) {
	authors, err := db.New(s.db).IndexAuthors(ctx)
	if err != nil {
		return nil, err
	}

	var resp []*authorV1.Author
	for _, author := range authors {
		resp = append(resp, &authorV1.Author{
			Id:        fmt.Sprintf("%x", author.ID.Bytes),
			Name:      author.Name,
			CreatedAt: author.CreatedAt.Time.String(),
		})
	}

	return &authorV1.IndexAuthorResponse{Authors: resp}, nil
}
