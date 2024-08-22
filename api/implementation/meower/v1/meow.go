package v1

import (
	"context"
	"fmt"

	"github.com/AlyxPink/meower/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type meower struct {
	UnimplementedMeowerSvcServer
	db *pgxpool.Pool
}

func Meower(db *pgxpool.Pool) MeowerSvcServer {
	return &meower{db: db}
}

func (s *meower) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	meow, err := db.New(s.db).CreateMeow(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &CreateResponse{
		Meow: &Meow{
			Id:        fmt.Sprintf("%x", meow.ID.Bytes),
			Name:      meow.Name,
			CreatedAt: meow.CreatedAt.Time.String(),
		},
	}, nil
}

func (s *meower) Index(ctx context.Context, req *IndexRequest) (*IndexResponse, error) {
	meows, err := db.New(s.db).IndexMeows(ctx)
	if err != nil {
		return nil, err
	}

	var resp []*Meow
	for _, meow := range meows {
		resp = append(resp, &Meow{
			Id:        fmt.Sprintf("%x", meow.ID.Bytes),
			Name:      meow.Name,
			CreatedAt: meow.CreatedAt.Time.String(),
		})
	}

	return &IndexResponse{Meows: resp}, nil
}
