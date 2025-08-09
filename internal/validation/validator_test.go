package validation

import (
	"testing"
)

func TestProjectValidator_ValidateProjectName(t *testing.T) {
	validator := &ProjectValidator{}

	tests := []struct {
		name        string
		projectName string
		expectError bool
	}{
		// Valid cases
		{
			name:        "valid project name",
			projectName: "my-awesome-project",
			expectError: false,
		},
		{
			name:        "valid simple name",
			projectName: "my-project",
			expectError: false,
		},
		{
			name:        "valid single character",
			projectName: "a1",
			expectError: false,
		},
		{
			name:        "valid with numbers",
			projectName: "project123",
			expectError: false,
		},
		{
			name:        "valid simple lowercase",
			projectName: "myproject",
			expectError: false,
		},
		{
			name:        "valid with hyphens in middle",
			projectName: "my-awesome-project-name",
			expectError: false,
		},
		// Invalid cases - Mixed case variations
		{
			name:        "uppercase letters - MyProject",
			projectName: "MyProject",
			expectError: true,
		},
		{
			name:        "mixed case with hyphens - My-Project",
			projectName: "My-Project",
			expectError: true,
		},
		{
			name:        "mixed case - myProject",
			projectName: "myProject",
			expectError: true,
		},
		{
			name:        "all uppercase",
			projectName: "MYPROJECT",
			expectError: true,
		},
		{
			name:        "camelCase",
			projectName: "myAwesomeProject",
			expectError: true,
		},
		// Invalid cases - Structure issues
		{
			name:        "empty name",
			projectName: "",
			expectError: true,
		},
		{
			name:        "too short",
			projectName: "a",
			expectError: true,
		},
		{
			name:        "too long",
			projectName: "this-is-a-very-long-project-name-that-exceeds-fifty-characters-limit",
			expectError: true,
		},
		{
			name:        "starts with hyphen",
			projectName: "-invalid",
			expectError: true,
		},
		{
			name:        "ends with hyphen",
			projectName: "invalid-",
			expectError: true,
		},
		{
			name:        "double hyphens",
			projectName: "my--project",
			expectError: true,
		},
		{
			name:        "contains spaces",
			projectName: "invalid project",
			expectError: true,
		},
		{
			name:        "special characters",
			projectName: "project@123",
			expectError: true,
		},
		{
			name:        "starts with number",
			projectName: "123project",
			expectError: true,
		},
		{
			name:        "underscore instead of hyphen",
			projectName: "my_project",
			expectError: true,
		},
		{
			name:        "dot characters",
			projectName: "my.project",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateProjectName(tt.projectName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for project name '%s' but got none", tt.projectName)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for project name '%s' but got: %v", tt.projectName, err)
			}
		})
	}
}

func TestProjectValidator_ValidateModulePath(t *testing.T) {
	validator := &ProjectValidator{}

	tests := []struct {
		name        string
		modulePath  string
		expectError bool
	}{
		// Valid cases - Standard module paths
		{
			name:        "valid github path",
			modulePath:  "github.com/user/project",
			expectError: false,
		},
		{
			name:        "valid gitlab path",
			modulePath:  "gitlab.com/user/project",
			expectError: false,
		},
		{
			name:        "valid custom domain",
			modulePath:  "example.com/project",
			expectError: false,
		},
		{
			name:        "valid with subdirectories",
			modulePath:  "github.com/org/team/project",
			expectError: false,
		},
		// Valid cases - Common scenarios from user request
		{
			name:        "simple module name - myproject",
			modulePath:  "myproject",
			expectError: false,
		},
		{
			name:        "simple with hyphens - my-project",
			modulePath:  "my-project",
			expectError: false,
		},
		{
			name:        "mixed case simple - My-Project",
			modulePath:  "My-Project",
			expectError: false,
		},
		{
			name:        "github with simple name - github.com/AlyxPink/myproject",
			modulePath:  "github.com/AlyxPink/myproject",
			expectError: false,
		},
		{
			name:        "github with hyphens - github.com/Alyx-Pink/My-Project",
			modulePath:  "github.com/Alyx-Pink/My-Project",
			expectError: false,
		},
		{
			name:        "github with hyphens in project - github.com/AlyxPink/my-project",
			modulePath:  "github.com/AlyxPink/my-project",
			expectError: false,
		},
		{
			name:        "valid with dots in domain",
			modulePath:  "git.example.org/user/project",
			expectError: false,
		},
		{
			name:        "valid with numbers in path",
			modulePath:  "github.com/user123/project456",
			expectError: false,
		},
		{
			name:        "valid deep nested path",
			modulePath:  "github.com/org/team/subteam/project/module",
			expectError: false,
		},
		{
			name:        "valid with uppercase domain",
			modulePath:  "GitHub.com/User/Project",
			expectError: false,
		},
		{
			name:        "valid bitbucket path",
			modulePath:  "bitbucket.org/user/project",
			expectError: false,
		},
		{
			name:        "valid sourcehut path",
			modulePath:  "git.sr.ht/~user/project",
			expectError: false,
		},
		{
			name:        "valid codeberg path",
			modulePath:  "codeberg.org/user/project",
			expectError: false,
		},
		// Invalid cases
		{
			name:        "empty path",
			modulePath:  "",
			expectError: true,
		},
		{
			name:        "invalid characters - @",
			modulePath:  "github.com/user/project@123",
			expectError: true,
		},
		{
			name:        "invalid characters - #",
			modulePath:  "github.com/user/project#branch",
			expectError: true,
		},
		{
			name:        "invalid characters - ?",
			modulePath:  "github.com/user/project?param=value",
			expectError: true,
		},
		{
			name:        "starts with slash",
			modulePath:  "/github.com/user/project",
			expectError: true,
		},
		{
			name:        "ends with slash",
			modulePath:  "github.com/user/project/",
			expectError: true,
		},
		{
			name:        "contains spaces",
			modulePath:  "github.com/user/my project",
			expectError: true,
		},
		{
			name:        "double slash",
			modulePath:  "github.com//user/project",
			expectError: true,
		},
		{
			name:        "protocol included",
			modulePath:  "https://github.com/user/project",
			expectError: true,
		},
		{
			name:        "invalid characters - asterisk",
			modulePath:  "github.com/user/*project",
			expectError: true,
		},
		{
			name:        "invalid characters - percent",
			modulePath:  "github.com/user/project%20name",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateModulePath(tt.modulePath)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for module path '%s' but got none", tt.modulePath)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for module path '%s' but got: %v", tt.modulePath, err)
			}
		})
	}
}

