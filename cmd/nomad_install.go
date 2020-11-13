package cmd

import (
	"fmt"
	"github.com/hashicorp/go-checkpoint"
	"github.com/jsiebens/hashi-up/pkg/config"
	"github.com/jsiebens/operator"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/thanhpk/randstr"
	"strings"
)

func InstallNomadCommand() *cobra.Command {

	var ip string
	var user string
	var sshKey string
	var sshPort int
	var local bool
	var show bool

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

	var command = &cobra.Command{
		Use:          "install",
		SilenceUsage: true,
	}

	command.Flags().StringVar(&ip, "ip", "127.0.0.1", "Public IP of node")
	command.Flags().StringVar(&user, "user", "root", "Username for SSH login")
	command.Flags().StringVar(&sshKey, "ssh-key", "", "The ssh key to use for remote login")
	command.Flags().IntVar(&sshPort, "ssh-port", 22, "The port on which to connect for ssh")
	command.Flags().BoolVar(&local, "local", false, "Running the installation locally, without ssh")
	command.Flags().BoolVar(&show, "show", false, "Just show the generated config instead of deploying Consul")

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

	command.RunE = func(command *cobra.Command, args []string) error {
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

		if len(version) == 0 {
			updateParams := &checkpoint.CheckParams{
				Product: "nomad",
				Version: "0.0.0",
				Force:   true,
			}

			check, err := checkpoint.Check(updateParams)

			if err != nil {
				return errors.Wrapf(err, "unable to get latest version number, define a version manually with the --version flag")
			}

			version = check.CurrentVersion
		}

		callback := func(op operator.CommandOperator) error {
			dir := "/tmp/nomad-installation." + randstr.String(6)

			defer op.Execute("rm -rf " + dir)

			_, err := op.Execute("mkdir " + dir)
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			if enableTLS {
				err = op.UploadFile(caFile, dir+"/nomad-agent-ca.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload nomad ca file: %s", err)
				}

				err = op.UploadFile(certFile, dir+"/nomad-agent-cert.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload nomad cert file: %s", err)
				}

				err = op.UploadFile(keyFile, dir+"/nomad-agent-key.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload nomad key file: %s", err)
				}
			}

			err = op.Upload(strings.NewReader(nomadConfig), dir+"/nomad.hcl", "0640")
			if err != nil {
				return fmt.Errorf("error received during upload nomad configuration: %s", err)
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

			_, err = op.Execute(fmt.Sprintf("cat %s/install.sh | TMP_DIR='%s' NOMAD_VERSION='%s' sh -\n", dir, dir, version))
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			return nil
		}

		if local {
			return operator.ExecuteLocal(callback)
		} else {
			if sshKey == "" {
				return operator.ExecuteRemote(ip, sshPort, user, callback)
			} else {
				return operator.ExecuteRemoteWithPrivateKey(ip, sshPort, user, sshKey, callback)
			}
		}
	}

	return command
}
