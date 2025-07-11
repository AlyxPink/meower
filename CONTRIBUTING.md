# Contributing to Meower 🐱

Thank you for your interest in contributing to Meower! This document provides guidelines and information for contributors.

## 🚀 Quick Start for Contributors

### Prerequisites

- **Go 1.23+**: Latest stable version
- **Docker & Docker Compose**: For development environment
- **Git**: For version control

### Development Setup

```bash
# 1. Fork and clone the repository
git clone https://github.com/YOUR-USERNAME/meower.git
cd meower

# 2. Install dependencies
go mod tidy

# 3. Build the CLI
go build -o meower ./cmd/meower

# 4. Test the CLI
./meower new test-project -m github.com/test/test-project
cd test-project
docker-compose up
```

## 🎯 Areas for Contribution

### 🔧 **High Priority**
- **Additional Generators**: Model generator, migration generator
- **Route Auto-Update**: Automatically update routing files when creating handlers
- **Template Validation**: Enhanced validation for generated templates
- **Testing Framework**: Unit and integration tests for CLI

### 🛠️ **Medium Priority**
- **Configuration Management**: Project-level configuration files
- **Plugin System**: Extensible generator system
- **Documentation**: API documentation generation
- **Performance**: Optimization of template processing

### 🎨 **Enhancement Ideas**
- **Interactive Mode**: `meower create` with prompts
- **Project Analytics**: Usage statistics and insights
- **Template Customization**: User-defined templates
- **Multi-language Support**: Support for other backend languages

## 📋 Development Guidelines

### Code Style

**Go Code Standards:**
```go
// ✅ Good: Clear, documented, descriptive
// GenerateHandler creates a complete gRPC service implementation.
// This includes proto files, server handlers, and web client stubs.
func (g *Generator) GenerateHandler(serviceName string) error {
    if err := g.validateInput(serviceName); err != nil {
        return fmt.Errorf("invalid input: %w", err)
    }
    // ... implementation
}

// ❌ Bad: No documentation, unclear naming
func (g *Generator) gen(s string) error {
    // ... implementation
}
```

**Template Standards:**
- Use clear placeholder names: `TEMPLATE_SERVICE_NAME` not `TMPL_SVC`
- Include helpful TODO comments in generated code
- Ensure generated code follows Go conventions
- Test templates with various input combinations

### Documentation Standards

**Function Documentation:**
```go
// processTemplate applies template variables to file content.
// 
// This function is responsible for:
// 1. Reading template files with proper encoding
// 2. Applying string replacements safely
// 3. Preserving file permissions and structure
// 4. Handling edge cases (empty files, binary content)
//
// Security: This function performs string replacement on file contents.
// Template variables are validated before reaching this function.
func processTemplate(content string, vars map[string]string) string {
    // implementation
}
```

### Testing Standards

**Unit Tests:**
```go
func TestTemplateVars_SetProject(t *testing.T) {
    tests := []struct {
        name        string
        projectName string
        modulePath  string
        wantErr     bool
        validate    func(*TemplateVars) error
    }{
        {
            name:        "valid project",
            projectName: "my-app",
            modulePath:  "github.com/user/my-app",
            wantErr:     false,
            validate: func(tv *TemplateVars) error {
                if tv.ProjectName != "my-app" {
                    return fmt.Errorf("expected 'my-app', got %s", tv.ProjectName)
                }
                return nil
            },
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tv := NewTemplateVars()
            err := tv.SetProject(tt.projectName, tt.modulePath)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("SetProject() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.validate != nil && err == nil {
                if err := tt.validate(tv); err != nil {
                    t.Error(err)
                }
            }
        })
    }
}
```

## 🔄 Contribution Workflow

### 1. **Issue Creation**

Before starting work:
- **Search existing issues** to avoid duplicates
- **Create detailed issues** with:
  - Clear problem description
  - Expected vs actual behavior
  - Steps to reproduce
  - Environment details

**Issue Templates:**

```markdown
**Bug Report:**
- Meower version: 
- Go version:
- OS/Platform:
- Command run:
- Error output:
- Expected behavior:

**Feature Request:**
- Use case:
- Proposed solution:
- Alternative solutions:
- Additional context:
```

### 2. **Development Process**

