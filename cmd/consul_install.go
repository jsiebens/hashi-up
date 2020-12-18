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

func InstallConsulCommand() *cobra.Command {

	var show bool
	var skipEnable bool
	var skipStart bool
	var binary string

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

	var configFile string
	var additionalConfigFiles []string

	var command = &cobra.Command{
		Use:          "install",
		SilenceUsage: true,
	}

	command.Flags().BoolVar(&show, "show", false, "Just show the generated config instead of deploying Consul")
	command.Flags().StringVar(&binary, "package", "", "Upload and use this Consul package instead of downloading")
	command.Flags().BoolVar(&skipEnable, "skip-enable", false, "If set to true will not enable or start Consul service")
	command.Flags().BoolVar(&skipStart, "skip-start", false, "If set to true will not start Consul service")

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

	command.Flags().StringVar(&configFile, "config-file", "consul.hcl", "Name of the generated config file")
	command.Flags().StringArrayVar(&additionalConfigFiles, "additional-config-file", []string{}, "Additional configuration file to upload")

	command.RunE = func(command *cobra.Command, args []string) error {
		if !show && !runLocal && len(sshTargetAddr) == 0 {
			return fmt.Errorf("required ssh-target-addr flag is missing")
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

		if len(configFile) == 0 {
			return fmt.Errorf("config-file cannot be empty")
		}

		if !strings.HasSuffix(configFile, ".hcl") {
			configFile = configFile + ".hcl"
		}

		if len(binary) == 0 && len(version) == 0 {
			latest, err := config.GetLatestVersion("consul")

			if err != nil {
				return errors.Wrapf(err, "unable to get latest version number, define a version manually with the --version flag")
			}

			version = latest
		}

		callback := func(op operator.CommandOperator) error {
			dir := "/tmp/consul-installation." + randstr.String(6)

			//defer op.Execute("rm -rf " + dir)

			_, err := op.Execute("mkdir -p " + dir + "/config")
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			if len(binary) != 0 {
				info("Uploading Consul package...")
				err = op.UploadFile(binary, dir+"/consul.zip", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload Consul package: %s", err)
				}
			}

			info("Uploading Consul configuration and certificates...")
			if enableTLS {
				err = op.UploadFile(caFile, dir+"/config/consul-agent-ca.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul ca file: %s", err)
				}

				err = op.UploadFile(certFile, dir+"/config/consul-agent-cert.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul cert file: %s", err)
				}

				err = op.UploadFile(keyFile, dir+"/config/consul-agent-key.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul key file: %s", err)
				}
			}

			err = op.Upload(strings.NewReader(consulConfig), dir+"/config/"+configFile, "0640")
			if err != nil {
				return fmt.Errorf("error received during upload consul configuration: %s", err)
			}

			for _, s := range additionalConfigFiles {
				_, filename := filepath.Split(expandPath(s))
				err = op.UploadFile(expandPath(s), dir+"/config/"+filename, "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul ca file: %s", err)
				}
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

			info("Installing Consul...")
			_, err = op.Execute(fmt.Sprintf("cat %s/install.sh | TMP_DIR='%s' SERVICE_TYPE='%s' CONSUL_VERSION='%s' SKIP_ENABLE='%t' SKIP_START='%t' sh -\n", dir, dir, serviceType, version, skipEnable, skipStart))
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
