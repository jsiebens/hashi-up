package cmd

import (
	"fmt"
	"github.com/jsiebens/hashi-up/pkg/operator"
	"github.com/markbates/pkger"
	"github.com/spf13/cobra"
	"github.com/thanhpk/randstr"
)

func UninstallNomadCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:          "uninstall",
		SilenceUsage: true,
	}

	command.RunE = func(command *cobra.Command, args []string) error {
		callback := func(op operator.CommandOperator) error {
			dir := "/tmp/hashi-up." + randstr.String(6)

			defer op.Execute("rm -rf " + dir)

			_, err := op.Execute("mkdir -p " + dir + "/config")
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			installScript, err := pkger.Open("/scripts/uninstall_nomad.sh")

			if err != nil {
				return err
			}

			defer installScript.Close()

			err = op.Upload(installScript, dir+"/uninstall.sh", "0755")
			if err != nil {
				return fmt.Errorf("error received during upload install script: %s", err)
			}

			info("Uninstalling Nomad ...")
			_, err = op.Execute(fmt.Sprintf("cat %s/uninstall.sh | sh -\n", dir))
			if err != nil {
				return fmt.Errorf("error received during uninstallation: %s", err)
			}

			info("Done.")

			return nil
		}

		if runLocal {
			return operator.ExecuteLocal(callback)
		} else {
			return operator.ExecuteRemote(sshTargetAddr, sshTargetUser, sshTargetKey, callback)
		}
	}

	return command
}
