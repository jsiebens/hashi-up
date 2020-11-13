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

func InstallConsulCommand() *cobra.Command {

	var host string
	var user string
	var sshKey string
	var sshPort int
	var local bool
	var show bool

	var version string
	var datacenter string
	var bind string
	var advertise string
	var client string
	var server bool
	var boostrapExpect int64
	var retryJoin []string
	var encrypt string
	var caFile string
	var certFile string
	var keyFile string
	var enableConnect bool
	var enableACL bool
	var agentToken string

	var command = &cobra.Command{
		Use:          "install",
		SilenceUsage: true,
	}

	command.Flags().StringVar(&host, "host", "", "Remote target host")
	command.Flags().StringVar(&user, "user", "root", "Username for SSH login")
	command.Flags().StringVar(&sshKey, "ssh-key", "", "The ssh key to use for remote login")
	command.Flags().IntVar(&sshPort, "ssh-port", 22, "The port on which to connect for ssh")
	command.Flags().BoolVar(&local, "local", false, "Running the installation locally, without ssh")
	command.Flags().BoolVar(&show, "show", false, "Just show the generated config instead of deploying Consul")

	command.Flags().StringVar(&version, "version", "", "Version of Consul to install, default to latest available")
	command.Flags().BoolVar(&server, "server", false, "Consul: switches agent to server mode. (see Consul documentation for more info)")
	command.Flags().StringVar(&datacenter, "datacenter", "dc1", "Consul: specifies the data center of the local agent. (see Consul documentation for more info)")
	command.Flags().StringVar(&bind, "bind", "", "Consul: sets the bind address for cluster communication. (see Consul documentation for more info)")
	command.Flags().StringVar(&advertise, "advertise", "", "Consul: sets the advertise address to use. (see Consul documentation for more info)")
	command.Flags().StringVar(&client, "client", "", "Consul: sets the address to bind for client access. (see Consul documentation for more info)")
	command.Flags().Int64Var(&boostrapExpect, "bootstrap-expect", 1, "Consul: sets server to expect bootstrap mode. (see Consul documentation for more info)")
	command.Flags().StringArrayVar(&retryJoin, "retry-join", []string{}, "Consul: address of an agent to join at start time with retries enabled. Can be specified multiple times. (see Consul documentation for more info)")
	command.Flags().StringVar(&encrypt, "encrypt", "", "Consul: provides the gossip encryption key. (see Consul documentation for more info)")
	command.Flags().StringVar(&caFile, "ca-file", "", "Consul: the certificate authority used to check the authenticity of client and server connections. (see Consul documentation for more info)")
	command.Flags().StringVar(&certFile, "cert-file", "", "Consul: the certificate to verify the agent's authenticity. (see Consul documentation for more info)")
	command.Flags().StringVar(&keyFile, "key-file", "", "Consul: the key used with the certificate to verify the agent's authenticity. (see Consul documentation for more info)")
	command.Flags().BoolVar(&enableConnect, "connect", false, "Consul: enables the Connect feature on the agent. (see Consul documentation for more info)")
	command.Flags().BoolVar(&enableACL, "acl", false, "Consul: enables Consul ACL system. (see Consul documentation for more info)")
	command.Flags().StringVar(&agentToken, "agent-token", "", "Consul: the token that the agent will use for internal agent operations.. (see Consul documentation for more info)")

	command.RunE = func(command *cobra.Command, args []string) error {
		if !show && !local && len(host) == 0 {
			return fmt.Errorf("required host flag is missing")
		}

		var enableTLS = false

		if len(caFile) != 0 && len(certFile) != 0 && len(keyFile) != 0 {
			enableTLS = true
		}

		if !enableTLS && (len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0) {
			return fmt.Errorf("ca-file, cert-file and key-file are all required when enabling tls, at least on of them is missing")
		}

		consulConfig := config.NewConsulConfiguration(datacenter, bind, advertise, client, server, boostrapExpect, retryJoin, encrypt, enableTLS, enableACL, agentToken, enableConnect)

		if show {
			fmt.Println(consulConfig)
			return nil
		}

		if len(version) == 0 {
			updateParams := &checkpoint.CheckParams{
				Product: "consul",
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
			dir := "/tmp/consul-installation." + randstr.String(6)

			defer op.Execute("rm -rf " + dir)

			_, err := op.Execute("mkdir " + dir)
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			if enableTLS {
				err = op.UploadFile(caFile, dir+"/consul-agent-ca.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul ca file: %s", err)
				}

				err = op.UploadFile(certFile, dir+"/consul-agent-cert.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul cert file: %s", err)
				}

				err = op.UploadFile(keyFile, dir+"/consul-agent-key.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul key file: %s", err)
				}
			}

			err = op.Upload(strings.NewReader(consulConfig), dir+"/consul.hcl", "0640")
			if err != nil {
				return fmt.Errorf("error received during upload consul configuration: %s", err)
			}

			installScript, err := pkger.Open("/scripts/install_consul.sh")

			if err != nil {
				return err
			}

			defer installScript.Close()

			err = op.Upload(installScript, dir+"/install.sh", "0755")
			if err != nil {
				return fmt.Errorf("error received during upload install script: %s", err)
			}

			var serviceType = "notify"
			if len(retryJoin) == 0 {
				serviceType = "exec"
			}

			_, err = op.Execute(fmt.Sprintf("cat %s/install.sh | TMP_DIR='%s' SERVICE_TYPE='%s' CONSUL_VERSION='%s' sh -\n", dir, dir, serviceType, version))
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			return nil
		}

		if local {
			return operator.ExecuteLocal(callback)
		} else {
			if sshKey == "" {
				return operator.ExecuteRemote(host, sshPort, user, callback)
			} else {
				return operator.ExecuteRemoteWithPrivateKey(host, sshPort, user, sshKey, callback)
			}
		}
	}

	return command
}
