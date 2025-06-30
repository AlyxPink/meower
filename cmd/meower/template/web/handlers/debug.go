package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// Debug handler to check session state
func (a *Auth) Debug(c *fiber.Ctx) error {
	sess, err := a.App.SessionStore.Get(c)
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Session error: %v", err))
	}

	userID := sess.Get("user_id")
	username := sess.Get("username")
	displayName := sess.Get("display_name")
	token := sess.Get("token")

	return c.JSON(fiber.Map{
		"user_id":      userID,
		"username":     username,
		"display_name": displayName,
		"token":        token,
		"session_id":   sess.ID(),
	})
}
