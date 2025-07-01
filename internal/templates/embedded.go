package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// EmbeddedFiles will be set by the main package that has access to the root files
var EmbeddedFiles embed.FS

// EmbeddedFileProcessor handles template processing from embedded files
type EmbeddedFileProcessor struct {
	vars *TemplateVars
}

// NewEmbeddedFileProcessor creates a processor that uses embedded files
func NewEmbeddedFileProcessor(vars *TemplateVars) *EmbeddedFileProcessor {
	return &EmbeddedFileProcessor{
		vars: vars,
	}
}

// ProcessEmbeddedFiles processes embedded template files to a destination directory
func (efp *EmbeddedFileProcessor) ProcessEmbeddedFiles(destDir string) error {
	replacements := efp.vars.ToReplacementMap()

	return fs.WalkDir(EmbeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip CLI-specific paths that are embedded but shouldn't be copied
		if efp.shouldSkipEmbedded(path, d) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Remove leading path prefixes from embedded paths
		cleanPath := path
		// Handle different possible prefixes depending on build location
		prefixes := []string{"template/", "cmd/meower/template/"}
		for _, prefix := range prefixes {
			if strings.HasPrefix(path, prefix) {
				cleanPath = strings.TrimPrefix(path, prefix)
				break
			}
		}

		if cleanPath == "" || cleanPath == path {
			// Skip root or paths that don't start with expected prefixes
			if path == "." {
				// Allow traversal of root directory but don't create it
				return nil
			}
			return nil
		}

		// Handle .template files (rename them to remove .template extension)
		if strings.HasSuffix(cleanPath, ".template") {
			cleanPath = strings.TrimSuffix(cleanPath, ".template")
		}

		destPath := filepath.Join(destDir, cleanPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		// Process file
		return efp.processEmbeddedFile(path, destPath, replacements)
	})
}

// processEmbeddedFile processes a single embedded file
func (efp *EmbeddedFileProcessor) processEmbeddedFile(srcPath, destPath string, replacements map[string]string) error {
	// Read embedded file
	content, err := EmbeddedFiles.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded file %s: %w", srcPath, err)
	}

	// Apply replacements
	processedContent := string(content)
	for placeholder, replacement := range replacements {
		processedContent = strings.ReplaceAll(processedContent, placeholder, replacement)
	}

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
	}

	// Write processed file
	if err := os.WriteFile(destPath, []byte(processedContent), 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	return nil
}

// shouldSkipEmbedded determines if an embedded file should be skipped
func (efp *EmbeddedFileProcessor) shouldSkipEmbedded(path string, d fs.DirEntry) bool {
	// Skip CLI-specific files that might be embedded
	skipPaths := []string{
		"cmd/meower/template/cmd",
		"cmd/meower/template/internal",
		"cmd/meower/template/CONTRIBUTING.md",
		"cmd/meower/template/.git",
		"cmd/meower/template/test_",
		"cmd/meower/template/debug_",
		"template/cmd",
		"template/internal",
		"template/CONTRIBUTING.md",
		"template/.git",
		"template/test_",
		"template/debug_",
	}

	for _, skip := range skipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}

	// Skip hidden files (except .gitkeep which we want, and "." which is root)
	if strings.HasPrefix(d.Name(), ".") && d.Name() != ".meowed" && d.Name() != ".gitkeep" && d.Name() != "." {
		return true
	}

	return false
}
