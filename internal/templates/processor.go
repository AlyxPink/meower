package templates

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileProcessor handles template file processing and placeholder replacement.
// This is the core engine that transforms the Meower template project into
// a customized user project by recursively processing files and applying
// string replacements based on TemplateVars.
//
// Key responsibilities:
// - Recursive directory traversal with smart filtering
// - File content processing with placeholder replacement
// - Permission preservation during file copying
// - Prevention of infinite recursion via .meowed marker detection
type FileProcessor struct {
	vars *TemplateVars
}

// NewFileProcessor creates a new file processor with template variables
func NewFileProcessor(vars *TemplateVars) *FileProcessor {
	return &FileProcessor{
		vars: vars,
	}
}

// ProcessDirectory recursively processes all files in a directory, applying template replacements
func (fp *FileProcessor) ProcessDirectory(srcDir, destDir string) error {
	replacements := fp.vars.ToReplacementMap()

	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip certain directories and files
		if fp.shouldSkip(path, d) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Calculate relative path and destination
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		// Process file
		return fp.processFile(path, destPath, replacements)
	})
}

// ProcessFile processes a single file, applying template replacements
func (fp *FileProcessor) processFile(srcPath, destPath string, replacements map[string]string) error {
	// Get source file info
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", srcPath, err)
	}

	// Read source file
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", srcPath, err)
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

	// Write processed file with original permissions
	if err := os.WriteFile(destPath, []byte(processedContent), srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	return nil
}

// shouldSkip determines if a file or directory should be skipped during processing.
// This function implements the core filtering logic that prevents:
// 1. Processing hidden files/directories (security)
// 2. Infinite recursion via .meowed marker detection
// 3. Processing large irrelevant directories (performance)
// 4. Including binary files that shouldn't be templated
//
// The .meowed check is critical - it prevents the CLI from recursively
// processing its own generated projects, which would create infinite nested
// directory structures.
func (fp *FileProcessor) shouldSkip(path string, d fs.DirEntry) bool {
	name := d.Name()

	// Skip hidden files and directories
	if strings.HasPrefix(name, ".") {
		return true
	}

	// Skip if this is a directory with a .meowed marker file (generated project)
	// This is the key mechanism that prevents infinite recursion when running
	// the CLI from within a directory that contains generated projects.
	if d.IsDir() {
		meowedPath := filepath.Join(path, ".meowed")
		if _, err := os.Stat(meowedPath); err == nil {
			// This directory has been meowed, skip it to avoid recursion! 🐱
			return true
		}
	}

	// Skip specific directories
	skipDirs := []string{
		"node_modules",
		".git",
		"vendor",
		"dist",
		"build",
		"tmp",
		".next",
		".nuxt",
		"cmd",      // Skip CLI command directory
		"internal", // Skip internal CLI code
	}

	for _, skipDir := range skipDirs {
		if name == skipDir {
			return true
		}
	}

	// Skip binary files
	if !d.IsDir() {
		skipExtensions := []string{
			".exe", ".bin", ".so", ".dylib", ".dll",
			".jpg", ".jpeg", ".png", ".gif", ".svg",
			".mp4", ".mp3", ".wav", ".avi",
			".zip", ".tar", ".gz", ".7z",
			".pdf", ".doc", ".docx",
		}

		ext := strings.ToLower(filepath.Ext(name))
		for _, skipExt := range skipExtensions {
			if ext == skipExt {
				return true
			}
		}
	}

	return false
}

// ValidateTemplateFiles scans files for unknown placeholders
func ValidateTemplateFiles(rootDir string) error {
	var errors []string

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || shouldSkipForValidation(path, d) {
			return nil
		}

		// Check file for unknown placeholders
		if fileErrors := validateFileTemplates(path); len(fileErrors) > 0 {
			errors = append(errors, fileErrors...)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(errors) > 0 {
		return fmt.Errorf("template validation errors:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// validateFileTemplates checks a single file for unknown template placeholders
func validateFileTemplates(filePath string) []string {
	var errors []string

	file, err := os.Open(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading %s: %v", filePath, err)}
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Warning: failed to close file: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Find all TEMPLATE_ placeholders
		if strings.Contains(line, "TEMPLATE_") {
			placeholders := extractPlaceholders(line)
			for _, placeholder := range placeholders {
				if !isKnownPlaceholder(placeholder) {
					errors = append(errors, fmt.Sprintf("%s:%d - Unknown placeholder: %s", filePath, lineNum, placeholder))
				}
			}
		}
	}

	return errors
}

// extractPlaceholders finds all TEMPLATE_ placeholders in a line
func extractPlaceholders(line string) []string {
	var placeholders []string
	words := strings.Fields(line)

	for _, word := range words {
		if strings.HasPrefix(word, "TEMPLATE_") {
			// Clean up the placeholder (remove punctuation, quotes, etc.)
			placeholder := strings.Trim(word, "\"'`,;()[]{}:.")
			if placeholder != "" && strings.HasPrefix(placeholder, "TEMPLATE_") {
				placeholders = append(placeholders, placeholder)
			}
		}
	}

	return placeholders
}

// isKnownPlaceholder checks if a placeholder is in the known list
func isKnownPlaceholder(placeholder string) bool {
	for _, known := range AllPlaceholders {
		if placeholder == known {
			return true
		}
	}
	return false
}

// shouldSkipForValidation determines if a file should be skipped during validation
func shouldSkipForValidation(path string, d fs.DirEntry) bool {
	if d.IsDir() {
		return false
	}

	name := d.Name()

	// Only validate text files
	textExtensions := []string{
		".go", ".proto", ".sql", ".yaml", ".yml", ".json", ".toml",
		".md", ".txt", ".sh", ".dockerfile", ".templ",
	}

	ext := strings.ToLower(filepath.Ext(name))
	for _, textExt := range textExtensions {
		if ext == textExt {
			return false
		}
	}

	// Skip non-text files
	return true
}
