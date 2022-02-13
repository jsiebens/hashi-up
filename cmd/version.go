package cmd

import (
	"fmt"

	"github.com/muesli/coral"
)

var (
	Version   string
	GitCommit string
)

func VersionCommand() *coral.Command {
	var command = &coral.Command{
		Use:          "version",
		Short:        "Prints the hashi-up version",
		Long:         "Prints the hashi-up version",
		SilenceUsage: true,
	}

	command.Run = func(cmd *coral.Command, args []string) {
		if len(Version) == 0 {
			fmt.Println("Version: dev")
		} else {
			fmt.Println("Version:", Version)
		}
		fmt.Println("Git Commit:", GitCommit)
	}
	return command
}
