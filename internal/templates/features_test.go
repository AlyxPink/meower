package templates

import "testing"

func TestSkipForFeatures(t *testing.T) {
	tests := []struct {
		name         string
		features     Features
		path         string
		wantSkip     bool
		wantRenameTo string
	}{
		// Auth feature gating.
		{"auth on keeps auth handler", Features{Auth: true, Workers: true}, "web/handlers/auth.go", false, ""},
		{"auth off skips auth handler", Features{Auth: false, Workers: true}, "web/handlers/auth.go", true, ""},
		{"auth off skips login view", Features{Auth: false, Workers: true}, "web/views/login.templ", true, ""},
		{"auth on keeps login view", Features{Auth: true, Workers: true}, "web/views/login.templ", false, ""},

		// Auth variant selection.
		{"auth on keeps canonical routing", Features{Auth: true, Workers: true}, "web/routing/routing.go", false, ""},
		{"auth on drops noauth variant", Features{Auth: true, Workers: true}, "web/routing/routing.noauth.go", true, ""},
		{"auth off drops canonical routing", Features{Auth: false, Workers: true}, "web/routing/routing.go", true, ""},
		{"auth off renames noauth variant", Features{Auth: false, Workers: true}, "web/routing/routing.noauth.go", false, "web/routing/routing.go"},

		// Workers feature gating.
		{"workers on keeps harness", Features{Auth: true, Workers: true}, "api/server/workers/worker.go", false, ""},
		{"workers off skips harness dir", Features{Auth: true, Workers: false}, "api/server/workers/worker.go", true, ""},

		// Workers variant selection.
		{"workers off renames noop wiring", Features{Auth: true, Workers: false}, "api/server/workers_wiring.noworkers.go", false, "api/server/workers_wiring.go"},
		{"workers off drops canonical wiring", Features{Auth: true, Workers: false}, "api/server/workers_wiring.go", true, ""},
		{"workers on drops noop wiring", Features{Auth: true, Workers: true}, "api/server/workers_wiring.noworkers.go", true, ""},

		// Prefix safety: a sibling file with a shared name prefix is not skipped.
		{"auth_utils not matched by auth.go prefix", Features{Auth: true, Workers: true}, "web/handlers/auth_utils.go", false, ""},

		// Unrelated files always pass through.
		{"unrelated file untouched", Features{Auth: false, Workers: false}, "web/main.go", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skip, renameTo := tt.features.skipForFeatures(tt.path)
			if skip != tt.wantSkip {
				t.Errorf("skip = %v, want %v", skip, tt.wantSkip)
			}
			if renameTo != tt.wantRenameTo {
				t.Errorf("renameTo = %q, want %q", renameTo, tt.wantRenameTo)
			}
		})
	}
}
