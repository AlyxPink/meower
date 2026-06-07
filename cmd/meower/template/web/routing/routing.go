package routing

import (
	"TEMPLATE_MODULE_PATH/pkg/urls"
	"TEMPLATE_MODULE_PATH/web/handlers"
	"TEMPLATE_MODULE_PATH/web/router"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// RegisterRoutes wires every route to its handler. Routes are defined as typed
// structs in pkg/urls (the single source of truth for paths and names) and
// registered here via router.Register, so the registered Fiber pattern is
// always exactly what pkg/urls builds URLs against.
func RegisterRoutes(app *handlers.App) {
	web := app.Web

	// Static assets. In Fiber v3 app.Static is gone; the static middleware
	// mounted on a wildcard route serves files and strips the route prefix
	// (so /static/css/x.css resolves to /src/web/static/public/css/x.css).
	web.Get("/static/*", static.New("/src/web/static/public")).Name("static")

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

	// Meows (authenticated users only).
	//
	// Order matters: the static /meows/new must register before the
	// parameterized /meows/:id, or Fiber would match "new" as an :id.
	authed := handlers.AuthMiddleware(app.SessionStore)
	meower := handlers.Meower{App: app}
	router.Register(web, fiber.MethodGet, urls.MeowIndex{}, authed, meower.Index)
	router.Register(web, fiber.MethodGet, urls.MeowNew{}, authed, meower.New)
	router.Register(web, fiber.MethodPost, urls.MeowCreate{}, authed, meower.Create)
	router.Register(web, fiber.MethodGet, urls.Meow{}, authed, meower.Show)
	router.Register(web, fiber.MethodGet, urls.MeowEdit{}, authed, meower.Edit)
	router.Register(web, fiber.MethodPost, urls.MeowUpdate{}, authed, meower.Update)
	router.Register(web, fiber.MethodPost, urls.MeowDelete{}, authed, meower.Delete)

	// Server-sent events: the browser subscribes at /events/stream and the hub
	// fans published events out to it. /events/health exposes hub metrics.
	sseHandler := &handlers.SSEHandler{App: app, Hub: app.Hub}
	router.Register(web, fiber.MethodGet, urls.SSEStream{}, sseHandler.Stream)
	router.Register(web, fiber.MethodGet, urls.SSEHealth{}, sseHandler.Health)
}
