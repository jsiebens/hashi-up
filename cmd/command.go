package cmd

import (
	"fmt"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

type Installer func() *cobra.Command

func Execute() error {

	rootCmd := baseCommand("hashi-up")
	rootCmd.AddCommand(TlsCommands())
	rootCmd.AddCommand(VersionCommand())
	rootCmd.AddCommand(CompletionCommand())
	rootCmd.AddCommand(productCommand("consul", InstallConsulCommand))
	rootCmd.AddCommand(productCommand("nomad", InstallNomadCommand))
	rootCmd.AddCommand(productCommand("vault", InstallVaultCommand))
	rootCmd.AddCommand(productCommand("boundary", InstallBoundaryCommand, InitBoundaryDatabaseCommand))
	rootCmd.AddCommand(productCommand("terraform"))
	rootCmd.AddCommand(productCommand("packer"))
	rootCmd.AddCommand(productCommand("vagrant"))
	rootCmd.AddCommand(productCommand("waypoint"))
	rootCmd.AddCommand(productCommand("levant"))
	rootCmd.AddCommand(productCommand("consul-template"))
	rootCmd.AddCommand(productCommand("envconsul"))

	return rootCmd.Execute()
}

func baseCommand(name string) *cobra.Command {
	return &cobra.Command{
		Use: name,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		SilenceErrors: true,
	}
}

func productCommand(name string, installer ...Installer) *cobra.Command {
	command := baseCommand(name)
	command.Short = fmt.Sprintf("Install or download %s", strings.Title(name))
	command.Long = fmt.Sprintf("Install or download %s", strings.Title(name))
	command.AddCommand(GetCommand(name))
	if installer != nil {
		for _, y := range installer {
			command.AddCommand(y())
		}
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
