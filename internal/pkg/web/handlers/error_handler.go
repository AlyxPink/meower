package handlers

import (
	"errors"

	"github.com/AlyxPink/meower/internal/pkg/web/ui/pages/custom_errors"
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func ErrorHandler(ctx *fiber.Ctx, err error) error {
	log.Error(err)
	// Status code defaults to 500
	code := fiber.StatusInternalServerError

	// Retrieve the custom status code if it's a *fiber.Error
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	switch code {
	case 500:
		err = renderTempl(ctx, custom_errors.Error500(), templ.WithStatus(code))
	case 404:
		err = renderTempl(ctx, custom_errors.Error404(), templ.WithStatus(code))
	default:
		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	if err != nil {
		// In case we cannot render the template
		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	return nil
}
