package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIProjectGeneration tests the complete CLI project generation
// This is an integration test that builds and runs the actual CLI binary
func TestCLIProjectGeneration(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the CLI binary
	cliPath := buildCLIBinary(t)
	defer os.Remove(cliPath)

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Change to temp directory for test
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer func() {
		os.Chdir(originalDir)
	}()

	// Test project generation
	projectName := "cli-test-project"
	modulePath := "github.com/test/cli-test-project"

	cmd := exec.Command(cliPath, "new", projectName, "--module", modulePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, string(output))
	}

	// Verify the output contains expected messages
	outputStr := string(output)
	expectedMessages := []string{
		"Creating new Meower project",
		"Project: " + projectName,
		"Module: " + modulePath,
		"Copying project structure",
		"Project created successfully",
	}

	for _, expectedMsg := range expectedMessages {
		if !strings.Contains(outputStr, expectedMsg) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nFull output: %s", expectedMsg, outputStr)
		}
	}

	// Verify project directory was created
	projectDir := filepath.Join(tempDir, projectName)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Errorf("Project directory was not created: %s", projectDir)
	}

	// Test core files exist
	coreFiles := []string{
		".meowed",
		"go.mod",
		"docker-compose.yml",
		"README.md",
	}

	for _, file := range coreFiles {
		filePath := filepath.Join(projectDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected core file not found: %s", file)
		}
	}

	// Verify go.mod contains correct module path
	goModPath := filepath.Join(projectDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Errorf("Failed to read go.mod: %v", err)
	} else {
		if !strings.Contains(string(content), "module "+modulePath) {
			t.Errorf("go.mod does not contain correct module path.\nExpected: module %s\nContent: %s", 
				modulePath, string(content))
		}
	}

	// Count generated files
	fileCount := countFiles(t, projectDir)
	if fileCount < 50 { // Expect at least 50 files for a real project
		t.Errorf("Expected at least 50 files, but got %d", fileCount)
	}

	t.Logf("✅ CLI integration test passed. Generated %d files", fileCount)
}

// TestCLIValidation tests CLI validation through the actual binary
func TestCLIValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cliPath := buildCLIBinary(t)
	defer os.Remove(cliPath)

	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer func() {
		os.Chdir(originalDir)
	}()

	// Test invalid project name
	cmd := exec.Command(cliPath, "new", "Invalid-Project-Name", "--module", "github.com/test/invalid")
	output, err := cmd.CombinedOutput()
	
	// Command should fail with validation error
	if err == nil {
		t.Error("Expected CLI to fail with invalid project name, but it succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Project generation failed") {
		t.Errorf("Expected validation error in output, got: %s", outputStr)
	}

	// Test directory already exists (without force flag)
	validProject := "valid-project"
	
	// Create the project first
	cmd = exec.Command(cliPath, "new", validProject, "--module", "github.com/test/valid")
	_, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create initial project: %v", err)
	}

	// Try to create it again (should fail)
	cmd = exec.Command(cliPath, "new", validProject, "--module", "github.com/test/valid")
	output, err = cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected CLI to fail when directory exists, but it succeeded")
	}

	outputStr = string(output)
	if !strings.Contains(outputStr, "directory already exists") {
		t.Errorf("Expected 'directory already exists' error, got: %s", outputStr)
	}

	// Test with force flag (should succeed)
	cmd = exec.Command(cliPath, "new", validProject, "--module", "github.com/test/valid", "--force")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Expected CLI to succeed with --force flag, but got error: %v\nOutput: %s", 
			err, string(output))
	}
}

// TestCLIHelp tests the help functionality
func TestCLIHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cliPath := buildCLIBinary(t)
	defer os.Remove(cliPath)

	// Test main help
	cmd := exec.Command(cliPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI help command failed: %v", err)
	}

	outputStr := string(output)
	expectedInHelp := []string{
		"Meower Framework",
		"GoFiber web server with gRPC API",
		"Available Commands:",
		"new",
		"create",
	}

	for _, expected := range expectedInHelp {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected help to contain '%s', but it didn't.\nFull output: %s", expected, outputStr)
		}
	}

	// Test new command help
	cmd = exec.Command(cliPath, "new", "--help")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI new help command failed: %v", err)
	}

	outputStr = string(output)
	expectedInNewHelp := []string{
		"Create New Project",
		"project-name",
		"--module",
		"--force",
	}

	for _, expected := range expectedInNewHelp {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected new help to contain '%s', but it didn't.\nFull output: %s", expected, outputStr)
		}
	}
}

// buildCLIBinary builds the CLI binary for testing
func buildCLIBinary(t *testing.T) string {
	// Get the project root directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate to project root (assuming we're in internal/cli)
	projectRoot := filepath.Join(wd, "..", "..")
	
	// Create temporary binary
	tempBinary := filepath.Join(t.TempDir(), "meower-test")
	
	// Build the CLI
	cmd := exec.Command("go", "build", "-o", tempBinary, "./cmd/meower")
	cmd.Dir = projectRoot
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build CLI binary: %v\nOutput: %s", err, string(output))
	}

	return tempBinary
}

// countFiles counts the number of files in a directory recursively
func countFiles(t *testing.T, dir string) int {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	
	if err != nil {
		t.Fatalf("Failed to count files in %s: %v", dir, err)
	}
	
	return count
}