package handlers

import (
	"os"

	meowerv1 "github.com/AlyxPink/meower/api/gen/meower/v1"
	"github.com/AlyxPink/meower/internal/pkg/client"
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
	Services *services
}

type services struct {
	Meower meowerv1.MeowerSvcClient
}

func NewServer() *Server {
	// Create a new gRPC client
	grpcConn := client.NewClient()

	// Initialize the service clients
	services := &services{
		Meower: meowerv1.NewMeowerSvcClient(grpcConn),
	}

	// Create the Fiber app
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler:      ErrorHandler,
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
		App:      fiberApp,
		Services: services,
	}

	// Mount public routes
	server.SetupPublicRoutes()

	return server
}
