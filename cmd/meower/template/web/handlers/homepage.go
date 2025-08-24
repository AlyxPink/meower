package handlers

import (
	"TEMPLATE_MODULE_PATH/web/views"

	"github.com/gofiber/fiber/v2"
)

type Homepage struct{ *App }

func (h *Homepage) Homepage(c *fiber.Ctx) error {
	return renderTempl(c, views.Homepage(c))
}
