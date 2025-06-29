package cli

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/AlyxPink/meower/internal/generators"
	"github.com/AlyxPink/meower/internal/templates"
	"github.com/spf13/cobra"
)

// Flags for create handler command
var methods []string

// createHandlerCmd represents the create handler command
var createHandlerCmd = &cobra.Command{
	Use:   "handler [service-name]",
	Short: "Generate a gRPC service handler",
	Long: titleStyle.Render("📡 Generate gRPC Handler") + "\n\n" +
		subtitleStyle.Render("Generate a complete gRPC service with:") + "\n" +
		subtitleStyle.Render("• Protocol buffer service definition") + "\n" +
		subtitleStyle.Render("• Server-side handler implementation") + "\n" +
		subtitleStyle.Render("• Web client integration") + "\n" +
		subtitleStyle.Render("• Route registration") + "\n",
	Args: cobra.ExactArgs(1),
	RunE: runCreateHandlerCommand,
}

func init() {
	createCmd.AddCommand(createHandlerCmd)

	createHandlerCmd.Flags().StringSliceVarP(&methods, "methods", "m", []string{"Create", "Get", "Update", "Delete", "List"}, "gRPC methods to generate")
}

func runCreateHandlerCommand(cmd *cobra.Command, args []string) error {
	serviceName := args[0]

	// Validate we're in a Meower project
	if !isInMeowerProject() {
		fmt.Println(errorStyle.Render("❌ Not in a Meower project"))
		fmt.Println(subtitleStyle.Render("Run 'meower new project-name' to create a new project"))
		return nil
	}

	// Validate service name
	if err := validateServiceName(serviceName); err != nil {
		fmt.Println(errorStyle.Render("❌ Invalid service name:"), err)
		return nil
	}

	// Get current module path
	modulePath, err := getCurrentModulePath()
	if err != nil {
		fmt.Println(errorStyle.Render("❌ Error getting module path:"), err)
		return nil
	}

	// Create template variables
	vars := templates.NewTemplateVars()
	if err := vars.SetService(serviceName); err != nil {
		fmt.Println(errorStyle.Render("❌ Error setting service variables:"), err)
		return nil
	}

	// Set module path (extract from current project)
	vars.ModulePath = modulePath

	fmt.Println(titleStyle.Render("📡 Generating gRPC handler"))
	fmt.Println(subtitleStyle.Render("Service:"), serviceName)
	fmt.Println(subtitleStyle.Render("Methods:"), strings.Join(methods, ", "))
	fmt.Println()

	// Generate protocol buffer definition
	fmt.Println(subtitleStyle.Render("📝 Generating protobuf definition..."))
	generator := generators.NewHandlerGenerator(vars)
	if err := generator.GenerateProto(methods); err != nil {
		fmt.Println(errorStyle.Render("❌ Error generating proto:"), err)
		return nil
	}

	// Generate server handler
	fmt.Println(subtitleStyle.Render("🖥️  Generating server handler..."))
	if err := generator.GenerateServerHandler(methods); err != nil {
		fmt.Println(errorStyle.Render("❌ Error generating server handler:"), err)
		return nil
	}

	// Generate web handler
	fmt.Println(subtitleStyle.Render("🌐 Generating web handler..."))
	if err := generator.GenerateWebHandler(methods); err != nil {
		fmt.Println(errorStyle.Render("❌ Error generating web handler:"), err)
		return nil
	}

	// Update route registration
	fmt.Println(subtitleStyle.Render("🛣️  Updating routes..."))
	if err := generator.UpdateRoutes(); err != nil {
		fmt.Println(warningStyle.Render("⚠️  Could not automatically update routes:"), err)
		fmt.Println(subtitleStyle.Render("Please manually add routes for your new handler"))
	}

	fmt.Println(successStyle.Render("✅ Handler generated successfully!"))
	fmt.Println()
	fmt.Println(titleStyle.Render("🚀 Next steps:"))
	fmt.Println(subtitleStyle.Render("1. Run 'go generate ./...' to update protobuf files"))
	fmt.Println(subtitleStyle.Render("2. Implement your business logic in the handler"))
	fmt.Println(subtitleStyle.Render("3. Add any required database queries"))
	fmt.Println(subtitleStyle.Render("4. Test your new endpoints"))

	return nil
}

func isInMeowerProject() bool {
	// Check for the .meowed marker file - simple and reliable! 🐱
	_, err := os.Stat(".meowed")
	return err == nil
}

func getCurrentModulePath() (string, error) {
	// Try root go.mod first, then api/go.mod
	goModPaths := []string{"go.mod", "api/go.mod"}

	for _, goModPath := range goModPaths {
		if _, err := os.Stat(goModPath); err == nil {
			content, err := os.ReadFile(goModPath)
			if err != nil {
				continue
			}

			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					modulePath := strings.TrimSpace(strings.TrimPrefix(line, "module"))
					// Remove /api suffix if present
					return strings.TrimSuffix(modulePath, "/api"), nil
				}
			}
		}
	}

	return "", fmt.Errorf("go.mod not found in current directory or api/")
}

// validateServiceName ensures the service name follows Meower conventions.
// Service names must:
// - Be in PascalCase (e.g., UserService, not userService)
// - End with "Service" suffix for clarity and consistency
// - Be long enough to be meaningful (minimum 8 characters)
// - Not contain special characters that could break code generation
func validateServiceName(name string) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	// Must end with "Service" for consistency and clarity
	if !strings.HasSuffix(name, "Service") {
		return fmt.Errorf("service name must end with 'Service' (e.g. UserService, PostService, AuthService)")
	}

	// Minimum length check (at least "XService" = 8 chars)
	if len(name) < 8 {
		return fmt.Errorf("service name too short (minimum 8 characters: e.g., 'MyService')")
	}

	// Must start with uppercase (PascalCase)
	if !unicode.IsUpper(rune(name[0])) {
		return fmt.Errorf("service name must start with uppercase letter (PascalCase)")
	}

	// Check for invalid characters that could break code generation
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return fmt.Errorf("service name can only contain letters and numbers")
		}
	}

	return nil
}
