package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/AlyxPink/meower/internal/cli"
	"github.com/AlyxPink/meower/internal/templates"
)

// Embed template files (relative to this main.go file's location)
//
//go:embed all:template
var embeddedTemplateFiles embed.FS

func main() {
	// Set the embedded template files for the templates package
	templates.EmbeddedFiles = embeddedTemplateFiles

	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
