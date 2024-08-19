package web

import (
	meowerv1 "github.com/AlyxPink/meower/internal/pkg/api/meower/v1"
	ui "github.com/AlyxPink/meower/web/ui/services/meows/v1"
	"github.com/gofiber/fiber/v2"
)

type Meow struct{ *Server }

func NewMeow(server *Server) *Meow {
	meow := &Meow{server}
	meow.registerRoutes()
	return meow
}

func (h *Meow) registerRoutes() {
	h.Get("/meow/", h.new)
	h.Post("/meow/", h.create)
	h.Get("/meows/", h.list)
}

func (h *Meow) new(c *fiber.Ctx) error {
	return renderTempl(c, ui.New())
}

func (h *Meow) create(c *fiber.Ctx) error {
	name := c.FormValue("name")
	req := &meowerv1.CreateRequest{Name: name}

	resp, err := h.Services.Meower.Create(c.Context(), req)

	if err != nil {
		return err
	}

	return renderTempl(c, ui.Create(resp))
}

func (h *Meow) list(c *fiber.Ctx) error {
	req := &meowerv1.ListRequest{}

	resp, err := h.Services.Meower.List(c.Context(), req)

	if err != nil {
		return err
	}

	return renderTempl(c, ui.List(resp))
}