```bash
# 1. Create feature branch
git checkout -b feature/amazing-new-generator

# 2. Make changes with tests
# - Add functionality
# - Write/update tests
# - Update documentation

# 3. Verify everything works
go test ./...
./meower new test-project -m github.com/test/test
cd test-project && docker-compose up

# 4. Commit with clear messages
git commit -m "feat: add model generator with CRUD operations

- Generates SQLC-compatible SQL schema
- Creates proto definitions for model
- Includes validation and error handling
- Adds comprehensive tests

Closes #123"
```

### 3. **Pull Request Guidelines**

**PR Template:**
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manually tested CLI commands
- [ ] Generated projects work correctly

## Screenshots/Examples
(If applicable)

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] Tests added/updated
```

### 4. **Code Review Process**

**Review Criteria:**
- ✅ **Functionality**: Does it work as intended?
- ✅ **Code Quality**: Is it readable and maintainable?
- ✅ **Testing**: Are there adequate tests?
- ✅ **Documentation**: Is it properly documented?
- ✅ **Security**: Are there any security concerns?
- ✅ **Performance**: Is it efficient?

**Review Etiquette:**
- Be constructive and specific
- Explain the "why" behind suggestions
- Offer alternative solutions
- Recognize good work

## 🏗️ Architecture Guidelines

### CLI Structure

```
cmd/meower/          # CLI entry point
internal/
├── cli/             # Cobra commands
├── templates/       # Template processing system
├── generators/      # Code generators
└── utils/           # Shared utilities
```

### Design Principles

1. **Single Responsibility**: Each component has one clear purpose
2. **Composition**: Build complex functionality from simple parts
3. **Error Handling**: Explicit error handling with helpful messages
4. **Type Safety**: Leverage Go's type system to prevent bugs
5. **User Experience**: Optimize for developer happiness

### Adding New Generators

```go
// 1. Define in internal/generators/
type ModelGenerator struct {
    vars *templates.TemplateVars
}

func (g *ModelGenerator) Generate() error {
    // Implementation
}

// 2. Add CLI command in internal/cli/
var createModelCmd = &cobra.Command{
    Use:   "model [model-name]",
    Short: "Generate a database model",
    RunE:  runCreateModelCommand,
}

// 3. Register in internal/cli/create.go
func init() {
    createCmd.AddCommand(createModelCmd)
}
```

## 🧪 Testing Strategy

### Test Categories

1. **Unit Tests**: Individual functions and methods
2. **Integration Tests**: Component interactions
3. **CLI Tests**: End-to-end command execution
4. **Template Tests**: Generated code validation

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/templates

# Run with race detection
go test -race ./...
```

### Test Data

```bash
# Test fixtures location
testdata/
├── projects/        # Sample project structures
├── templates/       # Template test files
└── golden/          # Expected output files
```

## 📝 Documentation Standards

### Code Documentation

- **Public APIs**: Always documented with examples
- **Complex Logic**: Inline comments explaining the "why"
- **Security Considerations**: Document security implications
- **Performance Notes**: Explain performance trade-offs

### User Documentation

- **Clear Examples**: Show real-world usage
- **Progressive Disclosure**: Basic to advanced
- **Troubleshooting**: Common issues and solutions
- **Visual Aids**: Screenshots and diagrams where helpful

## 🔒 Security Considerations

### Input Validation

```go
// Always validate user input
func validateProjectName(name string) error {
    // Check for path traversal
    if strings.Contains(name, "..") {
        return fmt.Errorf("project name cannot contain '..'")
    }
    
    // Check for shell injection characters
    if regexp.MustCompile(`[;&|]`).MatchString(name) {
        return fmt.Errorf("project name contains invalid characters")
    }
    
    return nil
}
```

### File Operations

```go
// Always use absolute paths and validate destinations
func createFile(projectDir, filename string) error {
    // Ensure we're writing within project directory
    fullPath := filepath.Join(projectDir, filename)
    if !strings.HasPrefix(fullPath, projectDir) {
        return fmt.Errorf("invalid file path")
    }
    
    // ... safe file creation
}
```

## 🎉 Recognition

Contributors will be recognized in:
- **README.md**: Major contributors section
- **Release Notes**: Feature contributions
- **GitHub Discussions**: Community highlights

## 📞 Getting Help

- **GitHub Discussions**: General questions and ideas
- **GitHub Issues**: Bug reports and feature requests
- **Discord**: Real-time community chat (link in README)

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for helping make Meower amazing! 🐱✨**

*Every contribution, no matter how small, makes a difference.*