func TestServiceValidator_ValidateServiceName(t *testing.T) {
	validator := &ServiceValidator{}

	tests := []struct {
		name        string
		serviceName string
		expectError bool
	}{
		{
			name:        "valid service name",
			serviceName: "UserService",
			expectError: false,
		},
		{
			name:        "valid with numbers",
			serviceName: "OrderService2",
			expectError: false,
		},
		{
			name:        "valid single word",
			serviceName: "Auth",
			expectError: false,
		},
		{
			name:        "empty name",
			serviceName: "",
			expectError: true,
		},
		{
			name:        "too short",
			serviceName: "A",
			expectError: true,
		},
		{
			name:        "lowercase start",
			serviceName: "userService",
			expectError: true,
		},
		{
			name:        "contains hyphen",
			serviceName: "User-Service",
			expectError: true,
		},
		{
			name:        "contains underscore",
			serviceName: "User_Service",
			expectError: true,
		},
		{
			name:        "contains spaces",
			serviceName: "User Service",
			expectError: true,
		},
		{
			name:        "reserved word - Service",
			serviceName: "Service",
			expectError: true,
		},
		{
			name:        "reserved word - Handler",
			serviceName: "Handler",
			expectError: true,
		},
		{
			name:        "reserved word - Controller",
			serviceName: "Controller",
			expectError: true,
		},
		{
			name:        "reserved word - Manager",
			serviceName: "Manager",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateServiceName(tt.serviceName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for service name '%s' but got none", tt.serviceName)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for service name '%s' but got: %v", tt.serviceName, err)
			}
		})
	}
}

func TestServiceValidator_ValidateHTTPMethods(t *testing.T) {
	validator := &ServiceValidator{}

	tests := []struct {
		name        string
		methods     []string
		expectError bool
	}{
		{
			name:        "valid methods",
			methods:     []string{"GET", "POST", "PUT", "DELETE"},
			expectError: false,
		},
		{
			name:        "single method",
			methods:     []string{"GET"},
			expectError: false,
		},
		{
			name:        "case insensitive",
			methods:     []string{"get", "post", "Put"},
			expectError: false,
		},
		{
			name:        "all valid methods",
			methods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
			expectError: false,
		},
		{
			name:        "empty methods",
			methods:     []string{},
			expectError: true,
		},
		{
			name:        "invalid method",
			methods:     []string{"GET", "INVALID"},
			expectError: true,
		},
		{
			name:        "empty string method",
			methods:     []string{"GET", ""},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateHTTPMethods(tt.methods)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for methods %v but got none", tt.methods)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for methods %v but got: %v", tt.methods, err)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "project name",
		Value:   "invalid-name",
		Rule:    "format",
		Message: "must be lowercase",
	}

	expected := "validation failed for project name 'invalid-name': must be lowercase"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestMultiError(t *testing.T) {
	multiErr := MultiError{
		Errors: []error{
			ValidationError{Field: "name", Value: "test", Rule: "length", Message: "too short"},
			ValidationError{Field: "module", Value: "test", Rule: "format", Message: "invalid format"},
		},
	}

	if !multiErr.HasErrors() {
		t.Error("Expected MultiError to have errors")
	}

	errorMsg := multiErr.Error()
	if !contains(errorMsg, "multiple validation errors") {
		t.Errorf("Expected error message to contain 'multiple validation errors', got: %s", errorMsg)
	}

	// Test single error
	singleErr := MultiError{
		Errors: []error{
			ValidationError{Field: "name", Value: "test", Rule: "length", Message: "too short"},
		},
	}

	singleMsg := singleErr.Error()
	if contains(singleMsg, "multiple validation errors") {
		t.Errorf("Expected single error message, got: %s", singleMsg)
	}

	// Test no errors
	noErr := MultiError{Errors: []error{}}
	if noErr.HasErrors() {
		t.Error("Expected MultiError to have no errors")
	}
}

func TestNewValidator(t *testing.T) {
	validator := NewValidator()

	if validator == nil {
		t.Error("Expected NewValidator to return non-nil validator")
		return
	}

	if validator.Project == nil {
		t.Error("Expected validator to have Project validator")
	}

	if validator.Service == nil {
		t.Error("Expected validator to have Service validator")
	}
}

// Helper function for string contains check
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		(len(str) > len(substr) &&
			(str[:len(substr)] == substr ||
				str[len(str)-len(substr):] == substr ||
				indexOf(str, substr) >= 0)))
}

func indexOf(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
