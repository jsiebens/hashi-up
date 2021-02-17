package cmd

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/jsiebens/hashi-up/pkg/operator"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const SshTargetPassword = "SSH_TARGET_PASSWORD"
const SshTargetSudoPass = "SSH_TARGET_SUDO_PASS"

type Target struct {
	Addr     string
	User     string
	Key      string
	Password string
	SudoPass string
	Local    bool
}

func (t *Target) prepareCommand(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&t.Addr, "ssh-target-addr", "r", "", "Remote SSH target address (e.g. 127.0.0.1:22")
	cmd.Flags().StringVarP(&t.User, "ssh-target-user", "u", "root", "Username for SSH login")
	cmd.Flags().StringVarP(&t.Key, "ssh-target-key", "k", "", "The ssh key to use for SSH login")
	cmd.Flags().StringVarP(&t.Password, "ssh-target-password", "p", "", "The ssh password to use for SSH login")
	cmd.Flags().StringVarP(&t.SudoPass, "ssh-target-sudo-pass", "s", "", "The ssh password to use for SSH login")
	cmd.Flags().BoolVar(&t.Local, "local", false, "Running the installation locally, without ssh")
}

func (t *Target) execute(callback operator.Callback) error {
	if t.Local {
		return operator.ExecuteLocal(callback)
	} else {
		pwd, err := pathOrContents(getenv(SshTargetPassword, t.Password))
		if err != nil {
			return err
		}
		return operator.ExecuteRemote(t.Addr, t.User, t.Key, pwd, callback)
	}
}

func (t *Target) sudoPass() (string, error) {
	sudoPass := getenv(SshTargetSudoPass, t.SudoPass)
	if len(sudoPass) != 0 {
		pwd, err := pathOrContents(sudoPass)
		if err != nil {
			return "", err
		} else {
			return pwd, nil
		}
	} else {
		pwd, err := pathOrContents(getenv(SshTargetPassword, t.Password))
		if err != nil {
			return "", err
		} else {
			return pwd, nil
		}
	}
}

func pathOrContents(poc string) (string, error) {
	if len(poc) == 0 {
		return poc, nil
	}

	path := poc
	if path[0] == '~' {
		var err error
		path, err = homedir.Expand(path)
		if err != nil {
			return path, err
		}
	}

	if _, err := os.Stat(path); err == nil {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return string(contents), err
		}
		return strings.TrimSpace(string(contents)), nil
	}

	return poc, nil
}

func getenv(key, override string) string {
	if len(override) != 0 {
		return override
	}
	value := os.Getenv(key)
	if len(value) != 0 {
		return value
	}
	return ""
}
