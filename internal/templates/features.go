package templates

import (
	"path/filepath"
	"strings"
)

// Features controls which optional parts of the template are included in a
// generated project. Each feature defaults to enabled (batteries-included); the
// CLI's --no-* flags disable them, which removes the corresponding files
// entirely rather than leaving dead code behind.
type Features struct {
	Auth    bool
	Workers bool
}

// DefaultFeatures returns the batteries-included default (everything on).
func DefaultFeatures() Features {
	return Features{Auth: true, Workers: true}
}

// featurePaths lists the template path prefixes (relative to the template root,
// i.e. after the "template/" or "cmd/meower/template/" prefix is stripped) that
// belong exclusively to a feature. When the feature is disabled, every file
// under these prefixes is skipped.
//
// Paths here are the cleaned, in-project paths — e.g. "web/handlers/auth.go",
// not "cmd/meower/template/web/handlers/auth.go".
var featurePaths = map[string][]string{
	"auth": {
		"web/handlers/auth.go",
		"web/handlers/auth_utils.go",
		"web/handlers/middleware.go",
		"web/handlers/debug.go",
		"web/views/login.templ",
		"web/views/signup.templ",
	},
	"workers": {
		"api/server/workers",
	},
}

// variantSuffix is inserted before the file extension on the no-feature variant
// of a shared file. For example, with auth disabled the processor uses
// "web/routing.noauth.go.template" in place of "web/routing/routing.go".
//
// variants maps the canonical (feature-enabled) cleaned path to its
// no-feature variant's cleaned path. When the feature is OFF, the canonical
// file is skipped and the variant is renamed into its place. When the feature
// is ON, the variant file is skipped.
var authVariants = map[string]string{
	"web/routing/routing.go": "web/routing/routing.noauth.go",
}

// workersVariants maps the canonical (workers-enabled) file to its no-workers
// variant, selected the same way as authVariants.
var workersVariants = map[string]string{
	"api/server/workers_wiring.go": "api/server/workers_wiring.noworkers.go",
}

// skipForFeatures reports whether a cleaned template path should be skipped
// given the enabled features, and — when a variant file is being used — the
// cleaned path it should be written to instead (renaming the variant into the
// canonical file's place). The returned bool is true when the file must be
// skipped entirely.
func (f Features) skipForFeatures(cleanPath string) (skip bool, renameTo string) {
	// 1. Whole-feature directories/files.
	if !f.Auth && pathHasPrefix(cleanPath, featurePaths["auth"]) {
		return true, ""
	}
	if !f.Workers && pathHasPrefix(cleanPath, featurePaths["workers"]) {
		return true, ""
	}

	// 2. Variant files. For each feature, when it's ON we skip the no-feature
	//    variant; when it's OFF we skip the canonical file and rename the
	//    variant into its place.
	if skip, renameTo, matched := applyVariants(authVariants, f.Auth, cleanPath); matched {
		return skip, renameTo
	}
	if skip, renameTo, matched := applyVariants(workersVariants, f.Workers, cleanPath); matched {
		return skip, renameTo
	}

	return false, ""
}

// applyVariants resolves a single variant map against the feature's enabled
// state. matched is true when cleanPath is one of the map's canonical or variant
// files (so the caller stops checking other maps).
func applyVariants(variants map[string]string, enabled bool, cleanPath string) (skip bool, renameTo string, matched bool) {
	for canonical, variant := range variants {
		switch {
		case cleanPath == variant && enabled:
			return true, "", true // feature on → drop the no-feature variant
		case cleanPath == canonical && !enabled:
			return true, "", true // feature off → drop the canonical file
		case cleanPath == variant && !enabled:
			return false, canonical, true // feature off → rename variant into place
		case cleanPath == canonical && enabled:
			return false, "", true // feature on → keep canonical as-is
		}
	}
	return false, "", false
}

// pathHasPrefix reports whether cleanPath equals or is nested under any of the
// given prefixes (path-segment aware, so "web/handlers/auth.go" does not match
// the prefix "web/handlers/auth_utils.go").
func pathHasPrefix(cleanPath string, prefixes []string) bool {
	for _, p := range prefixes {
		if cleanPath == p {
			return true
		}
		if strings.HasPrefix(cleanPath, p+string(filepath.Separator)) || strings.HasPrefix(cleanPath, p+"/") {
			return true
		}
	}
	return false
}
