package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/spf13/cobra"
)

var customCmd = &cobra.Command{
	Use:   "custom",
	Short: "Manage custom key-value pairs in config",
	Long: `Manage custom key-value pairs that can be used in templates.

Custom variables are stored in .versionator.yaml and available as {{KeyName}} in templates.

Examples:
  versionator custom set AppName "My Application"
  versionator custom set BuildEnv production
  versionator custom get AppName
  versionator custom list
  versionator custom delete AppName

Then use in templates:
  versionator version -t "{{AppName}} v{{MajorMinorPatch}}"`,
}

var customSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a custom key-value pair",
	Long: `Set a custom key-value pair in .versionator.yaml.

The key becomes a template variable accessible as {{Key}}.

Examples:
  versionator custom set AppName "My Application"
  versionator custom set Environment production
  versionator custom set Copyright "2024 Acme Inc"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if err := config.SetCustom(key, value); err != nil {
			return fmt.Errorf("error setting custom value: %w", err)
		}

		fmt.Printf("Custom value set: %s = %s\n", key, value)
		return nil
	},
}

var customGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a custom value by key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		value, ok, err := config.GetCustom(key)
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		if !ok {
			return fmt.Errorf("%s: %s", ErrCustomKeyNotFound, key)
		}

		fmt.Println(value)
		return nil
	},
}

var customListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all custom key-value pairs",
	RunE: func(cmd *cobra.Command, args []string) error {
		custom, err := config.GetAllCustom()
		if err != nil {
			return fmt.Errorf("error loading custom values: %w", err)
		}

		if len(custom) == 0 {
			fmt.Println("No custom values defined")
			return nil
		}

		// Sort keys for consistent output
		keys := make([]string, 0, len(custom))
		for k := range custom {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Find max key length for alignment
		maxLen := 0
		for _, k := range keys {
			if len(k) > maxLen {
				maxLen = len(k)
			}
		}

		for _, k := range keys {
			fmt.Printf("{{%s}}%s = %s\n", k, strings.Repeat(" ", maxLen-len(k)), custom[k])
		}
		return nil
	},
}

var customDeleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete a custom key-value pair",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		if err := config.DeleteCustom(key); err != nil {
			return fmt.Errorf("error deleting custom value: %w", err)
		}

		fmt.Printf("Custom value deleted: %s\n", key)
		return nil
	},
}

func init() {
	customCmd.AddCommand(customSetCmd)
	customCmd.AddCommand(customGetCmd)
	customCmd.AddCommand(customListCmd)
	customCmd.AddCommand(customDeleteCmd)
	rootCmd.AddCommand(customCmd)
}
