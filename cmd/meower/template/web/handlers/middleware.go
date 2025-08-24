package handlers

import (
	"github.com/charmbracelet/log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// AuthMiddleware checks if the user is authenticated
func AuthMiddleware(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Debug("AuthMiddleware called",
			"path", c.Path(),
			"method", c.Method())

		// Get user ID and verify authentication
		userID, err := RequireAuthentication(c, store)
		if err != nil {
			log.Debug("Authentication failed",
				"path", c.Path(),
				"error", err)
			return c.Redirect("/login")
		}

		log.Debug("User authenticated",
			"user_id", userID,
			"path", c.Path())

		// Set authentication data in locals
		err = SetAuthLocals(c, store)
		if err != nil {
			log.Debug("Failed to set auth locals",
				"user_id", userID,
				"path", c.Path(),
				"error", err)
			return c.Redirect("/login")
		}

		return c.Next()
	}
}

// GuestMiddleware redirects authenticated users away from guest-only pages
func GuestMiddleware(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Next()
		}

		userID := sess.Get("user_id")
		if userID != nil {
			return c.Redirect("/")
		}

		return c.Next()
	}
}

// OptionalAuthMiddleware sets user context if authenticated but allows all users through
func OptionalAuthMiddleware(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Attempt to get user authentication info (will succeed silently if not authenticated)
		_, err := GetAuthenticatedUser(c, store)
		if err != nil {
			// Session error, but we'll allow the request to proceed anyway
			return c.Next()
		}

		// Set authentication data in locals if the user is authenticated
		// (this is a no-op if the user is not authenticated)
		SetAuthLocals(c, store)

		return c.Next()
	}
}
