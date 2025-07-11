package handlers

import (
	"errors"

	"TEMPLATE_MODULE_PATH/web/grpc"
	"TEMPLATE_MODULE_PATH/web/views/pages/custom_errors"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/session"

	"github.com/charmbracelet/log"
)

type App struct {
	Web          *fiber.App
	API          *grpc.Client
	SessionStore *session.Store
}

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
		err = renderTempl(ctx, custom_errors.Error500(ctx), templ.WithStatus(code))
	case 404:
		err = renderTempl(ctx, custom_errors.Error404(ctx), templ.WithStatus(code))
	default:
		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	if err != nil {
		// In case we cannot render the template
		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	return nil
}

// helper function to render a component using a-h/templ
func renderTempl(c *fiber.Ctx, component templ.Component, options ...func(*templ.ComponentHandler)) error {
	componentHandler := templ.Handler(component)
	for _, o := range options {
		o(componentHandler)
	}
	return adaptor.HTTPHandler(componentHandler)(c)
}
