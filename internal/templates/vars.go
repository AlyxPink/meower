package templates

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// TemplateVars holds all template variables with validation and auto-generation.
// This is the core data structure that powers the Meower CLI's code generation.
// It maintains a type-safe registry of all placeholder values used throughout
// the template system, ensuring consistent naming conventions and validation.
//
// Design principles:
// - Auto-generation: Derived fields are computed automatically from base values
// - Validation: All inputs are validated before being set
// - Consistency: Naming follows established conventions (kebab, snake, pascal)
// - Type Safety: Compile-time checks prevent template variable typos
type TemplateVars struct {
	// Project-level variables
	ProjectName      string `json:"project_name"`       // my-social-app
	ProjectNameUpper string `json:"project_name_upper"` // MY_SOCIAL_APP
	ProjectNameCamel string `json:"project_name_camel"` // MySocialApp
	ModulePath       string `json:"module_path"`        // github.com/user/my-social-app

	// Service-level variables
	ServiceName      string `json:"service_name"`       // UserService
	ServiceNameLower string `json:"service_name_lower"` // user
	ServiceNameSnake string `json:"service_name_snake"` // user_service
	ServiceNameKebab string `json:"service_name_kebab"` // user-service

	// Model-level variables
	ModelName       string `json:"model_name"`        // User
	ModelNameLower  string `json:"model_name_lower"`  // user
	ModelNamePlural string `json:"model_name_plural"` // users
	TableName       string `json:"table_name"`        // users

	// API version
	APIVersion string `json:"api_version"` // v1
}

// Template placeholder constants - these are used in actual code files
const (
	TEMPLATE_PROJECT_NAME       = "TEMPLATE_PROJECT_NAME"       // meower
	TEMPLATE_PROJECT_NAME_UPPER = "TEMPLATE_PROJECT_NAME_UPPER" // MEOWER
	TEMPLATE_PROJECT_NAME_CAMEL = "TEMPLATE_PROJECT_NAME_CAMEL" // Meower
	TEMPLATE_MODULE_PATH        = "TEMPLATE_MODULE_PATH"

	TEMPLATE_SERVICE_NAME       = "TEMPLATE_SERVICE_NAME"       // UserService
	TEMPLATE_SERVICE_NAME_LOWER = "TEMPLATE_SERVICE_NAME_LOWER" // user
	TEMPLATE_SERVICE_NAME_SNAKE = "TEMPLATE_SERVICE_NAME_SNAKE" // user_service
	TEMPLATE_SERVICE_NAME_KEBAB = "TEMPLATE_SERVICE_NAME_KEBAB" // user-service

	TEMPLATE_MODEL_NAME        = "TEMPLATE_MODEL_NAME"        // User
	TEMPLATE_MODEL_NAME_LOWER  = "TEMPLATE_MODEL_NAME_LOWER"  // user
	TEMPLATE_MODEL_NAME_PLURAL = "TEMPLATE_MODEL_NAME_PLURAL" // users
	TEMPLATE_TABLE_NAME        = "TEMPLATE_TABLE_NAME"        // users

	TEMPLATE_API_VERSION = "TEMPLATE_API_VERSION" // v1
)

// AllPlaceholders contains all known template placeholders for validation
var AllPlaceholders = []string{
	TEMPLATE_PROJECT_NAME,
	TEMPLATE_PROJECT_NAME_UPPER,
	TEMPLATE_PROJECT_NAME_CAMEL,
	TEMPLATE_MODULE_PATH,
	TEMPLATE_SERVICE_NAME,
	TEMPLATE_SERVICE_NAME_LOWER,
	TEMPLATE_SERVICE_NAME_SNAKE,
	TEMPLATE_SERVICE_NAME_KEBAB,
	TEMPLATE_MODEL_NAME,
	TEMPLATE_MODEL_NAME_LOWER,
	TEMPLATE_MODEL_NAME_PLURAL,
	TEMPLATE_TABLE_NAME,
	TEMPLATE_API_VERSION,
}

// NewTemplateVars creates a new TemplateVars with auto-generated fields
func NewTemplateVars() *TemplateVars {
	return &TemplateVars{
		APIVersion: "v1",
	}
}

// SetProject sets project-related variables and auto-generates derived fields
func (tv *TemplateVars) SetProject(projectName, modulePath string) error {
	if err := validateProjectName(projectName); err != nil {
		return fmt.Errorf("invalid project name: %w", err)
	}

	if err := validateModulePath(modulePath); err != nil {
		return fmt.Errorf("invalid module path: %w", err)
	}

	tv.ProjectName = projectName
	tv.ProjectNameUpper = strings.ToUpper(strings.ReplaceAll(projectName, "-", "_"))
	tv.ProjectNameCamel = toPascalCase(projectName)
	tv.ModulePath = modulePath

	return nil
}

