package urls

// This file defines the application's routes. Each route is a struct that
// implements URLer (URL/Pattern/Name). Add your own routes here following the
// two patterns below.
//
// Static routes (no path parameters) are the simplest — URL and Pattern are the
// same literal string:
//
//	type Things struct{}
//	func (Things) URL() string     { return "/things" }
//	func (Things) Pattern() string { return "/things" }
//	func (Things) Name() string    { return "things.index" }
//
// Parameterized routes carry path parameters as struct fields tagged with
// `param:"..."`. They define a route() once and derive both URL() (concrete
// values) and Pattern() (Fiber :param placeholders) from it, so the two can
// never drift. See Meow below for the full pattern.

// ─── Homepage ────────────────────────────────────────────────────────────────

// Homepage is the application root.
type Homepage struct{}

func (Homepage) URL() string     { return "/" }
func (Homepage) Pattern() string { return "/" }
func (Homepage) Name() string    { return "home.index" }

// ─── Authentication ──────────────────────────────────────────────────────────

// LoginShow renders the login form.
type LoginShow struct{}

func (LoginShow) URL() string     { return "/login" }
func (LoginShow) Pattern() string { return "/login" }
func (LoginShow) Name() string    { return "auth.login.show" }

// Login submits the login form.
type Login struct{}

func (Login) URL() string     { return "/login" }
func (Login) Pattern() string { return "/login" }
func (Login) Name() string    { return "auth.login" }

// SignupShow renders the signup form.
type SignupShow struct{}

func (SignupShow) URL() string     { return "/signup" }
func (SignupShow) Pattern() string { return "/signup" }
func (SignupShow) Name() string    { return "auth.signup.show" }

// Signup submits the signup form.
type Signup struct{}

func (Signup) URL() string     { return "/signup" }
func (Signup) Pattern() string { return "/signup" }
func (Signup) Name() string    { return "auth.signup" }

// Logout ends the session.
type Logout struct{}

func (Logout) URL() string     { return "/logout" }
func (Logout) Pattern() string { return "/logout" }
func (Logout) Name() string    { return "auth.logout" }

// ─── Meows (example RESTful resource) ────────────────────────────────────────
//
// These routes model a conventional REST resource and double as a worked
// example of the parameterized-route pattern. Replace them with your own
// resources.

// MeowIndex lists meows — GET /meows.
type MeowIndex struct{}

func (MeowIndex) URL() string     { return "/meows" }
func (MeowIndex) Pattern() string { return "/meows" }
func (MeowIndex) Name() string    { return "meows.index" }

// MeowNew renders the new-meow form — GET /meows/new.
type MeowNew struct{}

func (MeowNew) URL() string     { return "/meows/new" }
func (MeowNew) Pattern() string { return "/meows/new" }
func (MeowNew) Name() string    { return "meows.new" }

// MeowCreate creates a meow — POST /meows.
type MeowCreate struct{}

func (MeowCreate) URL() string     { return "/meows" }
func (MeowCreate) Pattern() string { return "/meows" }
func (MeowCreate) Name() string    { return "meows.create" }

// Meow shows a single meow — GET /meows/:id.
//
// This is the canonical parameterized-route pattern: one route() definition
// feeds both URL() and Pattern(). The Pattern() is computed once at package
// init via the cached _meow value, so reflection runs only at startup.
type Meow struct {
	ID string `param:"id"`
}

func (m *Meow) route() *Route {
	return R(
		"meows.show",
		Lit("meows"), Param(&m.ID),
	)
}

func (m Meow) URL() string { return (&m).route().Build() }

var _meow = func() Compiled {
	m := &Meow{}
	return m.route().Compile(m)
}()

func (Meow) Pattern() string { return _meow.Pattern }
func (Meow) Name() string    { return _meow.Name }

// MeowEdit renders the edit form for a meow — GET /meows/:id/edit.
type MeowEdit struct {
	ID string `param:"id"`
}

func (m *MeowEdit) route() *Route {
	return R(
		"meows.edit",
		Lit("meows"), Param(&m.ID), Lit("edit"),
	)
}

func (m MeowEdit) URL() string { return (&m).route().Build() }

var _meowEdit = func() Compiled {
	m := &MeowEdit{}
	return m.route().Compile(m)
}()

func (MeowEdit) Pattern() string { return _meowEdit.Pattern }
func (MeowEdit) Name() string    { return _meowEdit.Name }
