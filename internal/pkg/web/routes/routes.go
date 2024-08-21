package routes

const (
	/* Routes for a common resource using RESTful conventions.

	| HTTP Verb | URL              | Function name | Description                                  |
	| --------- | ---------------- | ------------- | -------------------------------------------- |
	| GET       | /photos          | index         | display a list of all photos                 |
	| GET       | /photos/new      | new           | return an HTML form for creating a new photo |
	| POST      | /photos          | create        | create a new photo                           |
	| GET       | /photos/:id      | show          | display a specific photo                     |
	| GET       | /photos/:id/edit | edit          | return an HTML form for editing a photo      |
	| PATCH/PUT | /photos/:id      | update        | update a specific photo                      |
	| DELETE    | /photos/:id      | destroy       | delete a specific photo                      |

	Feel free to use anything else that makes sense for your endpoints, resources and application. */

	// Static routes
	StaticPath = "/static"

	// Homepage
	HomepageIndex = "/"

	// Project routes
	MeowIndex  = "/meow/"
	MeowNew    = "/meow/new"
	MeowCreate = "/meow/"
)
