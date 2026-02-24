package cmd

import (
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for versionator.

To load completions:

Bash:
  $ source <(versionator completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ versionator completion bash > /etc/bash_completion.d/versionator
  # macOS:
  $ versionator completion bash > $(brew --prefix)/etc/bash_completion.d/versionator

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ versionator completion zsh > "${fpath[1]}/_versionator"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ versionator completion fish | source

  # To load completions for each session, execute once:
  $ versionator completion fish > ~/.config/fish/completions/versionator.fish

PowerShell:
  PS> versionator completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> versionator completion powershell > versionator.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletionV2(out, true)
		case "zsh":
			return cmd.Root().GenZshCompletion(out)
		case "fish":
			return cmd.Root().GenFishCompletion(out, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(out)
		}
		return nil
	},
}

func init() {
	// Disable the default completion command from Cobra
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(completionCmd)
}
