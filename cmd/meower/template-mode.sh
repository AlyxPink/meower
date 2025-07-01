#!/bin/bash

# Template Mode Script - Prepares template directory for embedding
# This script converts working go.mod files back to .template files for go:embed

set -e

TEMPLATE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/template"

echo "📦 Enabling template mode..."

# Function to convert working files back to templates
convert_to_template() {
    local dir="$1"
    local template_module_name="$2"

    if [ -f "$dir/go.mod" ]; then
        echo "  🔄 Converting go.mod to template in $dir"
        # Replace actual module name with template variable
        sed "s/myapp/TEMPLATE_PROJECT_NAME/g" "$dir/go.mod" > "$dir/go.mod.template"
        rm "$dir/go.mod"
    fi

    if [ -f "$dir/go.sum" ]; then
        echo "  🔄 Converting go.sum to template in $dir"
        cp "$dir/go.sum" "$dir/go.sum.template"
        rm "$dir/go.sum"
    fi
}

# Convert working files back to templates
convert_to_template "$TEMPLATE_DIR/api" "{{.ProjectName}}-api"
convert_to_template "$TEMPLATE_DIR/web" "{{.ProjectName}}-web"

# Clean up any generated files that shouldn't be in templates
echo "  🧹 Cleaning up generated files"
find "$TEMPLATE_DIR" -name "*.pb.go" -delete 2>/dev/null || true
find "$TEMPLATE_DIR" -name "query.*.sql.go" -delete 2>/dev/null || true
find "$TEMPLATE_DIR" -name "*_templ.go" -delete 2>/dev/null || true

echo "✅ Template mode enabled!"
echo ""
echo "📋 Files are now ready for embedding:"
echo "  - go.mod files converted to .template files"
echo "  - Generated files cleaned up"
echo "  - Ready for git commit"
