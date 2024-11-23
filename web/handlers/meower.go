package handlers

import (
	meowV1 "github.com/AlyxPink/meower/api/proto/meow/v1"

	"github.com/AlyxPink/meower/web/grpc"
	viewMeowV1 "github.com/AlyxPink/meower/web/views/services/meows/v1"
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
	content := c.FormValue("name")
	req := &meowV1.CreateMeowRequest{Content: content}

	resp, err := h.GrpcClient.MeowService.CreateMeow(c.Context(), req)

	if err != nil {
		return err
	}

	return renderTempl(c, viewMeowV1.Create(c, resp))
}

func (h *Meower) Index(c *fiber.Ctx) error {
	req := &meowV1.IndexMeowRequest{}

	resp, err := h.GrpcClient.MeowService.IndexMeow(c.Context(), req)

	if err != nil {
		return err
	}

	return renderTempl(c, viewMeowV1.Index(c, resp))
}
