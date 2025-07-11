package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlyxPink/meower/internal/templates"
	"github.com/AlyxPink/meower/internal/validation"
)

// ProjectConfig holds all configuration needed for project generation
type ProjectConfig struct {
	ProjectName string
	ModulePath  string
	Force       bool
	DestDir     string
}

// ProjectGenerator handles the project generation workflow
type ProjectGenerator struct {
	validator *validation.Validator
	config    *ProjectConfig
}

// NewProjectGenerator creates a new project generator
func NewProjectGenerator(config *ProjectConfig) *ProjectGenerator {
	return &ProjectGenerator{
		validator: validation.NewValidator(),
		config:    config,
	}
}

// ValidateAndPrepare validates the project configuration and prepares for generation
func (pg *ProjectGenerator) ValidateAndPrepare() error {
	// Validate project name
	if err := pg.validator.Project.ValidateProjectName(pg.config.ProjectName); err != nil {
		return fmt.Errorf("invalid project name: %w", err)
	}

	// Set default module path if not provided
	if pg.config.ModulePath == "" {
		pg.config.ModulePath = fmt.Sprintf("%s/%s", DefaultModulePrefix, pg.config.ProjectName)
		fmt.Println(warningStyle.Render("⚠️  No module path specified, using:"), pg.config.ModulePath)
	}

	// Validate module path
	if err := pg.validator.Project.ValidateModulePath(pg.config.ModulePath); err != nil {
		return fmt.Errorf("invalid module path: %w", err)
	}

	// Check if directory already exists
	if _, err := os.Stat(pg.config.ProjectName); err == nil && !pg.config.Force {
		return fmt.Errorf("directory already exists: %s (use --force flag to overwrite)", pg.config.ProjectName)
	}

	// Set destination directory
	pg.config.DestDir = filepath.Join(".", pg.config.ProjectName)

	return nil
}

// CreateProjectStructure creates the basic project directory structure
func (pg *ProjectGenerator) CreateProjectStructure() error {
	// Create destination directory
	if err := os.MkdirAll(pg.config.DestDir, 0o755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create marker file
	if err := pg.createMarkerFile(); err != nil {
		return fmt.Errorf("failed to create marker file: %w", err)
	}

	return nil
}

// ProcessTemplates processes and copies template files to the destination
func (pg *ProjectGenerator) ProcessTemplates() error {
	// Create template variables
	vars := templates.NewTemplateVars()
	if err := vars.SetProject(pg.config.ProjectName, pg.config.ModulePath); err != nil {
		return fmt.Errorf("failed to set project variables: %w", err)
	}

	// Use optimized processor for better performance
	processor := templates.NewOptimizedProcessorWithStats(vars)

	fmt.Println(subtitleStyle.Render("📂 Copying project structure..."))

	if err := processor.ProcessEmbeddedFiles(pg.config.DestDir); err != nil {
		// Fallback to local files (for development)
		return pg.fallbackToLocalFiles(vars)
	}

	// Show processing statistics
	stats := processor.GetStats()
	fmt.Printf(successStyle.Render("✅ Using embedded template files (%d files processed, %d skipped)\n"),
		stats.FilesProcessed, stats.FilesSkipped)

	return nil
}

// PostProcess performs post-processing steps after template generation
func (pg *ProjectGenerator) PostProcess() error {
	// Clean up CLI-specific files from the generated project
	cleanupGeneratedProject(pg.config.DestDir)

	// Copy guide to generated project
	copyGuideToProject(pg.config.DestDir)

	return nil
}

// ShowSuccessMessage displays the success message and next steps
func (pg *ProjectGenerator) ShowSuccessMessage() {
	fmt.Println(successStyle.Render("✅ Project created successfully!"))
	fmt.Println()
	fmt.Println(titleStyle.Render("🚀 Next steps:"))
	fmt.Println(subtitleStyle.Render("1. cd " + pg.config.ProjectName))
	fmt.Println(subtitleStyle.Render("2. docker-compose up"))
	fmt.Println(subtitleStyle.Render("3. Open http://localhost:" + DefaultHTTPPort))
	fmt.Println()
	fmt.Println(subtitleStyle.Render("Happy coding! 🎉"))
}

// Generate executes the complete project generation workflow
func (pg *ProjectGenerator) Generate() error {
	// Print header
	fmt.Println(titleStyle.Render("🐱 Creating new Meower project"))
	fmt.Println(subtitleStyle.Render("Project:"), pg.config.ProjectName)
	fmt.Println(subtitleStyle.Render("Module:"), pg.config.ModulePath)
	fmt.Println()

	// Execute generation steps
	steps := []struct {
		name string
		fn   func() error
	}{
		{"validate configuration", pg.ValidateAndPrepare},
		{"create project structure", pg.CreateProjectStructure},
		{"process templates", pg.ProcessTemplates},
		{"post-process", pg.PostProcess},
	}

	for _, step := range steps {
		if err := step.fn(); err != nil {
			return fmt.Errorf("failed to %s: %w", step.name, err)
		}
	}

	// Show success message
	pg.ShowSuccessMessage()
	return nil
}

// createMarkerFile creates the .meowed marker file
func (pg *ProjectGenerator) createMarkerFile() error {
	markerFile := filepath.Join(pg.config.DestDir, MarkerFileName)
	return os.WriteFile(markerFile, []byte(MarkerFileContent), 0o644)
}

// fallbackToLocalFiles handles fallback to local development files
func (pg *ProjectGenerator) fallbackToLocalFiles(vars *templates.TemplateVars) error {
	templateDir, err := getTemplateSourceDir()
	if err != nil {
		return fmt.Errorf("embedded files failed and no local source found: %w", err)
	}

	fmt.Println(warningStyle.Render("⚠️  Using local development files (embedded files failed)"))

	// Use local file processor
	localProcessor := templates.NewFileProcessor(vars)
	if err := localProcessor.ProcessDirectory(templateDir, pg.config.DestDir); err != nil {
		return fmt.Errorf("failed to process local templates: %w", err)
	}

	return nil
}
