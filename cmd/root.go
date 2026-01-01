package cmd

import (
	"fmt"
	"strings"

	"github.com/benjaminabbitt/versionator/cmd/output"
	"github.com/benjaminabbitt/versionator/internal/buildinfo"
	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/logging"
	"github.com/benjaminabbitt/versionator/pkg/plugin/loader"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var logOutput string
var verboseCount int

// pluginLoader holds the external plugin loader instance
var pluginLoader *loader.Loader

// Cleanup performs cleanup of resources (plugins, loggers, etc.)
// This should be called before process exit to ensure child processes are terminated.
func Cleanup() {
	if pluginLoader != nil {
		pluginLoader.Close()
		pluginLoader = nil
	}
}

// Marker for "flag provided without value" - use defaults
const useDefaultMarker = "\x00DEFAULT\x00"

var rootCmd = &cobra.Command{
	Use:   "versionator",
	Short: "A semantic version management tool",
	Long: `Versionator is a CLI tool for managing semantic versions.
It allows you to increment and decrement major, minor, and patch versions
stored in a VERSION file in the current directory.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// If log format wasn't explicitly set via flag, use config default
		if !cmd.PersistentFlags().Changed("log-format") {
			if cfg, err := config.ReadConfig(); err == nil {
				logOutput = cfg.Logging.Output
			}
		}

		// Initialize logger with the specified output format and verbosity
		if err := logging.InitLoggerWithVerbosity(logOutput, verboseCount); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		// Load external plugins
		logger := logging.GetLogger()
		pluginLoader = loader.NewLoaderWithVerbosity(logger, verboseCount)
		count, err := pluginLoader.DiscoverAndLoad()
		if err != nil {
			logger.Warn("error discovering plugins", zap.Error(err))
		}
		if count > 0 {
			logger.Debug("loaded external plugins")
		}

		// Share plugin loader with output subpackage
		output.PluginLoader = pluginLoader

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if pluginLoader != nil {
			pluginLoader.Close()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Set version for --version flag
	rootCmd.Version = buildinfo.Version

	// Add persistent flag for log output format
	rootCmd.PersistentFlags().StringVar(&logOutput, "log-format", "console", "Log output format (console, json, development)")

	// Add persistent verbose flag - can be repeated for more verbosity (-v, -vv, -vvv)
	rootCmd.PersistentFlags().CountVarP(&verboseCount, "verbose", "v", "Increase verbosity (-v=info, -vv=debug)")
}

// parseSetFlags parses --set key=value flags into a map
func parseSetFlags(setFlags []string) map[string]string {
	result := make(map[string]string)
	for _, s := range setFlags {
		if idx := strings.Index(s, "="); idx > 0 {
			key := s[:idx]
			value := s[idx+1:]
			result[key] = value
		}
	}
	return result
}
