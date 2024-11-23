package handlers

import (
	viewHomepage "github.com/AlyxPink/meower/web/views/pages/homepage"
	"github.com/gofiber/fiber/v2"
)

func HomepageIndex(c *fiber.Ctx) error {
	return renderTempl(c, viewHomepage.Index(c))
}
