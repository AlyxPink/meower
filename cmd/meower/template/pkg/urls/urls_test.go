package urls

import (
	"reflect"
	"strings"
	"testing"
)

// allRoutes is the canonical list of every route in the package. Tests that
// need to sweep all routes (name uniqueness, interface conformance) use it, so
// adding a route here is the only bookkeeping a new route requires.
func allRoutes() []URLer {
	return []URLer{
		Homepage{},
		LoginShow{},
		Login{},
		SignupShow{},
		Signup{},
		Logout{},
		MeowIndex{},
		MeowNew{},
		MeowCreate{},
		Meow{},
		MeowEdit{},
		SSEStream{},
		SSEHealth{},
	}
}

// TestURLPatternConsistency verifies that, for a parameterized route, swapping
// each value in the built URL for its :param placeholder yields the Pattern.
// This is the core invariant of the kit: URL and Pattern are derived from one
// definition and cannot drift.
func TestURLPatternConsistency(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		pattern string
		values  map[string]string // built-value -> param name
	}{
		{
			name:    "Meow",
			url:     Meow{ID: "ID"}.URL(),
			pattern: Meow{}.Pattern(),
			values:  map[string]string{"ID": "id"},
		},
		{
			name:    "MeowEdit",
			url:     MeowEdit{ID: "ID"}.URL(),
			pattern: MeowEdit{}.Pattern(),
			values:  map[string]string{"ID": "id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := tt.url
			for value, param := range tt.values {
				expected = strings.ReplaceAll(expected, value, ":"+param)
			}
			if expected != tt.pattern {
				t.Errorf("URL/Pattern mismatch:\n  URL:      %s\n  Pattern:  %s\n  Expected: %s", tt.url, tt.pattern, expected)
			}
		})
	}
}

// TestStaticRoutesURLEqualsPattern verifies that routes without parameters have
// identical URL and Pattern.
func TestStaticRoutesURLEqualsPattern(t *testing.T) {
	static := []URLer{
		Homepage{},
		LoginShow{}, Login{}, SignupShow{}, Signup{}, Logout{},
		MeowIndex{}, MeowNew{}, MeowCreate{},
		SSEStream{}, SSEHealth{},
	}

	for _, route := range static {
		t.Run(route.Name(), func(t *testing.T) {
			if route.URL() != route.Pattern() {
				t.Errorf("static route URL != Pattern:\n  URL:     %s\n  Pattern: %s", route.URL(), route.Pattern())
			}
		})
	}
}

// TestURLBuildWithValues verifies concrete URLs are built correctly.
func TestURLBuildWithValues(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"Homepage", Homepage{}.URL(), "/"},
		{"Login", Login{}.URL(), "/login"},
		{"MeowIndex", MeowIndex{}.URL(), "/meows"},
		{"MeowNew", MeowNew{}.URL(), "/meows/new"},
		{"Meow", Meow{ID: "abc-123"}.URL(), "/meows/abc-123"},
		{"MeowEdit", MeowEdit{ID: "abc-123"}.URL(), "/meows/abc-123/edit"},
		{"SSEStream", SSEStream{}.URL(), "/events/stream"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.url != tt.expected {
				t.Errorf("URL mismatch:\n  got:      %s\n  expected: %s", tt.url, tt.expected)
			}
		})
	}
}

// TestPatternGeneration verifies Fiber patterns are generated correctly.
func TestPatternGeneration(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{"Homepage", Homepage{}.Pattern(), "/"},
		{"Meow", Meow{}.Pattern(), "/meows/:id"},
		{"MeowEdit", MeowEdit{}.Pattern(), "/meows/:id/edit"},
		{"MeowIndex", MeowIndex{}.Pattern(), "/meows"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pattern != tt.expected {
				t.Errorf("Pattern mismatch:\n  got:      %s\n  expected: %s", tt.pattern, tt.expected)
			}
		})
	}
}

