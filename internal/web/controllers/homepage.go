package controllers

import (
	viewHomepage "github.com/AlyxPink/meower/internal/web/views/pages/homepage"
	"github.com/gofiber/fiber/v2"
)

type Homepage struct{ *fiber.App }

func NewHomepage(app *fiber.App) *Homepage {
	homepage := &Homepage{app}
	return homepage
}

func (h *Homepage) Index(c *fiber.Ctx) error {
	return renderTempl(c, viewHomepage.Index(c))
}
