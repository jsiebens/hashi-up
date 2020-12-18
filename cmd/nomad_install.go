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

	var show bool
	var skipEnable bool
	var skipStart bool
	var binary string

	var version string
	var datacenter string
	var address string
	var advertise string
	var server bool
	var client bool
	var bootstrapExpect int64
	var retryJoin []string
	var encrypt string
	var caFile string
	var certFile string
	var keyFile string
	var enableACL bool

	var configFile string
	var additionalConfigFiles []string

	var command = &cobra.Command{
		Use:          "install",
		SilenceUsage: true,
	}

	command.Flags().BoolVar(&show, "show", false, "Just show the generated config instead of deploying Nomad")
	command.Flags().StringVar(&binary, "package", "", "Upload and use this Nomad package instead of downloading")
	command.Flags().BoolVar(&skipEnable, "skip-enable", false, "If set to true will not enable or start Nomad service")
	command.Flags().BoolVar(&skipStart, "skip-start", false, "If set to true will not start Nomad service")

	command.Flags().StringVar(&version, "version", "", "Version of Nomad to install, default to latest available")
	command.Flags().BoolVar(&server, "server", false, "Nomad: enables the server mode of the agent. (see Nomad documentation for more info)")
	command.Flags().BoolVar(&client, "client", false, "Nomad: enables the client mode of the agent. (see Nomad documentation for more info)")
	command.Flags().StringVar(&datacenter, "datacenter", "dc1", "Nomad: specifies the data center of the local agent. (see Nomad documentation for more info)")
	command.Flags().StringVar(&address, "address", "", "Nomad: the address the agent will bind to for all of its various network services. (see Nomad documentation for more info)")
	command.Flags().StringVar(&advertise, "advertise", "", "Nomad: the address the agent will advertise to for all of its various network services. (see Nomad documentation for more info)")
	command.Flags().Int64Var(&bootstrapExpect, "bootstrap-expect", 1, "Nomad: sets server to expect bootstrap mode. (see Nomad documentation for more info)")
	command.Flags().StringArrayVar(&retryJoin, "retry-join", []string{}, "Nomad: address of an agent to join at start time with retries enabled. Can be specified multiple times. (see Nomad documentation for more info)")
	command.Flags().StringVar(&encrypt, "encrypt", "", "Nomad: Provides the gossip encryption key. (see Nomad documentation for more info)")
	command.Flags().StringVar(&caFile, "ca-file", "", "Nomad: the certificate authority used to check the authenticity of client and server connections. (see Nomad documentation for more info)")
	command.Flags().StringVar(&certFile, "cert-file", "", "Nomad: the certificate to verify the agent's authenticity. (see Nomad documentation for more info)")
	command.Flags().StringVar(&keyFile, "key-file", "", "Nomad: the key used with the certificate to verify the agent's authenticity. (see Nomad documentation for more info)")
	command.Flags().BoolVar(&enableACL, "acl", false, "Nomad: enables Nomad ACL system. (see Nomad documentation for more info)")

	command.Flags().StringVar(&configFile, "config-file", "nomad.hcl", "Name of the generated config file")
	command.Flags().StringArrayVar(&additionalConfigFiles, "additional-config-file", []string{}, "Additional configuration file to upload")

	command.RunE = func(command *cobra.Command, args []string) error {
		if !show && !runLocal && len(sshTargetAddr) == 0 {
			return fmt.Errorf("required ssh-target-addr flag is missing")
		}

		if !(server || client) {
			return fmt.Errorf("either server or client mode should be enabled")
		}

		var enableTLS = false

		if len(caFile) != 0 && len(certFile) != 0 && len(keyFile) != 0 {
			enableTLS = true
		}

		if !enableTLS && (len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0) {
			return fmt.Errorf("ca-file, cert-file and key-file are all required when enabling tls, at least on of them is missing")
		}

		nomadConfig := config.NewNomadConfiguration(datacenter, address, advertise, server, client, bootstrapExpect, retryJoin, encrypt, enableTLS, enableACL)

		if show {
			fmt.Println(nomadConfig)
			return nil
		}

		if len(configFile) == 0 {
			return fmt.Errorf("config-file cannot be empty")
		}

		if !strings.HasSuffix(configFile, ".hcl") {
			configFile = configFile + ".hcl"
		}

		if len(binary) == 0 && len(version) == 0 {
			versions, err := config.GetVersion()

			if err != nil {
				return errors.Wrapf(err, "unable to get latest version number, define a version manually with the --version flag")
			}

			version = versions.Nomad
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

			info("Uploading Nomad configuration and certificates...")
			if enableTLS {
				err = op.UploadFile(caFile, dir+"/config/nomad-agent-ca.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload nomad ca file: %s", err)
				}

				err = op.UploadFile(certFile, dir+"/config/nomad-agent-cert.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload nomad cert file: %s", err)
				}

				err = op.UploadFile(keyFile, dir+"/config/nomad-agent-key.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload nomad key file: %s", err)
				}
			}

			err = op.Upload(strings.NewReader(nomadConfig), dir+"/config/"+configFile, "0640")
			if err != nil {
				return fmt.Errorf("error received during upload nomad configuration: %s", err)
			}

			for _, s := range additionalConfigFiles {
				_, filename := filepath.Split(expandPath(s))
				err = op.UploadFile(expandPath(s), dir+"/config/"+filename, "0640")
				if err != nil {
					return fmt.Errorf("error received during upload nomad config file: %s", err)
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
