package validation

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Project name validation
	projectNameRegex = regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$`)

	// Service name validation
	serviceNameRegex = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)

	// Module path validation (supports SourceHut format with ~ character)
	modulePathRegex = regexp.MustCompile(`^[a-zA-Z0-9.-]+(/[a-zA-Z0-9.~-]+)*$`)
)

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string
	Value   string
	Rule    string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s '%s': %s", e.Field, e.Value, e.Message)
}

// ProjectValidator handles project-level validation
type ProjectValidator struct{}

// ValidateProjectName validates project naming conventions
func (v *ProjectValidator) ValidateProjectName(name string) error {
	if name == "" {
		return ValidationError{
			Field:   "project name",
			Value:   name,
			Rule:    "required",
			Message: "project name cannot be empty",
		}
	}

	if len(name) < 2 {
		return ValidationError{
			Field:   "project name",
			Value:   name,
			Rule:    "length",
			Message: "project name must be at least 2 characters",
		}
	}

	if len(name) > 50 {
		return ValidationError{
			Field:   "project name",
			Value:   name,
			Rule:    "length",
			Message: "project name must be at most 50 characters",
		}
	}

	if !projectNameRegex.MatchString(name) {
		return ValidationError{
			Field:   "project name",
			Value:   name,
			Rule:    "format",
			Message: "project name must be lowercase, start with letter, contain only letters/numbers/hyphens, and not end with hyphen",
		}
	}

	if strings.Contains(name, "--") {
		return ValidationError{
			Field:   "project name",
			Value:   name,
			Rule:    "format",
			Message: "project name cannot contain consecutive hyphens",
		}
	}

	return nil
}

// ValidateModulePath validates Go module path format
func (v *ProjectValidator) ValidateModulePath(path string) error {
	if path == "" {
		return ValidationError{
			Field:   "module path",
			Value:   path,
			Rule:    "required",
			Message: "module path cannot be empty",
		}
	}

	if !modulePathRegex.MatchString(path) {
		return ValidationError{
			Field:   "module path",
			Value:   path,
			Rule:    "format",
			Message: "module path must be a valid Go module path (e.g. github.com/user/project)",
		}
	}

	return nil
}

// ServiceValidator handles service-level validation
type ServiceValidator struct{}

// ValidateServiceName validates service naming conventions
func (v *ServiceValidator) ValidateServiceName(name string) error {
	if name == "" {
		return ValidationError{
			Field:   "service name",
			Value:   name,
			Rule:    "required",
			Message: "service name cannot be empty",
		}
	}

	if len(name) < 2 {
		return ValidationError{
			Field:   "service name",
			Value:   name,
			Rule:    "length",
			Message: "service name must be at least 2 characters",
		}
	}

	if !serviceNameRegex.MatchString(name) {
		return ValidationError{
			Field:   "service name",
			Value:   name,
			Rule:    "format",
			Message: "service name must be PascalCase (e.g. UserService, OrderService)",
		}
	}

	// Check for common reserved words
	reserved := []string{"Service", "Handler", "Controller", "Manager"}
	for _, word := range reserved {
		if name == word {
			return ValidationError{
				Field:   "service name",
				Value:   name,
				Rule:    "reserved",
				Message: fmt.Sprintf("'%s' is a reserved word, please choose a more specific name", word),
			}
		}
	}

	return nil
}

// ValidateHTTPMethods validates HTTP method list
func (v *ServiceValidator) ValidateHTTPMethods(methods []string) error {
	if len(methods) == 0 {
		return ValidationError{
			Field:   "HTTP methods",
			Value:   strings.Join(methods, ","),
			Rule:    "required",
			Message: "at least one HTTP method must be specified",
		}
	}

	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"PATCH":   true,
		"DELETE":  true,
		"HEAD":    true,
		"OPTIONS": true,
	}

	for _, method := range methods {
		method = strings.ToUpper(strings.TrimSpace(method))
		if !validMethods[method] {
			return ValidationError{
				Field:   "HTTP methods",
				Value:   method,
				Rule:    "format",
				Message: fmt.Sprintf("'%s' is not a valid HTTP method", method),
			}
		}
	}

	return nil
}

// MultiError represents multiple validation errors
type MultiError struct {
	Errors []error
}

func (m MultiError) Error() string {
	if len(m.Errors) == 0 {
		return "no errors"
	}

	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}

	var messages []string
	for _, err := range m.Errors {
		messages = append(messages, err.Error())
	}

	return fmt.Sprintf("multiple validation errors: %s", strings.Join(messages, "; "))
}

// HasErrors returns true if there are any errors
func (m MultiError) HasErrors() bool {
	return len(m.Errors) > 0
}

// Validator provides a unified interface for all validation
type Validator struct {
	Project *ProjectValidator
	Service *ServiceValidator
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		Project: &ProjectValidator{},
		Service: &ServiceValidator{},
	}
}
