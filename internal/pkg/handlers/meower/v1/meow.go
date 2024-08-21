package meower

import (
	"context"
	"fmt"

	"github.com/AlyxPink/meower/internal/db"

	v1 "github.com/AlyxPink/meower/api/gen/meower/v1"
)

func (s *meower) Create(ctx context.Context, req *v1.CreateRequest) (*v1.CreateResponse, error) {
	meow, err := db.New(s.db).CreateMeow(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &v1.CreateResponse{
		Meow: &v1.Meow{
			Id:        fmt.Sprintf("%x", meow.ID.Bytes),
			Name:      meow.Name,
			CreatedAt: meow.CreatedAt.Time.String(),
		},
	}, nil
}

func (s *meower) Index(ctx context.Context, req *v1.IndexRequest) (*v1.IndexResponse, error) {
	meows, err := db.New(s.db).IndexMeows(ctx)
	if err != nil {
		return nil, err
	}

	var resp []*v1.Meow
	for _, meow := range meows {
		resp = append(resp, &v1.Meow{
			Id:        fmt.Sprintf("%x", meow.ID.Bytes),
			Name:      meow.Name,
			CreatedAt: meow.CreatedAt.Time.String(),
		})
	}

	return &v1.IndexResponse{Meows: resp}, nil
}
