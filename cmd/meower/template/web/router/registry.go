// Package router registers urls.URLer routes with a Fiber app. It is the
// consuming half of the type-safe routing kit: route definitions live in
// pkg/urls (the single source of truth for both URLs and patterns), and this
// package wires those patterns to handlers.
//
// Registering a route through RegisterRoute (rather than calling app.Get with a
// string path) guarantees the registered pattern is exactly the one pkg/urls
// will build URLs against — rename a route and the build breaks instead of
// silently 404ing at runtime.
package router

import (
	"TEMPLATE_MODULE_PATH/pkg/urls"

	"github.com/gofiber/fiber/v2"
)

// Register binds a single route to one HTTP method, using the route's own
// Pattern() and Name(). It is shorthand for the common single-method case.
func Register(app fiber.Router, method string, route urls.URLer, handlers ...fiber.Handler) {
	RegisterRoute(app, []string{method}, route, handlers...)
}

// RegisterRoute binds a route to one or more HTTP methods. The route's
// Pattern() supplies the path and Name() supplies the Fiber route name (used
// for reverse URL lookup). Unknown methods are ignored.
func RegisterRoute(app fiber.Router, methods []string, route urls.URLer, handlers ...fiber.Handler) {
	for _, method := range methods {
		switch method {
		case fiber.MethodGet:
			app.Get(route.Pattern(), handlers...).Name(route.Name())
		case fiber.MethodPost:
			app.Post(route.Pattern(), handlers...).Name(route.Name())
		case fiber.MethodPut:
			app.Put(route.Pattern(), handlers...).Name(route.Name())
		case fiber.MethodDelete:
			app.Delete(route.Pattern(), handlers...).Name(route.Name())
		case fiber.MethodPatch:
			app.Patch(route.Pattern(), handlers...).Name(route.Name())
		case fiber.MethodOptions:
			app.Options(route.Pattern(), handlers...).Name(route.Name())
		}
	}
}
