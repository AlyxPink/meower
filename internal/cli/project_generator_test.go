package cli

import (
	"os"
	"testing"
)

func TestProjectGenerator_Generate(t *testing.T) {
	// Skip this test as it requires embedded files from the CLI binary
	// Use integration tests instead for full end-to-end testing
	t.Skip("Skipping unit test - use integration test for full project generation testing")
}

func TestProjectGenerator_ValidateAndPrepare(t *testing.T) {
	tests := []struct {
		name        string
		config      *ProjectConfig
		expectError bool
	}{
		{
			name: "valid project config",
			config: &ProjectConfig{
				ProjectName: "valid-project",
				ModulePath:  "github.com/user/valid-project",
				Force:       false,
			},
			expectError: false,
		},
		{
			name: "invalid project name - empty",
			config: &ProjectConfig{
				ProjectName: "",
				ModulePath:  "github.com/user/project",
				Force:       false,
			},
			expectError: true,
		},
		{
			name: "invalid project name - starts with hyphen",
			config: &ProjectConfig{
				ProjectName: "-invalid",
				ModulePath:  "github.com/user/invalid",
				Force:       false,
			},
			expectError: true,
		},
		{
			name: "invalid project name - contains spaces",
			config: &ProjectConfig{
				ProjectName: "invalid project",
				ModulePath:  "github.com/user/project",
				Force:       false,
			},
			expectError: true,
		},
		{
			name: "auto-generated module path",
			config: &ProjectConfig{
				ProjectName: "auto-module",
				ModulePath:  "", // Should be auto-generated
				Force:       false,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewProjectGenerator(tt.config)
			err := generator.ValidateAndPrepare()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Test auto-generated module path
			if tt.config.ModulePath == "" && !tt.expectError {
				expectedPath := "github.com/user/" + tt.config.ProjectName
				if tt.config.ModulePath != expectedPath {
					t.Errorf("Expected auto-generated module path %s, got %s",
						expectedPath, tt.config.ModulePath)
				}
			}
		})
	}
}

func TestProjectGenerator_DirectoryExists(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to change back to original directory: %v", err)
		}
	}()

	// Create a directory that already exists
	existingDir := "existing-project"
	if err := os.MkdirAll(existingDir, 0o755); err != nil {
		t.Fatalf("Failed to create existing directory: %v", err)
	}

	// Test without force flag - should fail
	config := &ProjectConfig{
		ProjectName: existingDir,
		ModulePath:  "github.com/test/existing",
		Force:       false,
	}

	generator := NewProjectGenerator(config)
	err = generator.ValidateAndPrepare()
	if err == nil {
		t.Error("Expected error when directory exists and force=false, but got none")
	}

	// Test with force flag - should succeed
	config.Force = true
	err = generator.ValidateAndPrepare()
	if err != nil {
		t.Errorf("Expected no error when force=true, but got: %v", err)
	}
}

func TestProjectGenerator_FileCount(t *testing.T) {
	// Skip this test as it requires embedded files from the CLI binary
	// Use integration tests instead for full end-to-end testing
	t.Skip("Skipping unit test - use integration test for file count testing")
}

// Benchmark test to ensure performance doesn't regress
func BenchmarkProjectGeneration(b *testing.B) {
	// Skip this benchmark as it requires embedded files from the CLI binary
	// Use integration benchmarks instead for full project generation testing
	b.Skip("Skipping unit benchmark - use integration benchmark for full project generation testing")
}
