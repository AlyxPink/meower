package routing

import (
	"github.com/AlyxPink/meower/web/constants"
	"github.com/AlyxPink/meower/web/handlers"
	"github.com/AlyxPink/meower/web/routes"
)

func RegisterRoutes(s *constants.Server) {
	// Static
	s.Web.Static("/static", "/src/web/static/public/").Name("static")

	// Homepage
	homepage := handlers.Homepage{}
	s.Web.Get(routes.Homepage.Path, homepage.Index).Name(routes.Homepage.Name)

	// Meower
	meower := handlers.Meower{}
	s.Web.Get(routes.MeowIndex.Path, meower.Index).Name(routes.MeowIndex.Name)
	s.Web.Get(routes.MeowNew.Path, meower.New).Name(routes.MeowNew.Name)
	s.Web.Post(routes.MeowCreate.Path, meower.Create).Name(routes.MeowCreate.Name)
}
