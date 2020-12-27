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

	var ignoreConfigFlags bool
	var skipEnable bool
	var skipStart bool
	var binary string
	var version string

	var generatedConfigFile string
	var configFiles []string

	var flags = config.NomadConfig{}

	var command = &cobra.Command{
		Use:          "install",
		SilenceUsage: true,
	}

	command.Flags().BoolVarP(&ignoreConfigFlags, "ignore-config-flags", "i", false, "If set to false will generate a configuration file based on CLI flags, otherwise the flags are ignored")
	command.Flags().BoolVar(&skipEnable, "skip-enable", false, "If set to true will not enable or start Nomad service")
	command.Flags().BoolVar(&skipStart, "skip-start", false, "If set to true will not start Nomad service")
	command.Flags().StringVarP(&binary, "package", "p", "", "Upload and use this Nomad package instead of downloading")
	command.Flags().StringVarP(&version, "version", "v", "", "Version of Nomad to install")

	command.Flags().StringVarP(&generatedConfigFile, "generated-config-file", "c", "nomad.hcl", "Name of the generated config file")
	command.Flags().StringArrayVarP(&configFiles, "file", "f", []string{}, "Additional configuration file to upload")

	command.Flags().BoolVar(&flags.Server, "server", false, "Nomad: enables the server mode of the agent. (see Nomad documentation for more info)")
	command.Flags().BoolVar(&flags.Client, "client", false, "Nomad: enables the client mode of the agent. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.Datacenter, "datacenter", "dc1", "Nomad: specifies the data center of the local agent. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.BindAddr, "address", "", "Nomad: the address the agent will bind to for all of its various network services. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.AdvertiseAddr, "advertise", "", "Nomad: the address the agent will advertise to for all of its various network services. (see Nomad documentation for more info)")
	command.Flags().Int64Var(&flags.BootstrapExpect, "bootstrap-expect", 1, "Nomad: sets server to expect bootstrap mode. (see Nomad documentation for more info)")
	command.Flags().StringArrayVar(&flags.RetryJoin, "retry-join", []string{}, "Nomad: address of an agent to join at start time with retries enabled. Can be specified multiple times. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.Encrypt, "encrypt", "", "Nomad: Provides the gossip encryption key. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.CaFile, "ca-file", "", "Nomad: the certificate authority used to check the authenticity of client and server connections. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.CertFile, "cert-file", "", "Nomad: the certificate to verify the agent's authenticity. (see Nomad documentation for more info)")
	command.Flags().StringVar(&flags.KeyFile, "key-file", "", "Nomad: the key used with the certificate to verify the agent's authenticity. (see Nomad documentation for more info)")
	command.Flags().BoolVar(&flags.EnableACL, "acl", false, "Nomad: enables Nomad ACL system. (see Nomad documentation for more info)")

	command.RunE = func(command *cobra.Command, args []string) error {
		if !runLocal && len(sshTargetAddr) == 0 {
			return fmt.Errorf("required ssh-target-addr flag is missing")
		}

		var generatedConfig string

		if !ignoreConfigFlags {
			generatedConfig = flags.GenerateConfigFile()
			if !strings.HasSuffix(generatedConfigFile, ".hcl") {
				generatedConfigFile = generatedConfigFile + ".hcl"
			}
		}

		if len(binary) == 0 && len(version) == 0 {
			latest, err := config.GetLatestVersion("nomad")

			if err != nil {
				return errors.Wrapf(err, "unable to get latest version number, define a version manually with the --version flag")
			}

			version = latest
		}

		callback := func(op operator.CommandOperator) error {
			dir := "/tmp/nomad-installation." + randstr.String(6)

			defer op.Execute("rm -rf " + dir)

			_, err := op.Execute("mkdir -p " + dir + "/config")
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			if len(binary) != 0 {
				info("Uploading Nomad package...")
				err = op.UploadFile(binary, dir+"/nomad.zip", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload nomad package: %s", err)
				}
			}

			if !ignoreConfigFlags {
				info("Uploading generated Nomad configuration...")
				err = op.Upload(strings.NewReader(generatedConfig), dir+"/config/"+generatedConfigFile, "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul configuration: %s", err)
				}

				if flags.EnableTLS() {
					configFiles = append([]string{flags.CaFile, flags.KeyFile, flags.CertFile}, configFiles...)
				}
			}

			for _, s := range configFiles {
				info(fmt.Sprintf("Uploading %s...", s))
				_, filename := filepath.Split(expandPath(s))
				err = op.UploadFile(expandPath(s), dir+"/config/"+filename, "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul ca file: %s", err)
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

			info("Installing Nomad...")
			_, err = op.Execute(fmt.Sprintf("cat %s/install.sh | TMP_DIR='%s' NOMAD_VERSION='%s' SKIP_ENABLE='%t' SKIP_START='%t' sh -\n", dir, dir, version, skipEnable, skipStart))
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

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
