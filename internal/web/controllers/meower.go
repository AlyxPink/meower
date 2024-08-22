package controllers

import (
	meowerv1 "github.com/AlyxPink/meower/api/implementation/meower/v1"
	"github.com/AlyxPink/meower/internal/web/grpc"
	viewMeowV1 "github.com/AlyxPink/meower/internal/web/views/services/meows/v1"
	"github.com/gofiber/fiber/v2"
)

type Meower struct {
	*fiber.App
	GrpcClient *grpc.Client
}

func NewMeower(app *fiber.App, GrpcClient *grpc.Client) *Meower {
	meow := &Meower{
		App:        app,
		GrpcClient: GrpcClient,
	}
	return meow
}

func (h *Meower) New(c *fiber.Ctx) error {
	return renderTempl(c, viewMeowV1.New(c))
}

func (h *Meower) Create(c *fiber.Ctx) error {
	name := c.FormValue("name")
	req := &meowerv1.CreateRequest{Name: name}

	resp, err := h.GrpcClient.Meower.Create(c.Context(), req)

	if err != nil {
		return err
	}

	return renderTempl(c, viewMeowV1.Create(c, resp))
}

func (h *Meower) Index(c *fiber.Ctx) error {
	req := &meowerv1.IndexRequest{}

	resp, err := h.GrpcClient.Meower.Index(c.Context(), req)

	if err != nil {
		return err
	}

	return renderTempl(c, viewMeowV1.Index(c, resp))
}
