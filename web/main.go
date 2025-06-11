package main

import (
	"os"

	"github.com/AlyxPink/meower/web/grpc"
	"github.com/AlyxPink/meower/web/handlers"
	"github.com/AlyxPink/meower/web/routes/routing"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
)

func main() {
	// Connect to the internal gRPC API
	GrpcClient := grpc.NewClient()

	// Create session store
	sessionStore := session.New(session.Config{
		KeyLookup:      "cookie:session_id",
		CookieDomain:   "",
		CookiePath:     "/",
		CookieSecure:   os.Getenv("ENV") == "production",
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
	})

	// Create the Fiber app
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler:      handlers.ErrorHandler,
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

	app := &handlers.App{
		Web:          fiberApp,
		API:          GrpcClient,
		SessionStore: sessionStore,
	}

	// Mount public routes
	routing.RegisterRoutes(app)
	app.Web.Listen("0.0.0.0:3000")
}
