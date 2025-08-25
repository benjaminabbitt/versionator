package cmd

import (
	"os"
	"versionator/internal/app"
	"versionator/internal/logging"

	"github.com/spf13/cobra"
)

var appInstance *app.App
var logOutput string

var rootCmd = &cobra.Command{
	Use:   "application",
	Short: "A semantic version management tool",
	Long: `Versionator is a CLI tool for managing semantic versions.
It allows you to increment and decrement major, minor, and patch versions
stored in a VERSION file in the current directory.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// If log format wasn't explicitly set via flag, use config default
		if !cmd.PersistentFlags().Changed("log-format") {
			if cfg, err := appInstance.ReadConfig(); err == nil {
				logOutput = cfg.Logging.Output
			}
		}

		// Initialize logger with the specified output format
		if err := logging.InitLogger(logOutput); err != nil {
			cmd.Printf("Failed to initialize logger: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Initialize the app with all dependencies
	appInstance = app.NewApp()

	// Add persistent flag for log output format
	rootCmd.PersistentFlags().StringVar(&logOutput, "log-format", "console", "Log output format (console, json, development)")

	// Add version command to show current version
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show current version",
		Run: func(cmd *cobra.Command, args []string) {
			version, err := appInstance.GetVersionWithSuffix()
			if err != nil {
				logger := logging.GetSugaredLogger()
				logger.Errorw("Error reading version", "error", err)
				os.Exit(1)
			}
			cmd.Println(version)
		},
	})
}
