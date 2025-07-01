# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Meower is a CLI framework for generating opinionated Go web applications with gRPC APIs, server-side rendering, and type-safe database queries. The project has been transformed from a monolithic application into a CLI tool that generates complete project scaffolding using embedded templates.

### Architecture

**CLI Framework Structure:**
```
cmd/meower/                 # CLI entry point with embedded templates
├── main.go                # Sets up embedded template files for the CLI
└── template/              # Complete project template (embedded via go:embed)
    ├── api/               # gRPC API server template
    ├── web/               # Fiber web server template  
    └── docker-compose.yml # Development environment

internal/
├── cli/                   # Cobra CLI commands and UI
├── templates/             # Template processing engine
├── generators/            # Code generators (handlers, services)
└── validation/            # Input validation framework
```

**Template System:**
- Uses `go:embed` to package complete project templates
- Template variable replacement with `{{.ModulePath}}`, `{{.ProjectName}}`, etc.
- Optimized processor with file counting and skip logic
- Fallback to local development files when embedded files fail

**Generated Project Structure:**
- **API Layer**: gRPC server with Protocol Buffers, SQLC database queries
- **Web Layer**: GoFiber HTTP server with Templ templates and TailwindCSS
- **Database**: PostgreSQL with type-safe SQLC-generated queries
- **Development**: Docker Compose environment with hot reload

## Essential Commands

### Development & Testing
```bash
# Build CLI binary
go build -o meower ./cmd/meower

# Run comprehensive integration tests (< 1 second execution)
go test ./internal/cli -v -run TestMeowerCLIComprehensive

# Run all CLI tests including validation and help
go test ./internal/cli -v

# Run specific test packages
go test ./internal/validation -v
go test ./internal/templates -v

# Test single functions
go test ./internal/cli -run TestCLIValidation -v
```

### CLI Usage
```bash
# Generate new project
./meower new project-name --module github.com/user/project-name

# Generate gRPC handler (must be run inside a Meower project)
./meower create handler ServiceName

# CLI help
./meower --help
./meower new --help
./meower create handler --help
```

### Project Testing Workflow
```bash
# Test complete CLI workflow
./meower new test-project --module github.com/test/test-project
cd test-project
docker-compose up
```

## Testing Framework

### Fast Integration Testing
- **TestMeowerCLIComprehensive**: Complete workflow test (< 1 second)
  - Project generation validation
  - Handler generation testing  
  - Build validation (with smart module error handling)
  - Server startup testing (mocked database)

### Test Project Safety
- All tests use standardized "test-project" name
- Protected by `.gitignore` patterns: `test-project/`, `test-*/`
- Tests run in temporary directories with cleanup

### Test Execution Strategy
- **Fast tests**: No external dependencies, mocked database connections
- **Build validation**: Tests API/web compilation in separate directories
- **Server startup**: Uses available ports, timeouts, graceful shutdown

## Key Implementation Details

### Template Processing
- **Embedded Templates**: `cmd/meower/main.go` sets `templates.EmbeddedFiles` via `go:embed`
- **Variable Replacement**: Uses simple string replacement for `{{.Variable}}` patterns
- **File Processing**: Handles `.template` files by removing extension after processing
- **Performance**: Optimized processor tracks file counts and skips appropriately

### Validation Framework
- **Project Names**: Lowercase, alphanumeric with hyphens, regex validation
- **Module Paths**: Go module path format validation
- **Service Names**: PascalCase validation for generated services
- **Error Handling**: Structured validation errors with field context

### CLI Architecture
- **Cobra Commands**: Root command with `new` and `create` subcommands
- **Styled Output**: Consistent UI using `charmbracelet/lipgloss`
- **Project Detection**: Uses `.meowed` marker file to detect Meower projects
- **Exit Codes**: Proper error handling for test validation

### Generated Project Features
- **Microservice Architecture**: Separate API and web servers with gRPC communication
- **Type Safety**: SQLC for database, Protocol Buffers for API, Templ for HTML
- **Development Environment**: Complete Docker Compose setup with hot reload
- **Build System**: Go modules with template-generated `go.mod` files

## Important File Locations

### Templates
- `cmd/meower/template/`: Complete embedded project template
- `internal/templates/`: Template processing engine
- `internal/templates/vars.go`: Template variable definitions

### CLI Implementation  
- `internal/cli/new.go`: Project generation command
- `internal/cli/create_handler.go`: Handler generation command
- `internal/cli/integration_test.go`: Comprehensive test suite

### Configuration
- `internal/cli/constants.go`: File paths, default values, marker content
- `internal/validation/`: Input validation framework

## Development Patterns

### Adding New Generators
1. Create generator in `internal/generators/`
2. Add CLI command in `internal/cli/`
3. Register command in `internal/cli/create.go`
4. Add integration tests in `internal/cli/integration_test.go`

### Template Modifications
- Edit files in `cmd/meower/template/`
- Use `{{.Variable}}` syntax for replacements
- Test with `go test ./internal/cli -v -run TestMeowerCLIComprehensive`

### Test Development
- Use `test-project` name for consistency
- Mock external dependencies (databases, networks)
- Target execution under 1 second for fast feedback
- Include cleanup in all test functions

## Generated Project Development

### Inside Generated Projects
```bash
# Start development environment
docker-compose up

# Generate protobuf files (if .proto files change)
./scripts/generate_protobuf.sh

# Generate SQLC code (if SQL queries change)  
cd api/db && sqlc generate

# Generate new handlers
meower create handler PaymentService
```

### Generated Project Structure
- **API Server**: `api/main.go` → gRPC server on :50051
- **Web Server**: `web/main.go` → HTTP server on :3000  
- **Database**: PostgreSQL on :5432 with pgweb UI on :5430
- **Hot Reload**: Automatic rebuilds on file changes