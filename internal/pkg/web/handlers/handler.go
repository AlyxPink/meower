package handlers

func (server *Server) SetupPublicRoutes() {
	// Serve static files
	server.Static("/static", "./internal/pkg/web/static/public")

	// Add handlers
	NewMeow(server)
	NewHomepage(server)
}
