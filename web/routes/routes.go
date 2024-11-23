package routes

import (
	"github.com/AlyxPink/meower/web/constants"
	"github.com/AlyxPink/meower/web/handlers"
)

func RegisterRoutes(s *constants.Server) {
	// Constants route names
	routes := constants.Routes

	// Static
	s.Web.Static("/static", "./web/static/public").Name("static")

	// Homepage
	s.Web.Get(routes.HomeIndex.Path, handlers.HomepageIndex).Name(routes.HomeIndex.Name)

	// Meower
	s.Web.Get(routes.IndexMeow.Path, handlers.MeowerIndex).Name(routes.IndexMeow.Name)
	s.Web.Get(routes.MeowNew.Path, handlers.MeowerNew).Name(routes.MeowNew.Name)
	s.Web.Post(routes.MeowCreate.Path, handlers.MeowerCreate).Name(routes.MeowCreate.Name)
}
