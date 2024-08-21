package handlers

import "github.com/AlyxPink/meower/internal/pkg/web/routes"

func (server *Server) SetupPublicRoutes() {
	server.Static(routes.StaticPath, "./internal/pkg/web/static/public")

	// Add handlers
	NewMeow(server)
	NewHomepage(server)
}
