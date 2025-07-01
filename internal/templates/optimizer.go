package templates

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// OptimizedProcessor provides high-performance template processing
type OptimizedProcessor struct {
	vars     *TemplateVars
	replacer *strings.Replacer
}

// NewOptimizedProcessor creates a new optimized template processor
func NewOptimizedProcessor(vars *TemplateVars) *OptimizedProcessor {
	replacements := vars.ToReplacementMap()

	// Build replacer pairs for strings.Replacer
	var pairs []string
	for key, value := range replacements {
		pairs = append(pairs, key, value)
	}

	return &OptimizedProcessor{
		vars:     vars,
		replacer: strings.NewReplacer(pairs...),
	}
}

// ProcessEmbeddedFiles processes embedded template files with optimized string replacement
func (op *OptimizedProcessor) ProcessEmbeddedFiles(destDir string) error {
	return fs.WalkDir(EmbeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip CLI-specific paths
		if op.shouldSkipEmbedded(path, d) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Process path
		cleanPath := op.cleanPath(path)
		if cleanPath == "" {
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

		// Process file with optimized replacement
		return op.processFileOptimized(path, destPath)
	})
}

// processFileOptimized processes a single file with optimized string replacement
func (op *OptimizedProcessor) processFileOptimized(srcPath, destPath string) error {
	// Read file content
	content, err := EmbeddedFiles.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded file %s: %w", srcPath, err)
	}

	// Apply all replacements in a single pass using strings.Replacer
	processedContent := op.replacer.Replace(string(content))

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

// cleanPath extracts the clean destination path from embedded path
func (op *OptimizedProcessor) cleanPath(path string) string {
	// Handle different possible prefixes depending on build location
	prefixes := []string{"template/", "cmd/meower/template/"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return strings.TrimPrefix(path, prefix)
		}
	}

	// Skip root or paths that don't start with expected prefixes
	if path == "." {
		return ""
	}

	return ""
}

// shouldSkipEmbedded determines if an embedded file should be skipped
func (op *OptimizedProcessor) shouldSkipEmbedded(path string, d fs.DirEntry) bool {
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

// FileProcessingStats provides statistics about template processing
type FileProcessingStats struct {
	FilesProcessed int
	FilesSkipped   int
	BytesProcessed int64
	Replacements   int
}

// OptimizedProcessorWithStats extends OptimizedProcessor with statistics tracking
type OptimizedProcessorWithStats struct {
	*OptimizedProcessor
	Stats FileProcessingStats
}

// NewOptimizedProcessorWithStats creates a processor that tracks statistics
func NewOptimizedProcessorWithStats(vars *TemplateVars) *OptimizedProcessorWithStats {
	return &OptimizedProcessorWithStats{
		OptimizedProcessor: NewOptimizedProcessor(vars),
		Stats:              FileProcessingStats{},
	}
}

// ProcessEmbeddedFiles processes files while tracking statistics
func (ops *OptimizedProcessorWithStats) ProcessEmbeddedFiles(destDir string) error {
	return fs.WalkDir(EmbeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip CLI-specific paths
		if ops.shouldSkipEmbedded(path, d) {
			if !d.IsDir() {
				ops.Stats.FilesSkipped++
			}
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Process path
		cleanPath := ops.cleanPath(path)
		if cleanPath == "" {
			return nil
		}

		// Handle .template files
		if strings.HasSuffix(cleanPath, ".template") {
			cleanPath = strings.TrimSuffix(cleanPath, ".template")
		}

		destPath := filepath.Join(destDir, cleanPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		// Process file and track stats
		if err := ops.processFileWithStats(path, destPath); err != nil {
			return err
		}

		ops.Stats.FilesProcessed++
		return nil
	})
}

// processFileWithStats processes a file and updates statistics
func (ops *OptimizedProcessorWithStats) processFileWithStats(srcPath, destPath string) error {
	// Read file content
	content, err := EmbeddedFiles.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded file %s: %w", srcPath, err)
	}

	ops.Stats.BytesProcessed += int64(len(content))

	// Apply replacements and count them
	originalContent := string(content)
	processedContent := ops.replacer.Replace(originalContent)

	// Estimate replacement count (rough approximation)
	if originalContent != processedContent {
		ops.Stats.Replacements++
	}

	// Create directory and write file
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
	}

	if err := os.WriteFile(destPath, []byte(processedContent), 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	return nil
}

// GetStats returns the current processing statistics
func (ops *OptimizedProcessorWithStats) GetStats() FileProcessingStats {
	return ops.Stats
}
