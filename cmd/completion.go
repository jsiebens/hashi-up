package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func CompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

$ source <(hashi-up completion bash)

# To load completions for each session, execute once:
Linux:
  $ hashi-up completion bash > /etc/bash_completion.d/hashi-up
MacOS:
  $ hashi-up completion bash > /usr/local/etc/bash_completion.d/hashi-up

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ hashi-up completion zsh > "${fpath[1]}/_hashi-up"

# You will need to start a new shell for this setup to take effect.

Fish:

$ hashi-up completion fish | source

# To load completions for each session, execute once:
$ hashi-up completion fish > ~/.config/fish/completions/hashi-up.fish

Powershell:

PS> hashi-up completion powershell | Out-String | Invoke-Expression

# To load completions for every new session, run:
PS> hashi-up completion powershell > hashi-up.ps1
# and source this file from your powershell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				_ = cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				_ = cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				_ = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				_ = cmd.Root().GenPowerShellCompletion(os.Stdout)
			}
		},
	}
}
