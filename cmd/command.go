package cmd

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

func Execute() error {

	rootCmd := baseCommand("hashi-up")
	rootCmd.AddCommand(VersionCommand())
	rootCmd.AddCommand(CompletionCommand())
	rootCmd.AddCommand(productCommand("consul", InstallConsulCommand))
	rootCmd.AddCommand(productCommand("nomad", InstallNomadCommand))
	rootCmd.AddCommand(productCommand("vault", InstallVaultCommand))
	rootCmd.AddCommand(productCommand("boundary", InstallBoundaryCommand))
	rootCmd.AddCommand(productCommand("terraform", nil))
	rootCmd.AddCommand(productCommand("packer", nil))
	rootCmd.AddCommand(productCommand("vagrant", nil))
	rootCmd.AddCommand(productCommand("waypoint", nil))

	return rootCmd.Execute()
}

func baseCommand(name string) *cobra.Command {
	return &cobra.Command{
		Use: name,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
}

func productCommand(name string, installer func() *cobra.Command) *cobra.Command {
	command := baseCommand(name)
	command.AddCommand(GetCommand(name))
	if installer != nil {
		command.AddCommand(installer())
		command.AddCommand(UninstallCommand(name))
	}
	return command
}

func expandPath(path string) string {
	res, _ := homedir.Expand(path)
	return res
}

func info(message string) {
	fmt.Println("[INFO] " + message)
}
