#!/bin/bash

# Dev Mode Script - Enables development mode for meower framework
# This script prepares the template directory for development by creating working go.mod files

set -e

TEMPLATE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/template"

echo "🔧 Enabling development mode..."

# Function to create working go.mod from template
create_working_gomod() {
    local dir="$1"
    local module_name="$2"

    if [ -f "$dir/go.mod.template" ]; then
        echo "  📝 Creating working go.mod in $dir"
        # Replace template variables with actual values
        sed "s/TEMPLATE_PROJECT_NAME/$module_name/g" "$dir/go.mod.template" > "$dir/go.mod"
    fi

    if [ -f "$dir/go.sum.template" ]; then
        echo "  📝 Creating working go.sum in $dir"
        cp "$dir/go.sum.template" "$dir/go.sum"
    fi
}

# Create working go.mod files for development
create_working_gomod "$TEMPLATE_DIR/api" "myapp"
create_working_gomod "$TEMPLATE_DIR/web" "myapp"

# Update web go.mod to reference local api module
if [ -f "$TEMPLATE_DIR/web/go.mod" ]; then
    echo "  🔗 Setting up local module references"
    # The replace directive is already in the template, no need to add it
fi

# Clean up any generated files for a clean development environment
echo "  🧹 Cleaning up generated files"
find "$TEMPLATE_DIR" -name "*_templ.go" -delete 2>/dev/null || true

echo "✅ Development mode enabled!"
echo ""
echo "📋 Next steps:"
echo "  1. cd cmd/meower/template"
echo "  2. docker compose up"
echo ""
echo "💡 When you're done developing:"
echo "  1. make stop"
echo "  2. git add . && git commit"
