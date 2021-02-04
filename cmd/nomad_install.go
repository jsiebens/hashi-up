package cmd

import (
	"fmt"
	"github.com/jsiebens/hashi-up/pkg/config"
	"github.com/jsiebens/hashi-up/pkg/operator"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/thanhpk/randstr"
	"path/filepath"
	"strings"
)

func InstallNomadCommand() *cobra.Command {

	var skipEnable bool
	var skipStart bool
	var binary string
	var version string

	var configFile string
	var files []string

	var flags = config.NomadConfig{}

	var command = &cobra.Command{
		Use:          "install",
		SilenceUsage: true,
	}

	var target = Target{}
	target.prepareCommand(command)

	command.Flags().BoolVar(&skipEnable, "skip-enable", false, "If set to true will not enable or start Nomad service")
	command.Flags().BoolVar(&skipStart, "skip-start", false, "If set to true will not start Nomad service")
	command.Flags().StringVarP(&binary, "package", "p", "", "Upload and use this Nomad package instead of downloading")
	command.Flags().StringVarP(&version, "version", "v", "", "Version of Nomad to install")

	command.Flags().StringVarP(&configFile, "config-file", "c", "", "Custom Nomad configuration file to upload, setting this will disable config file generation meaning the other flags are ignored")
	command.Flags().StringArrayVarP(&files, "file", "f", []string{}, "Additional files, e.g. certificates, to upload")

	command.Flags().BoolVar(&flags.Server, "server", false, "Nomad: enables the server mode of the agent. (see Nomad documentation for more info)")
	command.Flags().BoolVar(&flags.Client, "client", false, "Nomad: enables the client mode of the agent. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.Datacenter, "datacenter", "dc1", "Nomad: specifies the data center of the local agent. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.BindAddr, "address", "", "Nomad: the address the agent will bind to for all of its various network services. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.AdvertiseAddr, "advertise", "", "Nomad: the address the agent will advertise to for all of its various network services. (see Nomad documentation for more info)")
	command.Flags().Int64Var(&flags.BootstrapExpect, "bootstrap-expect", 1, "Nomad: sets server to expect bootstrap mode. 0 are less disables bootstrap mode. (see Nomad documentation for more info)")
	command.Flags().StringArrayVar(&flags.RetryJoin, "retry-join", []string{}, "Nomad: address of an agent to join at start time with retries enabled. Can be specified multiple times. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.Encrypt, "encrypt", "", "Nomad: Provides the gossip encryption key. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.CaFile, "ca-file", "", "Nomad: the certificate authority used to check the authenticity of client and server connections. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.CertFile, "cert-file", "", "Nomad: the certificate to verify the agent's authenticity. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.KeyFile, "key-file", "", "Nomad: the key used with the certificate to verify the agent's authenticity. (see Nomad documentation for more info)")
	command.Flags().BoolVar(&flags.EnableACL, "acl", false, "Nomad: enables Nomad ACL system. (see Nomad documentation for more info)")

	command.RunE = func(command *cobra.Command, args []string) error {
		if !target.Local && len(target.Addr) == 0 {
			return fmt.Errorf("required ssh-target-addr flag is missing")
		}

		ignoreConfigFlags := len(configFile) != 0

		var generatedConfig string

		if !ignoreConfigFlags {
			generatedConfig = flags.GenerateConfigFile()
		}

		if len(binary) == 0 && len(version) == 0 {
			latest, err := config.GetLatestVersion("nomad")

			if err != nil {
				return errors.Wrapf(err, "unable to get latest version number, define a version manually with the --version flag")
			}

			version = latest
		}

		callback := func(op operator.CommandOperator) error {
			dir := "/tmp/hashi-up." + randstr.String(6)

			defer op.Execute("rm -rf " + dir)

			_, err := op.Execute("mkdir -p " + dir + "/config")
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			if len(binary) != 0 {
				info("Uploading Nomad package ...")
				err = op.UploadFile(binary, dir+"/nomad.zip", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload nomad package: %s", err)
				}
			}

			if !ignoreConfigFlags {
				info("Uploading generated Nomad configuration ...")
				err = op.Upload(strings.NewReader(generatedConfig), dir+"/config/nomad.hcl", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload nomad configuration: %s", err)
				}

				files = []string{}

				if flags.EnableTLS() {
					files = []string{flags.CaFile, flags.KeyFile, flags.CertFile}
				}
			} else {
				info(fmt.Sprintf("Uploading %s as nomad.hcl...", configFile))
				err = op.UploadFile(expandPath(configFile), dir+"/config/nomad.hcl", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload nomad configuration: %s", err)
				}
			}

			for _, s := range files {
				if len(s) != 0 {
					info(fmt.Sprintf("Uploading %s...", s))
					_, filename := filepath.Split(expandPath(s))
					err = op.UploadFile(expandPath(s), dir+"/config/"+filename, "0640")
					if err != nil {
						return fmt.Errorf("error received during upload file: %s", err)
					}
				}
			}

			installScript, err := pkger.Open("/scripts/install_nomad.sh")

			if err != nil {
				return err
			}

			defer installScript.Close()

			err = op.Upload(installScript, dir+"/install.sh", "0755")
			if err != nil {
				return fmt.Errorf("error received during upload install script: %s", err)
			}

			info("Installing Nomad ...")
			_, err = op.Execute(fmt.Sprintf("cat %s/install.sh | TMP_DIR='%s' NOMAD_VERSION='%s' SKIP_ENABLE='%t' SKIP_START='%t' sh -\n", dir, dir, version, skipEnable, skipStart))
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			info("Done.")

			return nil
		}

		if target.Local {
			return operator.ExecuteLocal(callback)
		} else {
			return operator.ExecuteRemote(target.Addr, target.User, target.Key, callback)
		}
	}

	return command
}
