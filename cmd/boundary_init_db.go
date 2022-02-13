package cmd

import (
	"fmt"
	"strings"

	"github.com/jsiebens/hashi-up/pkg/config"
	"github.com/jsiebens/hashi-up/pkg/operator"
	"github.com/jsiebens/hashi-up/scripts"
	"github.com/muesli/coral"
	"github.com/pkg/errors"
	"github.com/thanhpk/randstr"
)

func InitBoundaryDatabaseCommand() *coral.Command {

	var binary string
	var version string

	var configFile string

	var flags = config.BoundaryConfig{}

	var command = &coral.Command{
		Use:          "init-database",
		SilenceUsage: true,
	}

	var target = Target{}
	target.prepareCommand(command)

	command.Flags().StringVar(&binary, "package", "", "Upload and use this Boundary package instead of downloading")
	command.Flags().StringVarP(&version, "version", "v", "", "Version of Boundary to install")

	command.Flags().StringVarP(&configFile, "config-file", "c", "", "Custom Boundary configuration file to upload")

	command.Flags().StringVar(&flags.DatabaseURL, "db-url", "", "Boundary: configures the URL for connecting to Postgres")
	command.Flags().StringVar(&flags.RootKey, "root-key", "", "Boundary: a KEK (Key Encrypting Key) for the scope-specific KEKs (also referred to as the scope's root key).")

	command.RunE = func(command *coral.Command, args []string) error {
		if !target.Local && len(target.Addr) == 0 {
			return fmt.Errorf("required ssh-target-addr flag is missing")
		}

		ignoreConfigFlags := len(configFile) != 0

		var generatedConfig string

		if !ignoreConfigFlags {
			if !flags.HasDatabaseURL() {
				return fmt.Errorf("a db-url is required when initializing the database")
			}
			if !flags.HasRootKey() {
				return fmt.Errorf("a root-key when initializing the database")
			}

			generatedConfig = flags.GenerateDbConfigFile()
		}

		if len(binary) == 0 && len(version) == 0 {
			latest, err := config.GetLatestVersion("boundary")

			if err != nil {
				return errors.Wrapf(err, "unable to get latest version number, define a version manually with the --version flag")
			}

			version = latest
		}

		callback := func(op operator.CommandOperator) error {
			dir := "/tmp/hashi-up." + randstr.String(6)

			defer op.Execute("rm -rf " + dir)

			err := op.Execute("mkdir -p " + dir + "/config")
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			if len(binary) != 0 {
				info("Uploading Boundary package ...")
				err = op.UploadFile(binary, dir+"/boundary.zip", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload Boundary package: %s", err)
				}
			}

			if !ignoreConfigFlags {
				info("Uploading generated Boundary configuration ...")
				err = op.Upload(strings.NewReader(generatedConfig), dir+"/config/boundary.hcl", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload boundary configuration: %s", err)
				}
			} else {
				info(fmt.Sprintf("Uploading %s as boundary.hcl...", configFile))
				err = op.UploadFile(expandPath(configFile), dir+"/config/boundary.hcl", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload boundary configuration: %s", err)
				}
			}

			data := map[string]interface{}{
				"TmpDir":  dir,
				"Version": version,
			}

			installScript, err := scripts.RenderScript("install_boundary_db.sh", data)
			if err != nil {
				return err
			}

			err = op.Upload(installScript, dir+"/install.sh", "0755")
			if err != nil {
				return fmt.Errorf("error received during upload install script: %s", err)
			}

			info("Initializing Boundary database ...")
			sudoPass, err := target.sudoPass()
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}
			err = op.Execute(fmt.Sprintf("cat %s/install.sh | SUDO_PASS=\"%s\" sh -\n", dir, sudoPass))
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			info("Done.")

			return nil
		}

		return target.execute(callback)
	}

	return command
}
