package cmd

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

func Execute() error {
	var rootCmd = &cobra.Command{
		Use: "hashi-up",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		SilenceErrors: true,
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

	var terraform = &cobra.Command{
		Use: "terraform",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var packer = &cobra.Command{
		Use: "packer",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var vagrant = &cobra.Command{
		Use: "vagrant",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	nomad.AddCommand(InstallNomadCommand())
	nomad.AddCommand(UninstallCommand("nomad"))
	nomad.AddCommand(GetCommand("nomad"))

	consul.AddCommand(InstallConsulCommand())
	consul.AddCommand(UninstallCommand("consul"))
	consul.AddCommand(GetCommand("consul"))

	vault.AddCommand(InstallVaultCommand())
	vault.AddCommand(UninstallCommand("vault"))
	vault.AddCommand(GetCommand("vault"))

	terraform.AddCommand(GetCommand("terraform"))
	packer.AddCommand(GetCommand("packer"))
	vagrant.AddCommand(GetCommand("vagrant"))

	rootCmd.AddCommand(VersionCommand())
	rootCmd.AddCommand(CompletionCommand())
	rootCmd.AddCommand(nomad)
	rootCmd.AddCommand(consul)
	rootCmd.AddCommand(vault)
	rootCmd.AddCommand(terraform)
	rootCmd.AddCommand(packer)
	rootCmd.AddCommand(vagrant)

	return rootCmd.Execute()
}

func expandPath(path string) string {
	res, _ := homedir.Expand(path)
	return res
}

func info(message string) {
	fmt.Println("[INFO] " + message)
}
