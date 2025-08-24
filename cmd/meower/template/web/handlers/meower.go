package handlers

import (
	meowV1 "TEMPLATE_MODULE_PATH/api/proto/meow/v1"

	"TEMPLATE_MODULE_PATH/web/views"

	"github.com/gofiber/fiber/v2"
)

type Meower struct{ *App }

func (h *Meower) New(c *fiber.Ctx) error {
	return renderTempl(c, views.NewMeow(c))
}

func (h *Meower) Create(c *fiber.Ctx) error {
	content := c.FormValue("content")
	req := &meowV1.CreateMeowRequest{Content: content}

	resp, err := h.API.MeowService.CreateMeow(c.Context(), req)
	if err != nil {
		return err
	}

	return renderTempl(c, views.CreateMeow(c, resp))
}

func (h *Meower) Index(c *fiber.Ctx) error {
	req := &meowV1.IndexMeowRequest{}

	resp, err := h.API.MeowService.IndexMeow(c.Context(), req)
	if err != nil {
		return err
	}

	return renderTempl(c, views.IndexMeows(c, resp))
}
