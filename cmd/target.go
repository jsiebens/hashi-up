package cmd

import (
	"github.com/jsiebens/hashi-up/pkg/operator"
	"github.com/spf13/cobra"
)

type Target struct {
	Addr     string
	User     string
	Key      string
	Password string
	Local    bool
}

func (t *Target) prepareCommand(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&t.Addr, "ssh-target-addr", "r", "", "Remote SSH target address (e.g. 127.0.0.1:22")
	cmd.Flags().StringVarP(&t.User, "ssh-target-user", "u", "root", "Username for SSH login")
	cmd.Flags().StringVarP(&t.Key, "ssh-target-key", "k", "", "The ssh key to use for SSH login")
	cmd.Flags().StringVarP(&t.Password, "ssh-target-password", "s", "", "The ssh password to use for SSH login")
	cmd.Flags().BoolVar(&t.Local, "local", false, "Running the installation locally, without ssh")
}

func (t *Target) execute(callback operator.Callback) error {
	if t.Local {
		return operator.ExecuteLocal(callback)
	} else {
		return operator.ExecuteRemote(t.Addr, t.User, t.Key, t.Password, callback)
	}
}
