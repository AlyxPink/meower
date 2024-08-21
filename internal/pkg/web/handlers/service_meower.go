package handlers

import (
	meowerv1 "github.com/AlyxPink/meower/api/gen/meower/v1"
	"github.com/AlyxPink/meower/internal/pkg/web/routes"
	ui "github.com/AlyxPink/meower/internal/pkg/web/ui/services/meows/v1"
	"github.com/gofiber/fiber/v2"
)

type Meow struct{ *Server }

func NewMeow(server *Server) *Meow {
	meow := &Meow{server}
	meow.registerRoutes()
	return meow
}

func (h *Meow) registerRoutes() {
	h.Get(routes.MeowNew, h.new)
	h.Post(routes.MeowCreate, h.create)
	h.Get(routes.MeowIndex, h.index)
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

func (h *Meow) index(c *fiber.Ctx) error {
	req := &meowerv1.ListRequest{}

	resp, err := h.Services.Meower.List(c.Context(), req)

	if err != nil {
		return err
	}

	return renderTempl(c, ui.Index(resp))
}
