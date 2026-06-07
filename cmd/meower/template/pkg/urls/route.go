// Package urls provides type-safe URL building for the application.
// It is the single source of truth for URL patterns, shared across the web,
// api, and any other component that needs to build or register routes.
//
// Key properties:
//   - Compile-time safety: routes are structs, so a renamed field or route is
//     a build error, not a silent 404.
//   - One definition, two outputs: the same route definition produces both the
//     concrete URL (via Build) and the Fiber registration pattern (via Compile),
//     so they can never drift out of sync.
//
// Example:
//
//	// Build a URL with values
//	u := Meow{ID: "abc"}.URL()       // "/meows/abc"
//
//	// Get the Fiber pattern for route registration
//	p := Meow{}.Pattern()            // "/meows/:id"
package urls

import (
	"reflect"
	"strings"
	"unsafe"
)

// ParamTransform, when set, is applied to every parameter value as a URL is
// built. It is the extension point for projects that normalize identifiers in
// their URLs (for example, stripping dashes from UUIDs). It is nil by default,
// so parameter values are emitted verbatim.
//
// Set it once at startup, before any URLs are built:
//
//	urls.ParamTransform = func(s string) string {
//	    return strings.ReplaceAll(s, "-", "")
//	}
//
// ParamTransform only affects URL() / Build() output. Pattern() is unaffected,
// since patterns contain :param placeholders rather than concrete values.
var ParamTransform func(string) string

// applyParamTransform runs the configured ParamTransform if one is set,
// otherwise returns the value unchanged.
func applyParamTransform(s string) string {
	if ParamTransform != nil {
		return ParamTransform(s)
	}
	return s
}

// segment represents a path segment - either a literal string or a parameter
// pointer.
type segment struct {
	literal  string  // for Lit() - static path segment
	fieldPtr *string // for Param() - pointer to struct field
}

// Lit creates a literal path segment (e.g., "meows", "edit").
func Lit(s string) segment {
	return segment{literal: s}
}

// Param creates a parameter segment from a struct field pointer.
// The pointer is used to:
//  1. Read the value when building URLs
//  2. Find the field name via reflection for pattern generation
func Param(p *string) segment {
	return segment{fieldPtr: p}
}

// Route defines a URL structure with a name and path segments.
type Route struct {
	name     string
	segments []segment
}

// R creates a new Route with the given name and segments.
func R(name string, segs ...segment) *Route {
	return &Route{name: name, segments: segs}
}

// Build constructs a URL by joining literal segments and dereferencing
// parameter pointers. Parameter values pass through ParamTransform if one is
// configured.
func (r *Route) Build() string {
	if r == nil || len(r.segments) == 0 {
		return "/"
	}

	var parts []string
	for _, seg := range r.segments {
		if seg.literal != "" {
			parts = append(parts, seg.literal)
		} else if seg.fieldPtr != nil {
			parts = append(parts, applyParamTransform(*seg.fieldPtr))
		}
	}

	return "/" + strings.Join(parts, "/")
}

// Compiled holds the pre-computed route name and Fiber pattern.
type Compiled struct {
	Name    string
	Pattern string
}

// Compile generates the Fiber pattern via reflection. It finds each parameter
// pointer's corresponding field and reads its `param` tag.
func (r *Route) Compile(structPtr any) Compiled {
	if r == nil {
		return Compiled{}
	}

	val := reflect.ValueOf(structPtr)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	typ := val.Type()

	var parts []string
	for _, seg := range r.segments {
		if seg.literal != "" {
			parts = append(parts, seg.literal)
		} else if seg.fieldPtr != nil {
			paramName := findParamName(val, typ, seg.fieldPtr)
			if paramName != "" {
				parts = append(parts, ":"+paramName)
			}
		}
	}

	pattern := "/"
	if len(parts) > 0 {
		pattern = "/" + strings.Join(parts, "/")
	}

	return Compiled{
		Name:    r.name,
		Pattern: pattern,
	}
}

// findParamName finds the param tag for a field given its pointer. It compares
// pointer addresses to identify which field the pointer belongs to.
func findParamName(val reflect.Value, typ reflect.Type, ptr *string) string {
	ptrAddr := uintptr(unsafe.Pointer(ptr))

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() != reflect.String || !field.CanAddr() {
			continue
		}

		fieldAddr := field.Addr().Pointer()
		if fieldAddr == ptrAddr {
			// Found the field, get its param tag
			structField := typ.Field(i)
			tag := structField.Tag.Get("param")
			if tag == "" {
				// Fall back to lowercase field name if no tag
				return strings.ToLower(structField.Name)
			}
			// Handle tags like "id,omitempty"
			if before, _, ok := strings.Cut(tag, ","); ok {
				return before
			}
			return tag
		}
	}

	return ""
}

// URLer is the interface implemented by all route types.
type URLer interface {
	URL() string
	Pattern() string
	Name() string
}
