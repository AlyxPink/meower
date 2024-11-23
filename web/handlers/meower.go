package handlers

import (
	meowV1 "github.com/AlyxPink/meower/api/proto/meow/v1"

	"github.com/AlyxPink/meower/web/grpc"
	viewMeowV1 "github.com/AlyxPink/meower/web/views/services/meows/v1"
	"github.com/gofiber/fiber/v2"
)

func MeowerNew(c *fiber.Ctx) error {
	return renderTempl(c, viewMeowV1.New(c))
}

func MeowerCreate(c *fiber.Ctx) error {
	content := c.FormValue("name")
	req := &meowV1.CreateMeowRequest{Content: content}

	resp, err := grpc.NewClient().MeowService.CreateMeow(c.Context(), req)

	if err != nil {
		return err
	}

	return renderTempl(c, viewMeowV1.Create(c, resp))
}

func MeowerIndex(c *fiber.Ctx) error {
	req := &meowV1.IndexMeowRequest{}

	resp, err := grpc.NewClient().MeowService.IndexMeow(c.Context(), req)

	if err != nil {
		return err
	}

	return renderTempl(c, viewMeowV1.Index(c, resp))
}
