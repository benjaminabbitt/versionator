package cmd

import (
	"os"
	"versionator/internal/app"

	"github.com/spf13/cobra"
)

var appInstance *app.App
var applicationVersion string

var rootCmd = &cobra.Command{
	Use:   "application",
	Short: "A semantic version management tool",
	Long: `Versionator is a CLI tool for managing semantic versions.
It allows you to increment and decrement major, minor, and patch versions
stored in a VERSION file in the current directory.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

// SetApplicationVersion sets the versionator application version
func SetApplicationVersion(version string) {
	applicationVersion = version
	// Update the long help with version information
	rootCmd.Long = `Versionator is a CLI tool for managing semantic versions.
It allows you to increment and decrement major, minor, and patch versions
stored in a VERSION file in the current directory.

Versionator version: ` + version
}

func init() {
	// Initialize the app with all dependencies
	appInstance = app.NewApp()

	// Add version command to show current version
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show current version",
		Run: func(cmd *cobra.Command, args []string) {
			version, err := appInstance.GetVersionWithSuffix()
			if err != nil {
				cmd.Printf("Error reading version: %v\n", err)
				os.Exit(1)
			}
			cmd.Println(version)
		},
	})
}
