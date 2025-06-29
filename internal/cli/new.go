package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AlyxPink/meower/internal/templates"
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

// runNewCommand implements the core project scaffolding logic.
// This function orchestrates the entire project creation process:
// 1. Input validation and default value assignment
// 2. Directory creation and marker file placement
// 3. Template processing and file generation
// 4. Cleanup and success messaging
func runNewCommand(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Validate project name against our naming conventions
	// (lowercase, hyphens allowed, no leading/trailing hyphens)
	if err := validateProjectName(projectName); err != nil {
		fmt.Println(errorStyle.Render("❌ Invalid project name:"), err)
		return nil
	}

	// Set default module path if not provided
	if modulePath == "" {
		modulePath = fmt.Sprintf("github.com/user/%s", projectName)
		fmt.Println(warningStyle.Render("⚠️  No module path specified, using:"), modulePath)
	}

	// Check if directory already exists
	if _, err := os.Stat(projectName); err == nil && !force {
		fmt.Println(errorStyle.Render("❌ Directory already exists:"), projectName)
		fmt.Println(subtitleStyle.Render("Use --force flag to overwrite"))
		return nil
	}

	// Create template variables
	vars := templates.NewTemplateVars()
	if err := vars.SetProject(projectName, modulePath); err != nil {
		fmt.Println(errorStyle.Render("❌ Error setting project variables:"), err)
		return nil
	}

	// Get template source directory (current meower project structure)
	templateDir, err := getTemplateSourceDir()
	if err != nil {
		fmt.Println(errorStyle.Render("❌ Error finding template source:"), err)
		return nil
	}

	// Create destination directory
	destDir := filepath.Join(".", projectName)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		fmt.Println(errorStyle.Render("❌ Error creating project directory:"), err)
		return nil
	}

	// Create the .meowed marker file - this serves multiple purposes:
	// 1. Prevents infinite recursion during template processing
	// 2. Marks the project as Meower-generated for CLI command detection
	// 3. Provides a fun "badge" for users to discover
	markerFile := filepath.Join(destDir, ".meowed")
	funnyMessage := `🐱 This project has been MEOWED! 🐱

Congratulations! Your project was lovingly crafted by the Meower CLI.
You're now part of the exclusive club of developers who've been meowed.

May your code purr smoothly and your builds never hiss! 🚀

Generated with Meower Framework
https://github.com/AlyxPink/meower`
	if err := os.WriteFile(markerFile, []byte(funnyMessage), 0o644); err != nil {
		fmt.Println(errorStyle.Render("❌ Error creating marker file:"), err)
		return nil
	}

	fmt.Println(titleStyle.Render("🐱 Creating new Meower project"))
	fmt.Println(subtitleStyle.Render("Project:"), projectName)
	fmt.Println(subtitleStyle.Render("Module:"), modulePath)
	fmt.Println()

	// Process template files
	fmt.Println(subtitleStyle.Render("📂 Copying project structure..."))
	processor := templates.NewFileProcessor(vars)
	if err := processor.ProcessDirectory(templateDir, destDir); err != nil {
		fmt.Println(errorStyle.Render("❌ Error processing templates:"), err)
		return nil
	}

	// Clean up CLI-specific files from the generated project
	cleanupGeneratedProject(destDir)

	// Copy guide to generated project
	copyGuideToProject(destDir)

	fmt.Println(successStyle.Render("✅ Project created successfully!"))
	fmt.Println()
	fmt.Println(titleStyle.Render("🚀 Next steps:"))
	fmt.Println(subtitleStyle.Render("1. cd " + projectName))
	fmt.Println(subtitleStyle.Render("2. docker-compose up"))
	fmt.Println(subtitleStyle.Render("3. Open http://localhost:3000"))
	fmt.Println()
	fmt.Println(subtitleStyle.Render("Happy coding! 🎉"))

	return nil
}

func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if strings.Contains(name, " ") {
		return fmt.Errorf("project name cannot contain spaces")
	}

	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("project name cannot start or end with a hyphen")
	}

	// Check for invalid characters
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_') {
			return fmt.Errorf("project name contains invalid character: %c", char)
		}
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
		os.RemoveAll(fullPath)
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
	os.WriteFile(destPath, content, 0o644)
}
