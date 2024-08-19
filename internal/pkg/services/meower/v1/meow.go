package meower

import (
	"context"
	"fmt"

	"github.com/AlyxPink/meower/internal/db"

	v1 "github.com/AlyxPink/meower/internal/pkg/api/meower/v1"
)

func (s *meower) Salutation(ctx context.Context, req *v1.SalutationRequest) (*v1.SalutationResponse, error) {
	meow := fmt.Sprintf("Meow %s :3", req.Name)

	return &v1.SalutationResponse{Meow: meow}, nil
}

func (s *meower) Create(ctx context.Context, req *v1.CreateRequest) (*v1.CreateResponse, error) {
	meow, err := db.New(s.db).CreateMeow(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &v1.CreateResponse{Meow: meow.Name}, nil
}

func (s *meower) List(ctx context.Context, req *v1.ListRequest) (*v1.ListResponse, error) {
	meows, err := db.New(s.db).ListMeows(ctx)
	if err != nil {
		return nil, err
	}

	var resp []*v1.Meow
	for _, meow := range meows {
		resp = append(resp, &v1.Meow{Meow: meow.Name})
	}

	return &v1.ListResponse{Meows: resp}, nil
}
