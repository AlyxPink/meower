package cli

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	// Styles for consistent UI
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		MarginLeft(2)

	subtitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		MarginLeft(2)

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Bold(true)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444")).
		Bold(true)

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Bold(true)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "meower",
	Short: "A powerful Go web framework CLI",
	Long: titleStyle.Render("🐱 Meower Framework") + "\n\n" +
		subtitleStyle.Render("An opinionated Go web framework featuring:") + "\n" +
		subtitleStyle.Render("• GoFiber web server with gRPC API") + "\n" +
		subtitleStyle.Render("• PostgreSQL with SQLC type-safe queries") + "\n" +
		subtitleStyle.Render("• Protocol Buffers for API definitions") + "\n" +
		subtitleStyle.Render("• Templ for server-side rendering") + "\n" +
		subtitleStyle.Render("• TailwindCSS for styling") + "\n",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add global flags here if needed
}