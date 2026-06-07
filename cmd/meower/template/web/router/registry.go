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

	"github.com/gofiber/fiber/v3"
)

// Register binds a single route to one HTTP method, using the route's own
// Pattern() and Name(). It is shorthand for the common single-method case.
func Register(app fiber.Router, method string, route urls.URLer, handlers ...fiber.Handler) {
	RegisterRoute(app, []string{method}, route, handlers...)
}

// RegisterRoute binds a route to one or more HTTP methods. The route's
// Pattern() supplies the path and Name() supplies the Fiber route name (used
// for reverse URL lookup). A route with no handlers is ignored.
//
// Fiber v3's router takes the first handler as a distinct argument
// (Add(methods, path, handler, handlers...)) and accepts handlers as `any`, so
// we split the first handler out and widen the rest to []any.
func RegisterRoute(app fiber.Router, methods []string, route urls.URLer, handlers ...fiber.Handler) {
	if len(handlers) == 0 {
		return
	}

	rest := make([]any, len(handlers)-1)
	for i, h := range handlers[1:] {
		rest[i] = h
	}

	app.Add(methods, route.Pattern(), handlers[0], rest...).Name(route.Name())
}
