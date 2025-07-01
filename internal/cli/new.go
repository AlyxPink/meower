package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// Flags for new command
	modulePath string
	force      bool
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Create a new Meower project",
	Long: titleStyle.Render("🚀 Create New Project") + "\n\n" +
		subtitleStyle.Render("Generate a new Meower project with all the boilerplate:") + "\n" +
		subtitleStyle.Render("• GoFiber web server with gRPC API backend") + "\n" +
		subtitleStyle.Render("• PostgreSQL database with SQLC queries") + "\n" +
		subtitleStyle.Render("• Protocol Buffers for API definitions") + "\n" +
		subtitleStyle.Render("• Templ templates with TailwindCSS") + "\n" +
		subtitleStyle.Render("• Docker development environment") + "\n",
	Args: cobra.ExactArgs(1),
	RunE: runNewCommand,
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringVarP(&modulePath, "module", "m", "", "Go module path (e.g. github.com/user/project)")
	newCmd.Flags().BoolVarP(&force, "force", "f", false, "Force creation even if directory exists")
}

// implements the core project scaffolding logic using the refactored architecture
func runNewCommand(cmd *cobra.Command, args []string) error {
	// Create project configuration
	config := &ProjectConfig{
		ProjectName: args[0],
		ModulePath:  modulePath,
		Force:       force,
	}

	// Create and execute project generator
	generator := NewProjectGenerator(config)
	if err := generator.Generate(); err != nil {
		fmt.Println(errorStyle.Render("❌ Project generation failed:"), err)
		return err // Return error for proper exit codes in testing
	}

	return nil
}

func getTemplateSourceDir() (string, error) {
	// Get the directory where the CLI binary is located
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("unable to determine source location")
	}

	// Navigate up to the project root (from internal/cli to root)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))

	// For now, we'll use the current project structure as template
	// Later we can embed templates or use a separate template directory
	return projectRoot, nil
}

func cleanupGeneratedProject(projectDir string) {
	// Remove CLI-specific files that shouldn't be in generated projects
	filesToRemove := []string{
		"cmd/meower",
		"internal/cli",
		"internal/templates",
		"internal/generators",
		"CONTRIBUTING.md", // CLI development docs not needed in projects
	}

	for _, file := range filesToRemove {
		fullPath := filepath.Join(projectDir, file)
		if err := os.RemoveAll(fullPath); err != nil {
			// Log but don't fail - these files might not exist
			fmt.Printf("Warning: failed to remove %s: %v\n", fullPath, err)
		}
	}
}

// copyGuideToProject copies the GUIDE.md to the generated project
func copyGuideToProject(projectDir string) {
	// Get the source GUIDE.md path
	sourcePath := "GUIDE.md"
	destPath := filepath.Join(projectDir, "GUIDE.md")

	// Read the source file
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		// If GUIDE.md doesn't exist, silently continue
		return
	}

	// Write to destination
	if err := os.WriteFile(destPath, content, 0o644); err != nil {
		fmt.Printf("Warning: failed to write GUIDE.md: %v\n", err)
	}
}
