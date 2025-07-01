# Meower Development Makefile
# Automates the development workflow for the Meower framework

.PHONY: help dev stop build test clean template-mode dev-mode

# Default target
help:
	@echo "Meower Development Commands:"
	@echo ""
	@echo "  make dev        - Enable dev mode and start services"
	@echo "  make stop       - Stop services and switch to template mode"
	@echo "  make build      - Build the CLI binary"
	@echo "  make test       - Test CLI by creating a new project"
	@echo "  make clean      - Clean up test projects and binaries"
	@echo "  make dev-mode   - Enable development mode only"
	@echo "  make template-mode - Enable template mode only"
	@echo ""
	@echo "Quick workflow:"
	@echo "  1. make dev     (start developing)"
	@echo "  2. make stop    (when done)"
	@echo "  3. git commit   (commit changes)"

# Enable dev mode and start services
dev: dev-mode
	@echo "🚀 Starting development services..."
	@cd template && docker compose up

# Stop services and switch to template mode
stop:
	@echo "🛑 Stopping services..."
	@cd template && docker compose down || true
	@$(MAKE) template-mode
	@echo "✅ Ready for git commit!"

# Build the CLI binary
build: template-mode
	@echo "🔨 Building meower CLI..."
	@go build -o meower .
	@echo "✅ Built: ./meower"

# Test the CLI by creating a new project
test: build
	@echo "🧪 Testing CLI..."
	@rm -rf test-project || true
	@./meower new test-project
	@echo "✅ Test project created: ./test-project"
	@echo "💡 To test: cd test-project && docker compose up"

# Clean up test projects and binaries
clean:
	@echo "🧹 Cleaning up..."
	@rm -rf test-project || true
	@rm -f meower || true
	@find template -name "*_templ.go" -delete || true
	@cd template && docker compose down || true
	@echo "✅ Cleaned up"

# Enable development mode only
dev-mode:
	@echo "🔧 Enabling development mode..."
	@./dev-mode.sh

# Enable template mode only
template-mode:
	@echo "📦 Enabling template mode..."
	@./template-mode.sh

# Quick development cycle
quick-dev: dev-mode
	@echo "⚡ Quick dev mode (no docker compose up)"
	@echo "💡 Run 'cd template && docker compose up' manually"

# Full development cycle with rebuild
full-dev: clean dev-mode
	@echo "🔄 Full development cycle started"
	@cd template && docker compose up --build

# Prepare for production build
prod-build: template-mode build
	@echo "🏭 Production build ready"
	@echo "📦 Binary: ./meower"