# Meower Development Workflow

This guide explains how to work with the Meower framework template in development mode.

## Quick Start

### For Development

```bash
# 1. Clone the repository
git clone <your-repo>
cd meower/cmd/meower

# 2. Enable development mode
./dev-mode.sh

# 3. Start developing
cd template
docker compose up
```

### When Done Developing

```bash
# 1. Stop the services
docker compose down

# 2. Switch back to template mode
cd ..
./template-mode.sh

# 3. Commit your changes
git add .
git commit -m "Your changes"
git push
```

## How It Works

### The Problem

Go's `go:embed` directive ignores directories that contain `go.mod` files because they're treated as separate modules. This prevents the template files from being embedded into the CLI binary.

### The Solution

We use two modes:

1. **Template Mode** (default): `go.mod` and `go.sum` files are stored as `.template` files, allowing `go:embed` to work
2. **Development Mode**: `.template` files are converted to working `go.mod` and `go.sum` files for development

### File Structure

```
template/
├── api/
│   ├── go.mod.template     # Template version (embedded)
│   ├── go.sum.template     # Template version (embedded)
│   ├── go.mod              # Working version (dev mode only)
│   └── go.sum              # Working version (dev mode only)
└── web/
    ├── go.mod.template     # Template version (embedded)
    ├── go.sum.template     # Template version (embedded)
    ├── go.mod              # Working version (dev mode only)
    └── go.sum              # Working version (dev mode only)
```

## Scripts

### `dev-mode.sh`

- Creates working `go.mod` and `go.sum` files from `.template` files
- Sets up proper module names for development
- Configures local module references

### `template-mode.sh`

- Converts working files back to `.template` files
- Replaces development module names with template variables
- Cleans up generated files
- Prepares for git commit

## Development Tips

### Making Changes to Dependencies

1. Enable dev mode: `./dev-mode.sh`
2. Navigate to the specific module: `cd template/api` or `cd template/web`
3. Use normal Go commands: `go get`, `go mod tidy`, etc.
4. Test your changes with `docker compose up`
5. When done, switch back: `cd ../.. && ./template-mode.sh`

### Testing the CLI

```bash
# Build the CLI with embedded templates
go build -o meower .

# Test creating a new project
./meower new test-project
cd test-project
docker compose up
```

### Troubleshooting

**Problem**: `docker compose up` fails with module errors
**Solution**: Make sure you're in dev mode: `./dev-mode.sh`

**Problem**: `go:embed` not working
**Solution**: Make sure you're in template mode: `./template-mode.sh`

**Problem**: Git shows unexpected changes
**Solution**: Run `./template-mode.sh` before committing

## Automation

For even more convenience, use the provided Makefile:

```bash
make dev        # Enable dev mode and start services
make stop       # Stop services and switch to template mode
make build      # Build the CLI
make test       # Test the CLI by creating a new project
```