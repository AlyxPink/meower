package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	userV1 "github.com/AlyxPink/meower/api/proto/user/v1"
	"meower-template-web/routes"
	"meower-template-web/views/pages/auth"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth struct {
	App *App
}

// ShowLogin displays the login form
func (a *Auth) ShowLogin(c *fiber.Ctx) error {
	// Check if user is already logged in
	sess, err := a.App.SessionStore.Get(c)
	if err != nil {
		return err
	}

	if sess.Get("user_id") != nil {
		return c.Redirect(routes.MeowIndex.Path)
	}

	return renderTempl(c, auth.Login(c, ""))
}

// Login handles the login form submission
func (a *Auth) Login(c *fiber.Ctx) error {
	usernameOrEmail := c.FormValue("username_or_email")
	password := c.FormValue("password")

	if usernameOrEmail == "" || password == "" {
		return renderTempl(c, auth.Login(c, "Username/Email and password are required"))
	}

	// Call the gRPC API
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.App.API.UserService.Login(ctx, &userV1.LoginRequest{
		UsernameOrEmail: usernameOrEmail,
		Password:        password,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return renderTempl(c, auth.Login(c, "An error occurred"))
		}

		var errorMsg string
		switch st.Code() {
		case codes.Unauthenticated:
			errorMsg = "Invalid credentials"
		case codes.PermissionDenied:
			errorMsg = "Account is locked"
		default:
			errorMsg = "An error occurred"
		}

		return renderTempl(c, auth.Login(c, errorMsg))
	}

	// Store user session
	sess, err := a.App.SessionStore.Get(c)
	if err != nil {
		return err
	}

	sess.Set("user_id", resp.User.Id)
	sess.Set("username", resp.User.Username)
	sess.Set("display_name", resp.User.DisplayName)
	sess.Set("token", resp.Token)

	if err := sess.Save(); err != nil {
		return err
	}

	// Debug: Log successful login
	fmt.Printf("DEBUG: User %s logged in successfully, redirecting to /\n", resp.User.Username)

	return c.Redirect(routes.MeowIndex.Path)
}

// ShowSignup displays the signup form
func (a *Auth) ShowSignup(c *fiber.Ctx) error {
	// Check if user is already logged in
	sess, err := a.App.SessionStore.Get(c)
	if err != nil {
		return err
	}

	if sess.Get("user_id") != nil {
		return c.Redirect(routes.MeowIndex.Path)
	}

	return renderTempl(c, auth.Signup(c, ""))
}

// Signup handles the signup form submission
func (a *Auth) Signup(c *fiber.Ctx) error {
	username := c.FormValue("username")
	displayName := c.FormValue("display_name")
	email := c.FormValue("email")
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")

	// Basic validation
	if username == "" || displayName == "" || email == "" || password == "" {
		return renderTempl(c, auth.Signup(c, "All fields are required"))
	}

	if password != confirmPassword {
		return renderTempl(c, auth.Signup(c, "Passwords do not match"))
	}

	if len(password) < 8 {
		return renderTempl(c, auth.Signup(c, "Password must be at least 8 characters long"))
	}

	if !strings.Contains(email, "@") {
		return renderTempl(c, auth.Signup(c, "Please enter a valid email address"))
	}

	// Call the gRPC API
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.App.API.UserService.CreateUser(ctx, &userV1.CreateUserRequest{
		Username:    username,
		DisplayName: displayName,
		Email:       email,
		Password:    password,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return renderTempl(c, auth.Signup(c, "An error occurred"))
		}

		var errorMsg string
		switch st.Code() {
		case codes.AlreadyExists:
			if strings.Contains(st.Message(), "username") {
				errorMsg = "Username already exists"
			} else if strings.Contains(st.Message(), "email") {
				errorMsg = "Email already exists"
			} else {
				errorMsg = "User already exists"
			}
		default:
			errorMsg = "An error occurred"
		}

		return renderTempl(c, auth.Signup(c, errorMsg))
	}

	// Auto-login after successful signup
	sess, err := a.App.SessionStore.Get(c)
	if err != nil {
		return err
	}

	sess.Set("user_id", resp.User.Id)
	sess.Set("username", resp.User.Username)
	sess.Set("display_name", resp.User.DisplayName)

	if err := sess.Save(); err != nil {
		return err
	}

	return c.Redirect(routes.MeowIndex.Path)
}

// Logout handles user logout
func (a *Auth) Logout(c *fiber.Ctx) error {
	sess, err := a.App.SessionStore.Get(c)
	if err != nil {
		return err
	}

	// Call the gRPC API
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = a.App.API.UserService.Logout(ctx, &userV1.LogoutRequest{})
	if err != nil {
		// Log the error but don't fail the logout
		// In a real app, you might want to handle this differently
	}

	// Clear the session
	if err := sess.Destroy(); err != nil {
		return err
	}

	return c.Redirect(routes.Login.Path)
}
