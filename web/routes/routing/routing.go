package routing

import (
	"github.com/AlyxPink/meower/web/handlers"
	"github.com/AlyxPink/meower/web/routes"
)

func RegisterRoutes(app *handlers.App) {
	// Static
	app.Web.Static("/static", "/src/web/static/public/").Name("static")

	// Homepage
	homepage := handlers.Homepage{App: app}
	app.Web.Get(routes.Homepage.Path, homepage.Index).Name(routes.Homepage.Name)

	// Meower
	meower := handlers.Meower{App: app}
	app.Web.Get(routes.MeowIndex.Path, meower.Index).Name(routes.MeowIndex.Name)
	app.Web.Get(routes.MeowNew.Path, meower.New).Name(routes.MeowNew.Name)
	app.Web.Post(routes.MeowCreate.Path, meower.Create).Name(routes.MeowCreate.Name)
}
