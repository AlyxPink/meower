package web

import (
	"github.com/AlyxPink/meower/web/grpc"
	"github.com/AlyxPink/meower/web/handlers"
	"github.com/AlyxPink/meower/web/routes"
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, GrpcClient *grpc.Client) {
	// Static
	app.Static("/static", "./internal/web/static/public").Name("static")

	// Homepage
	homepage := handlers.NewHomepage(app)
	app.Get("/", homepage.Index).Name(routes.HomeIndex)

	// Meower
	meower := handlers.NewMeower(app, GrpcClient)
	app.Get("/meow/", meower.Index).Name(routes.MeowIndex)
	app.Get("/meow/new", meower.New).Name(routes.MeowNew)
	app.Post("/meow/", meower.Create).Name(routes.MeowCreate)
}
