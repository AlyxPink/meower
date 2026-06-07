package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlyxPink/meower/internal/templates"
	"github.com/AlyxPink/meower/internal/validation"
)

// ProjectConfig holds all configuration needed for project generation
type ProjectConfig struct {
	ProjectName string
	ModulePath  string
	Force       bool
	DestDir     string

	// Feature toggles. Enabled by default (batteries-included); the --no-*
	// flags turn them off, which removes the corresponding code from the
	// generated project entirely.
	Auth    bool
	Workers bool
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

	// Use optimized processor for better performance, honoring the feature
	// toggles so disabled features' files are skipped.
	processor := templates.NewOptimizedProcessorWithFeatures(vars, templates.Features{
		Auth:    pg.config.Auth,
		Workers: pg.config.Workers,
	})

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

	// Finalize the generated CLAUDE.md: drop feature sections for disabled
	// features and create the AGENTS.md -> CLAUDE.md symlink.
	pg.finalizeAgentDocs()

	// Make shell scripts executable (the template processor writes 0644).
	pg.makeScriptsExecutable()

	return nil
}

// makeScriptsExecutable sets the executable bit on generated shell scripts. The
// template processor writes every file as 0644, but scripts/*.sh are invoked
// directly (e.g. ./scripts/generate_protobuf.sh in docker-compose), so they
// need +x. Best-effort: a missing scripts dir is not fatal.
func (pg *ProjectGenerator) makeScriptsExecutable() {
	scriptsDir := filepath.Join(pg.config.DestDir, "scripts")
	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		return // no scripts dir
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sh") {
			continue
		}
		path := filepath.Join(scriptsDir, e.Name())
		if err := os.Chmod(path, 0o755); err != nil {
			fmt.Printf("Warning: failed to chmod %s: %v\n", path, err)
		}
	}
}

// finalizeAgentDocs strips the auth/workers sections from the generated
// CLAUDE.md when those features are disabled, then symlinks AGENTS.md to it.
// Best-effort: a missing CLAUDE.md (it has no effect on a working project) is
// not fatal.
func (pg *ProjectGenerator) finalizeAgentDocs() {
	claudePath := filepath.Join(pg.config.DestDir, "CLAUDE.md")
	content, err := os.ReadFile(claudePath)
	if err != nil {
		return // no CLAUDE.md; nothing to finalize
	}

	doc := string(content)
	// Remove a section entirely when its feature is off; otherwise just strip
	// the marker comments so the kept section reads cleanly.
	if pg.config.Auth {
		doc = stripDocMarkers(doc, "AUTH-SECTION")
	} else {
		doc = stripDocSection(doc, "AUTH-SECTION")
	}
	if pg.config.Workers {
		doc = stripDocMarkers(doc, "WORKERS-SECTION")
	} else {
		doc = stripDocSection(doc, "WORKERS-SECTION")
	}
	if doc != string(content) {
		if err := os.WriteFile(claudePath, []byte(doc), 0o644); err != nil {
			fmt.Printf("Warning: failed to update CLAUDE.md: %v\n", err)
		}
	}

	// AGENTS.md -> CLAUDE.md (matches the convention many agents look for).
	agentsPath := filepath.Join(pg.config.DestDir, "AGENTS.md")
	_ = os.Remove(agentsPath) // ignore if absent
	if err := os.Symlink("CLAUDE.md", agentsPath); err != nil {
		fmt.Printf("Warning: failed to create AGENTS.md symlink: %v\n", err)
	}
}

// stripDocSection removes the block between <!-- NAME --> and <!-- /NAME -->
// markers (inclusive) from a markdown document.
func stripDocSection(doc, name string) string {
	start := "<!-- " + name + " -->"
	end := "<!-- /" + name + " -->"
	si := strings.Index(doc, start)
	ei := strings.Index(doc, end)
	if si == -1 || ei == -1 || ei < si {
		return doc
	}
	ei += len(end)
	// Also consume a trailing newline left by the removed block.
	if ei < len(doc) && doc[ei] == '\n' {
		ei++
	}
	return doc[:si] + doc[ei:]
}

// stripDocMarkers removes just the <!-- NAME --> / <!-- /NAME --> marker
// comment lines, leaving the section content in place.
func stripDocMarkers(doc, name string) string {
	for _, marker := range []string{"<!-- " + name + " -->\n", "<!-- /" + name + " -->\n"} {
		doc = strings.Replace(doc, marker, "", 1)
	}
	return doc
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
