package web

func (server *Server) SetupPublicRoutes() {
	// Serve static files
	server.Static("/static", "./web/static/public")

	// Add handlers
	NewMeow(server)
	NewHomepage(server)
}
