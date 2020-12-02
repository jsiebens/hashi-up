package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   string
	GitCommit string
)

func VersionCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:          "version",
		SilenceUsage: true,
	}

	command.Run = func(cmd *cobra.Command, args []string) {
		if len(Version) == 0 {
			fmt.Println("Version: dev")
		} else {
			fmt.Println("Version:", Version)
		}
		fmt.Println("Git Commit:", GitCommit)
	}
	return command
}
