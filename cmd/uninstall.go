package cmd

import (
	"fmt"
	"strings"

	"github.com/jsiebens/hashi-up/pkg/operator"
	"github.com/jsiebens/hashi-up/scripts"
	"github.com/muesli/coral"
	"github.com/thanhpk/randstr"
)

func UninstallCommand(product string) *coral.Command {

	var command = &coral.Command{
		Use:          "uninstall",
		Short:        fmt.Sprintf("Uninstall %s on a server via SSH", strings.Title(product)),
		Long:         fmt.Sprintf("Uninstall %s on a server via SSH", strings.Title(product)),
		SilenceUsage: true,
	}

	var target = Target{}
	target.prepareCommand(command)

	command.RunE = func(command *coral.Command, args []string) error {
		if !target.Local && len(target.Addr) == 0 {
			return fmt.Errorf("required ssh-target-addr flag is missing")
		}

		callback := func(op operator.CommandOperator) error {
			dir := "/tmp/hashi-up." + randstr.String(6)

			defer op.Execute("rm -rf " + dir)

			err := op.Execute("mkdir -p " + dir + "/config")
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			installScript, err := scripts.Open("uninstall.sh")

			if err != nil {
				return err
			}

			defer installScript.Close()

			err = op.Upload(installScript, dir+"/run.sh", "0755")
			if err != nil {
				return fmt.Errorf("error received during upload install script: %s", err)
			}

			info(fmt.Sprintf("Uninstalling %s ...", strings.Title(product)))
			sudoPass, err := target.sudoPass()
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}
			err = op.Execute(fmt.Sprintf("cat %s/run.sh | SERVICE=%s SUDO_PASS=\"%s\" sh -\n", dir, product, sudoPass))
			if err != nil {
				return fmt.Errorf("error received during uninstallation: %s", err)
			}

			info("Done.")

			return nil
		}

		return target.execute(callback)
	}

	return command
}
