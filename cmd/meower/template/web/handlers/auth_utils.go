// Package handlers provides HTTP request handlers for the web application.
package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// Authentication errors
var (
	// ErrSessionError is returned when there's an error retrieving the session
	ErrSessionError = errors.New("session error")

	// ErrNotAuthenticated is returned when the user is not authenticated
	ErrNotAuthenticated = errors.New("not authenticated")
)

// GetAuthenticatedUser retrieves the user ID from session if authenticated.
//
// This function is ideal for routes that work with both authenticated and
// unauthenticated users, where you want to provide additional functionality
// for authenticated users.
//
// Parameters:
//   - c: The Fiber context for the current request
//   - store: The session store to use for authentication
//
// Returns:
//   - string: User ID if authenticated, empty string if not authenticated
//   - error: ErrSessionError if there's a problem with the session, nil otherwise
//
// Usage example:
//
//	userID, err := GetAuthenticatedUser(c, h.SessionStore)
//	if err != nil {
//	    return c.Status(fiber.StatusInternalServerError).SendString("Session error")
//	}
//
//	// If user is authenticated, userID will be non-empty
//	if userID != "" {
//	    // Do something with the authenticated user
//	} else {
//	    // Handle unauthenticated user
//	}
func GetAuthenticatedUser(c *fiber.Ctx, store *session.Store) (string, error) {
	session, err := store.Get(c)
	if err != nil {
		return "", ErrSessionError
	}

	// Get user ID from session (if authenticated)
	userIDHex := session.Get("user_id")
	if userIDHex == nil {
		return "", nil
	}

	return userIDHex.(string), nil
}

// RequireAuthentication checks if user is authenticated and returns user ID.
//
// This function is ideal for routes that should only be accessible to authenticated users.
// It checks if the user is authenticated and returns an error if they are not,
// which can be handled to return an appropriate response or redirect.
//
// Parameters:
//   - c: The Fiber context for the current request
//   - store: The session store to use for authentication
//
// Returns:
//   - string: User ID if authenticated
//   - error: ErrSessionError if there's a problem with the session,
//     ErrNotAuthenticated if the user is not authenticated
//
// Usage example:
//
//	userID, err := RequireAuthentication(c, h.SessionStore)
//	if err != nil {
//	    return HandleAuthError(c, err)
//	}
//
//	// User is authenticated, proceed with the request
//	// using userID for operations that require authentication
func RequireAuthentication(c *fiber.Ctx, store *session.Store) (string, error) {
	session, err := store.Get(c)
	if err != nil {
		return "", ErrSessionError
	}

	// Check if user is authenticated
	userIDHex := session.Get("user_id")
	if userIDHex == nil {
		return "", ErrNotAuthenticated
	}

	return userIDHex.(string), nil
}

// GetAuthenticatedUserID checks if user is authenticated and returns user ID if they are.
//
// This function is similar to RequireAuthentication but doesn't return an error if
// the user is not authenticated. It's useful for routes that work for both
// authenticated and guest users.
//
// Returns:
//   - userID: The user ID string if authenticated, empty string if not
//   - error: ErrSessionError if there's a session error, nil otherwise
//
// Usage example:
//
//	userID, err := GetAuthenticatedUserID(c, h.SessionStore)
//	if err != nil {
//	    // Handle session error
//	    return c.Status(fiber.StatusInternalServerError).SendString("Session error")
//	}
//	if userID != "" {
//	    // User is authenticated, provide personalized experience
//	} else {
//	    // User is not authenticated, provide public experience
//	}
func GetAuthenticatedUserID(c *fiber.Ctx, store *session.Store) (string, error) {
	session, err := store.Get(c)
	if err != nil {
		return "", ErrSessionError
	}

	// Check if user is authenticated
	userIDHex := session.Get("user_id")
	if userIDHex == nil {
		return "", nil // Not authenticated, but no error
	}

	return userIDHex.(string), nil
}

// SetAuthLocals sets the authentication-related locals in the fiber context.
//
// This is useful for templates that need to know about the authenticated user.
// It retrieves user data from the session and sets it in the Fiber context locals,
// which can then be accessed by templates.
//
// Parameters:
//   - c: The Fiber context for the current request
//   - store: The session store to use for authentication
//
// Returns:
//   - error: ErrSessionError if there's a problem with the session, nil otherwise
//
// Usage example:
//
//	err := SetAuthLocals(c, h.SessionStore)
//	if err != nil {
//	    return c.Status(fiber.StatusInternalServerError).SendString("Session error")
//	}
//
//	// Continue with the request, templates can now access user info via c.Locals()
func SetAuthLocals(c *fiber.Ctx, store *session.Store) error {
	session, err := store.Get(c)
	if err != nil {
		return ErrSessionError
	}

	userIDHex := session.Get("user_id")
	if userIDHex != nil {
		// Add user info to locals for use in templates
		c.Locals("user_id", userIDHex)
		c.Locals("username", session.Get("username"))
		c.Locals("display_name", session.Get("display_name"))
	}

	return nil
}

// HandleAuthError standardizes error responses for authentication errors.
//
// This function provides a consistent way to handle authentication errors
// across different handlers. It returns appropriate HTTP status codes and
// error messages based on the specific authentication error.
//
// Parameters:
//   - c: The Fiber context for the current request
//   - err: The authentication error (should be one of the defined auth errors)
//
// Returns:
//   - error: The processed Fiber response or the original error if not handled
//
// Usage example:
//
//	userID, err := RequireAuthentication(c, h.SessionStore)
//	if err != nil {
//	    return HandleAuthError(c, err)
//	}
func HandleAuthError(c *fiber.Ctx, err error) error {
	switch err {
	case ErrSessionError:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Session error",
		})
	case ErrNotAuthenticated:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	default:
		return err
	}
}

// HandleAuthErrorJSON standardizes JSON API responses for authentication errors.
//
// Similar to HandleAuthError, but formats the response using a success/message pattern
// common in JSON APIs, rather than an error format. This is particularly useful
// for AJAX endpoints that expect this response format.
//
// Parameters:
//   - c: The Fiber context for the current request
//   - err: The authentication error (should be one of the defined auth errors)
//
// Returns:
//   - error: The processed Fiber response or the original error if not handled
//
// Usage example:
//
//	userID, err := RequireAuthentication(c, h.SessionStore)
//	if err != nil {
//	    return HandleAuthErrorJSON(c, err)
//	}
func HandleAuthErrorJSON(c *fiber.Ctx, err error) error {
	switch err {
	case ErrSessionError:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Session error",
		})
	case ErrNotAuthenticated:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Not authenticated",
		})
	default:
		return err
	}
}
