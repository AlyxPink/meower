package main

import (
	"context"
	"os"
	"strings"
	"time"

	"TEMPLATE_MODULE_PATH/pkg/observability"
	"TEMPLATE_MODULE_PATH/web/config"
	"TEMPLATE_MODULE_PATH/web/grpc"
	"TEMPLATE_MODULE_PATH/web/handlers"
	"TEMPLATE_MODULE_PATH/web/middleware"
	webobs "TEMPLATE_MODULE_PATH/web/observability"
	"TEMPLATE_MODULE_PATH/web/routing"
	"TEMPLATE_MODULE_PATH/web/sse"

	"github.com/charmbracelet/log"
	otelfiber "github.com/gofiber/contrib/v3/otel"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/encryptcookie"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/storage/redis/v3"
	"github.com/gofiber/utils/v2"
)

func main() {
	ctx := context.Background()

	// Load configuration and configure structured logging.
	cfg := config.Load()
	observability.ConfigureLogging(cfg.Environment, cfg.LogLevel)

	// Initialize OpenTelemetry tracing (OTLP/HTTP). Spans export to
	// OTEL_EXPORTER_OTLP_ENDPOINT (defaults to localhost:4318).
	telemetry, err := webobs.InitTelemetry(ctx, webobs.DefaultConfig())
	if err != nil {
		log.Warn("Failed to initialize telemetry; continuing without tracing", "error", err)
	} else {
		defer func() {
			if err := telemetry.Shutdown(context.Background()); err != nil {
				log.Error("Failed to shut down telemetry", "error", err)
			}
		}()
	}

	// Connect to the internal gRPC API
	GrpcClient := grpc.NewClient()

	// Create Redis storage
	redisStore := redis.New(redis.Config{
		URL: cfg.RedisURL,
	})

	// Create session store with Redis storage
	sessionStore := session.NewStore(session.Config{
		Storage:        redisStore,
		Extractor:      extractors.FromCookie("session_id"),
		CookieDomain:   "",
		CookiePath:     "/",
		CookieSecure:   os.Getenv("ENV") == "production",
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
		IdleTimeout:    365 * 24 * time.Hour, // 1 year
	})

	// Create the Fiber app
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler: handlers.ErrorHandler,
	})

	// Observability middleware, mounted first so every request is traced:
	//   otelfiber          — opens the request span
	//   SetTraceIDHeader   — echoes the trace ID back as X-Trace-Id
	//   EnrichTraceContext — decorates the span with request-shape attributes
	//
	// Static assets are skipped: the v3 otel middleware wraps streamed
	// response bodies to measure their size, but the static middleware serves
	// files as a body stream that the wrapper can't drive, which hangs the
	// response. Static files don't need tracing, so we opt them out here.
	fiberApp.Use(otelfiber.Middleware(otelfiber.WithNext(func(c fiber.Ctx) bool {
		return strings.HasPrefix(c.Path(), "/static")
	})))
	fiberApp.Use(middleware.SetTraceIDHeader())
	fiberApp.Use(middleware.EnrichTraceWithContext())

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

	// Create the SSE hub and start its reaper (evicts stale connections).
	hub := sse.NewHub()
	go hub.StartReaper(ctx.Done())

	app := &handlers.App{
		Web:          fiberApp,
		API:          GrpcClient,
		SessionStore: sessionStore,
		Hub:          hub,
	}

	// Mount public routes
	routing.RegisterRoutes(app)
	if err := app.Web.Listen("0.0.0.0:3000", fiber.ListenConfig{EnablePrintRoutes: true}); err != nil {
		panic(err)
	}
}
