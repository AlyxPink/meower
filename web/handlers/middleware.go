package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// AuthMiddleware checks if the user is authenticated
func AuthMiddleware(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Printf("DEBUG: AuthMiddleware called for path: %s\n", c.Path())

		sess, err := store.Get(c)
		if err != nil {
			fmt.Printf("DEBUG: Session error in AuthMiddleware: %v\n", err)
			return c.Redirect("/login")
		}

		userID := sess.Get("user_id")
		if userID == nil {
			fmt.Printf("DEBUG: No user_id in session, redirecting to login\n")
			return c.Redirect("/login")
		}

		fmt.Printf("DEBUG: User authenticated: %v\n", userID)

		// Add user info to locals for use in handlers
		c.Locals("user_id", userID)
		c.Locals("username", sess.Get("username"))
		c.Locals("display_name", sess.Get("display_name"))

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
