package routes

const (
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

	// Homepage
	HomeIndex = "home.index"

	// Meower
	MeowIndex  = "meow.index"
	MeowNew    = "meow.new"
	MeowCreate = "meow.create"
)
