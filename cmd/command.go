package cmd

import (
	"fmt"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/muesli/coral"
)

type Installer func() *coral.Command

func Execute() error {

	rootCmd := baseCommand("hashi-up")
	rootCmd.AddCommand(TlsCommands())
	rootCmd.AddCommand(VersionCommand())
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
	rootCmd.AddCommand(productCommand("nomad-pack"))

	return rootCmd.Execute()
}

func baseCommand(name string) *coral.Command {
	return &coral.Command{
		Use: name,
		Run: func(cmd *coral.Command, args []string) {
			cmd.Help()
		},
		SilenceErrors: true,
	}
}

func productCommand(name string, installer ...Installer) *coral.Command {
	command := baseCommand(name)
	command.Short = fmt.Sprintf("Install or download %s", strings.Title(name))
	command.Long = fmt.Sprintf("Install or download %s", strings.Title(name))
	command.AddCommand(GetCommand(name))
	if installer != nil {
		for _, y := range installer {
			command.AddCommand(y())
		}
		command.AddCommand(ManageServiceCommand("stop", name))
		command.AddCommand(ManageServiceCommand("start", name))
		command.AddCommand(ManageServiceCommand("restart", name))
		command.AddCommand(ManageServiceCommand("reload", name))
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
