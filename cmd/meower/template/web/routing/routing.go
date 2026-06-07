package routing

import (
	"TEMPLATE_MODULE_PATH/pkg/urls"
	"TEMPLATE_MODULE_PATH/web/handlers"
	"TEMPLATE_MODULE_PATH/web/router"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes wires every route to its handler. Routes are defined as typed
// structs in pkg/urls (the single source of truth for paths and names) and
// registered here via router.Register, so the registered Fiber pattern is
// always exactly what pkg/urls builds URLs against.
func RegisterRoutes(app *handlers.App) {
	web := app.Web

	// Static assets
	web.Static("/static", "/src/web/static/public/").Name("static")

	// Homepage (available to all)
	homepage := handlers.Homepage{App: app}
	router.Register(web, fiber.MethodGet, urls.Homepage{},
		handlers.OptionalAuthMiddleware(app.SessionStore), homepage.Homepage)

	// Debug route (temporary)
	auth := handlers.Auth{App: app}
	web.Get("/debug/session", auth.Debug)

	// Authentication (guest only)
	guest := handlers.GuestMiddleware(app.SessionStore)
	router.Register(web, fiber.MethodGet, urls.LoginShow{}, guest, auth.ShowLogin)
	router.Register(web, fiber.MethodPost, urls.Login{}, guest, auth.Login)
	router.Register(web, fiber.MethodGet, urls.SignupShow{}, guest, auth.ShowSignup)
	router.Register(web, fiber.MethodPost, urls.Signup{}, guest, auth.Signup)

	// Logout (authenticated users only)
	router.Register(web, fiber.MethodPost, urls.Logout{},
		handlers.AuthMiddleware(app.SessionStore), auth.Logout)

	// Meows (authenticated users only)
	authed := handlers.AuthMiddleware(app.SessionStore)
	meower := handlers.Meower{App: app}
	router.Register(web, fiber.MethodGet, urls.MeowIndex{}, authed, meower.Index)
	router.Register(web, fiber.MethodGet, urls.MeowNew{}, authed, meower.New)
	router.Register(web, fiber.MethodPost, urls.MeowCreate{}, authed, meower.Create)

	// Server-sent events: the browser subscribes at /events/stream and the hub
	// fans published events out to it. /events/health exposes hub metrics.
	sseHandler := &handlers.SSEHandler{App: app, Hub: app.Hub}
	router.Register(web, fiber.MethodGet, urls.SSEStream{}, sseHandler.Stream)
	router.Register(web, fiber.MethodGet, urls.SSEHealth{}, sseHandler.Health)
}
