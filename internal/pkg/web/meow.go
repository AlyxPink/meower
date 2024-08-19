package web

import (
	meowerv1 "github.com/AlyxPink/meower/internal/pkg/api/meower/v1"
	"github.com/AlyxPink/meower/web/ui"
	"github.com/gofiber/fiber/v2"
)

type Meow struct{ *Server }

func NewMeow(server *Server) *Meow {
	meow := &Meow{server}
	meow.registerRoutes()
	return meow
}

func (h *Meow) registerRoutes() {
	h.Get("/meow/:name", h.name)
}

func (h *Meow) name(c *fiber.Ctx) error {
	name := c.Params("name")
	req := &meowerv1.SalutationRequest{Name: name}

	resp, err := h.Services.Meower.Salutation(c.Context(), req)

	if err != nil {
		return err
	}

	return renderTempl(c, ui.Meow(resp))
}
