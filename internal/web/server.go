package web

import (
	"os"

	"github.com/AlyxPink/meower/internal/web/controllers"
	"github.com/AlyxPink/meower/internal/web/grpc"
	"github.com/AlyxPink/meower/internal/web/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"
)

type Server struct {
	*fiber.App
	GrpcClient *grpc.Client
}

func NewServer() *Server {
	// Connect to the internal gRPC API
	GrpcClient := grpc.NewClient()

	// Create the Fiber app
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler:      controllers.ErrorHandler,
		EnablePrintRoutes: true,
	})

	// Add middlewares
	if os.Getenv("ENV") == "production" {
		fiberApp.Use(compress.New()) // Enable gzip compression in production only, templ proxy does not support brotli
		fiberApp.Use(csrf.New())
	} else {
		fiberApp.Use(logger.New()) // Enable request logging in development
	}
	fiberApp.Use(requestid.New(requestid.Config{Generator: utils.UUIDv4}))
	fiberApp.Use(encryptcookie.New(encryptcookie.Config{
		Key: os.Getenv("COOKIE_SECRET_KEY"),
	}))

	// Build server instance
	server := &Server{
		App:        fiberApp,
		GrpcClient: GrpcClient,
	}

	// Mount public routes
	RegisterRoutes(server.App, GrpcClient)

	return server
}

func RegisterRoutes(app *fiber.App, GrpcClient *grpc.Client) {
	// Static
	app.Static("/static", "./internal/web/static/public").Name("static")

	// Homepage
	homepage := controllers.NewHomepage(app)
	app.Get("/", homepage.Index).Name(routes.HomeIndex)

	// Meower
	meower := controllers.NewMeower(app, GrpcClient)
	app.Get("/meow/", meower.Index).Name(routes.MeowIndex)
	app.Get("/meow/new", meower.New).Name(routes.MeowNew)
	app.Post("/meow/", meower.Create).Name(routes.MeowCreate)
}