// SetService sets service-related variables and auto-generates derived fields
func (tv *TemplateVars) SetService(serviceName string) error {
	if err := validateServiceName(serviceName); err != nil {
		return fmt.Errorf("invalid service name: %w", err)
	}

	tv.ServiceName = serviceName
	tv.ServiceNameLower = strings.ToLower(serviceName)
	tv.ServiceNameSnake = toSnakeCase(serviceName)
	tv.ServiceNameKebab = toKebabCase(serviceName)

	return nil
}

// SetModel sets model-related variables and auto-generates derived fields
func (tv *TemplateVars) SetModel(modelName string) error {
	if err := validateModelName(modelName); err != nil {
		return fmt.Errorf("invalid model name: %w", err)
	}

	tv.ModelName = modelName
	tv.ModelNameLower = strings.ToLower(modelName)
	tv.ModelNamePlural = toPlural(modelName)
	tv.TableName = tv.ModelNamePlural

	return nil
}

// ToReplacementMap converts TemplateVars to a map for string replacement.
// This method is the bridge between our type-safe template variables and the
// actual string replacement process. It only includes non-empty values to prevent
// accidental replacement with empty strings.
//
// Security note: The hardcoded "github.com/AlyxPink/meower" replacement ensures
// that the original module path in template files gets properly replaced.
func (tv *TemplateVars) ToReplacementMap() map[string]string {
	replacements := make(map[string]string)

	// Only add non-empty replacements to avoid replacing with empty strings
	if tv.ProjectName != "" {
		replacements[TEMPLATE_PROJECT_NAME] = tv.ProjectName
		replacements[TEMPLATE_PROJECT_NAME_UPPER] = tv.ProjectNameUpper
		replacements[TEMPLATE_PROJECT_NAME_CAMEL] = tv.ProjectNameCamel
	}
	if tv.ModulePath != "" {
		replacements[TEMPLATE_MODULE_PATH] = tv.ModulePath
		// Also replace the original hardcoded module path
		replacements["github.com/AlyxPink/meower"] = tv.ModulePath
	}
	if tv.ServiceName != "" {
		replacements[TEMPLATE_SERVICE_NAME] = tv.ServiceName
		replacements[TEMPLATE_SERVICE_NAME_LOWER] = tv.ServiceNameLower
		replacements[TEMPLATE_SERVICE_NAME_SNAKE] = tv.ServiceNameSnake
		replacements[TEMPLATE_SERVICE_NAME_KEBAB] = tv.ServiceNameKebab
	}
	if tv.ModelName != "" {
		replacements[TEMPLATE_MODEL_NAME] = tv.ModelName
		replacements[TEMPLATE_MODEL_NAME_LOWER] = tv.ModelNameLower
		replacements[TEMPLATE_MODEL_NAME_PLURAL] = tv.ModelNamePlural
		replacements[TEMPLATE_TABLE_NAME] = tv.TableName
	}
	if tv.APIVersion != "" {
		replacements[TEMPLATE_API_VERSION] = tv.APIVersion
	}

	return replacements
}

// Validation functions
func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Allow lowercase letters, numbers, and hyphens
	matched, _ := regexp.MatchString(`^[a-z0-9-]+$`, name)
	if !matched {
		return fmt.Errorf("project name must contain only lowercase letters, numbers, and hyphens")
	}

	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("project name cannot start or end with a hyphen")
	}

	return nil
}

func validateModulePath(path string) error {
	if path == "" {
		return fmt.Errorf("module path cannot be empty")
	}

	// Basic validation for Go module path
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._/-]+$`, path)
	if !matched {
		return fmt.Errorf("module path contains invalid characters")
	}

	return nil
}

func validateServiceName(name string) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	// Must be PascalCase
	if !unicode.IsUpper(rune(name[0])) {
		return fmt.Errorf("service name must start with uppercase letter (PascalCase)")
	}

	return nil
}

func validateModelName(name string) error {
	if name == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	// Must be PascalCase
	if !unicode.IsUpper(rune(name[0])) {
		return fmt.Errorf("model name must start with uppercase letter (PascalCase)")
	}

	return nil
}

// String conversion utilities
func toPascalCase(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result.WriteRune('-')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// toPlural provides basic English pluralization rules.
// This is intentionally simple and can be enhanced with a proper pluralization
// library if more complex cases are needed (e.g., person->people, child->children).
// For most common programming use cases (User->users, Post->posts), this works well.
func toPlural(s string) string {
	s = strings.ToLower(s)
	// Handle words ending in 'y' (e.g., category -> categories)
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	// Handle words ending in 's', 'sh', 'ch' (e.g., class -> classes, dish -> dishes)
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "sh") || strings.HasSuffix(s, "ch") {
		return s + "es"
	}
	// Default: just add 's' (e.g., user -> users, post -> posts)
	return s + "s"
}
