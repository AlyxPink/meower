package cli

import (
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate code components",
	Long: titleStyle.Render("🛠️  Code Generators") + "\n\n" +
		subtitleStyle.Render("Generate various code components for your Meower project:") + "\n" +
		subtitleStyle.Render("• gRPC handlers with protobuf definitions") + "\n" +
		subtitleStyle.Render("• Database models with SQLC queries") + "\n" +
		subtitleStyle.Render("• Web handlers and routes") + "\n" +
		subtitleStyle.Render("• Service layer components") + "\n",
}

func init() {
	rootCmd.AddCommand(createCmd)
}