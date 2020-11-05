package main

import (
	"github.com/jsiebens/hashi-up/cmd"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "hashi-up",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var consul = &cobra.Command{
		Use: "consul",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var nomad = &cobra.Command{
		Use: "nomad",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var vault = &cobra.Command{
		Use: "vault",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	nomad.AddCommand(cmd.InstallNomadCommand())
	consul.AddCommand(cmd.InstallConsulCommand())
	vault.AddCommand(cmd.InstallVaultCommand())

	rootCmd.AddCommand(nomad)
	rootCmd.AddCommand(consul)
	rootCmd.AddCommand(vault)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
