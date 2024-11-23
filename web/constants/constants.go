package constants

import (
	"github.com/AlyxPink/meower/web/grpc"
	"github.com/gofiber/fiber/v2"
)

type route struct {
	Name string
	Path string
}

type routes struct {
	// Homepage
	HomeIndex route

	// Meower
	MeowCreate route
	IndexMeow  route
	MeowNew    route
}

type Server struct {
	Web        *fiber.App
	GrpcClient *grpc.Client
}

/*
Routes for a common resource using RESTful conventions.

| HTTP Verb | URL              | Route name    | Description                                  |
| --------- | ---------------- | ------------- | -------------------------------------------- |
| GET       | /photos          | index         | Display a list of all photos                 |
| GET       | /photos/new      | new           | Return an HTML form for creating a new photo |
| POST      | /photos          | create        | Create a new photo                           |
| GET       | /photos/:id      | show          | Display a specific photo                     |
| GET       | /photos/:id/edit | edit          | Return an HTML form for editing a photo      |
| PATCH/PUT | /photos/:id      | update        | Update a specific photo                      |
| DELETE    | /photos/:id      | destroy       | Delete a specific photo                      |

Feel free to use anything else that makes sense for your endpoints, resources and application.
*/
var Routes = routes{
	// Homepage
	HomeIndex: route{Name: "home.index", Path: "/"},

	// Meower
	MeowCreate: route{Name: "meow.create", Path: "/meows/"},
	IndexMeow:  route{Name: "meow.index", Path: "/meows/"},
	MeowNew:    route{Name: "meow.new", Path: "/meows/new"},
}
