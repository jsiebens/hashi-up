package cmd

import "github.com/spf13/cobra"

type Target struct {
	Addr  string
	User  string
	Key   string
	Local bool
}

func (t *Target) prepareCommand(cmd *cobra.Command) {
	cmd.Flags().StringVar(&t.Addr, "ssh-target-addr", "", "Remote SSH target address (e.g. 127.0.0.1:22")
	cmd.Flags().StringVar(&t.User, "ssh-target-user", "root", "Username for SSH login")
	cmd.Flags().StringVar(&t.Key, "ssh-target-key", "", "The ssh key to use for SSH login")
	cmd.Flags().BoolVar(&t.Local, "local", false, "Running the installation locally, without ssh")
}