// TestRouteNamesUnique verifies every route name is unique. Duplicate names
// would collide in Fiber's named-route table and break reverse URL lookup.
func TestRouteNamesUnique(t *testing.T) {
	seen := make(map[string]bool)
	for _, route := range allRoutes() {
		name := route.Name()
		if seen[name] {
			t.Errorf("duplicate route name: %s", name)
		}
		seen[name] = true
	}
}

// TestPatternsValidFiber verifies every pattern uses only valid :param_name
// segments (lowercase alphanumeric + underscore).
func TestPatternsValidFiber(t *testing.T) {
	for _, route := range allRoutes() {
		t.Run(route.Name(), func(t *testing.T) {
			pattern := route.Pattern()
			for seg := range strings.SplitSeq(pattern, "/") {
				if seg == "" || !strings.HasPrefix(seg, ":") {
					continue
				}
				paramName := seg[1:]
				if paramName == "" {
					t.Errorf("empty param name in pattern: %s", pattern)
				}
				for _, c := range paramName {
					if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
						t.Errorf("invalid character '%c' in param name: %s", c, pattern)
					}
				}
			}
		})
	}
}

// TestParamTagsExist verifies every string field on a parameterized route has a
// param tag, so Pattern() never falls back to a lowercase-field-name guess.
func TestParamTagsExist(t *testing.T) {
	structsToCheck := []any{
		&Meow{},
		&MeowEdit{},
	}

	for _, s := range structsToCheck {
		val := reflect.ValueOf(s).Elem()
		typ := val.Type()

		t.Run(typ.Name(), func(t *testing.T) {
			for i := 0; i < typ.NumField(); i++ {
				field := typ.Field(i)
				if field.Type.Kind() == reflect.String {
					if field.Tag.Get("param") == "" {
						t.Errorf("field %s.%s missing param tag", typ.Name(), field.Name)
					}
				}
			}
		})
	}
}

// TestURLerInterface verifies all route types implement URLer (compile-time).
func TestURLerInterface(t *testing.T) {
	for _, route := range allRoutes() {
		var _ URLer = route
	}
}

// TestRouteBuilder exercises the low-level Route builder directly.
func TestRouteBuilder(t *testing.T) {
	t.Run("nil route returns /", func(t *testing.T) {
		var r *Route
		if r.Build() != "/" {
			t.Errorf("nil route should return /, got %s", r.Build())
		}
	})

	t.Run("empty segments returns /", func(t *testing.T) {
		r := R("test")
		if r.Build() != "/" {
			t.Errorf("empty segments should return /, got %s", r.Build())
		}
	})

	t.Run("literals only", func(t *testing.T) {
		r := R("test", Lit("foo"), Lit("bar"))
		if got := r.Build(); got != "/foo/bar" {
			t.Errorf("expected /foo/bar, got %s", got)
		}
	})

	t.Run("literal and param", func(t *testing.T) {
		id := "xyz"
		r := R("test", Lit("foo"), Param(&id))
		if got := r.Build(); got != "/foo/xyz" {
			t.Errorf("expected /foo/xyz, got %s", got)
		}
	})
}

// TestEmptyParams documents that empty parameter values build empty path
// segments rather than erroring — callers are responsible for supplying values.
func TestEmptyParams(t *testing.T) {
	if got := (Meow{ID: ""}).URL(); got != "/meows/" {
		t.Errorf("empty param URL unexpected: %s", got)
	}
}

// TestParamTransform verifies the optional ParamTransform hook rewrites
// parameter values in URL() output while leaving Pattern() untouched. The hook
// is reset after the test so it does not leak into other tests.
func TestParamTransform(t *testing.T) {
	original := ParamTransform
	t.Cleanup(func() { ParamTransform = original })

	ParamTransform = func(s string) string {
		return strings.ReplaceAll(s, "-", "")
	}

	if got := (Meow{ID: "ab-cd-ef"}).URL(); got != "/meows/abcdef" {
		t.Errorf("ParamTransform not applied to URL: got %s, want /meows/abcdef", got)
	}

	// Pattern is unaffected by ParamTransform — it contains placeholders, not values.
	if got := (Meow{}).Pattern(); got != "/meows/:id" {
		t.Errorf("ParamTransform should not affect Pattern: got %s", got)
	}
}
