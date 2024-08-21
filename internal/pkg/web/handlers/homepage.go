package handlers

import (
	"github.com/AlyxPink/meower/internal/pkg/web/routes"
	"github.com/AlyxPink/meower/internal/pkg/web/ui/pages/homepage"
	"github.com/gofiber/fiber/v2"
)

type Homepage struct{ *Server }

func NewHomepage(server *Server) *Homepage {
	homepage := &Homepage{server}
	homepage.registerRoutes()
	return homepage
}

func (h *Homepage) registerRoutes() {
	h.Get(routes.HomepageIndex, h.index)
}

func (h *Homepage) index(c *fiber.Ctx) error {
	return renderTempl(c, homepage.Index())
}
