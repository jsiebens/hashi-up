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

	var skipEnable bool
	var skipStart bool
	var binary string
	var version string

	var configFile string
	var files []string

	var flags = config.ConsulConfig{}

	var command = &cobra.Command{
		Use:          "install",
		SilenceUsage: true,
	}

	command.Flags().BoolVar(&skipEnable, "skip-enable", false, "If set to true will not enable or start Consul service")
	command.Flags().BoolVar(&skipStart, "skip-start", false, "If set to true will not start Consul service")
	command.Flags().StringVarP(&binary, "package", "p", "", "Upload and use this Consul package instead of downloading")
	command.Flags().StringVarP(&version, "version", "v", "", "Version of Consul to install")

	command.Flags().StringVarP(&configFile, "config-file", "c", "", "Custom Consul configuration file to upload")
	command.Flags().StringArrayVarP(&files, "file", "f", []string{}, "Additional files, e.g. certificates, to upload")

	command.Flags().BoolVar(&flags.Server, "server", false, "Consul: switches agent to server mode. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.Datacenter, "datacenter", "dc1", "Consul: specifies the data center of the local agent. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.BindAddr, "bind", "", "Consul: sets the bind address for cluster communication. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.AdvertiseAddr, "advertise", "", "Consul: sets the advertise address to use. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.ClientAddr, "client", "", "Consul: sets the address to bind for client access. (see Consul documentation for more info)")
	command.Flags().Int64Var(&flags.BootstrapExpect, "bootstrap-expect", 1, "Consul: sets server to expect bootstrap mode. (see Consul documentation for more info)")
	command.Flags().StringArrayVar(&flags.RetryJoin, "retry-join", []string{}, "Consul: address of an agent to join at start time with retries enabled. Can be specified multiple times. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.Encrypt, "encrypt", "", "Consul: provides the gossip encryption key. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.CaFile, "ca-file", "", "Consul: the certificate authority used to check the authenticity of client and server connections. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.CertFile, "cert-file", "", "Consul: the certificate to verify the agent's authenticity. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.KeyFile, "key-file", "", "Consul: the key used with the certificate to verify the agent's authenticity. (see Consul documentation for more info)")
	command.Flags().BoolVar(&flags.EnableConnect, "connect", false, "Consul: enables the Connect feature on the agent. (see Consul documentation for more info)")
	command.Flags().BoolVar(&flags.EnableACL, "acl", false, "Consul: enables Consul ACL system. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.AgentToken, "agent-token", "", "Consul: the token that the agent will use for internal agent operations.. (see Consul documentation for more info)")

	command.RunE = func(command *cobra.Command, args []string) error {
		if !runLocal && len(sshTargetAddr) == 0 {
			return fmt.Errorf("required ssh-target-addr flag is missing")
		}

		ignoreConfigFlags := len(configFile) != 0

		var generatedConfig string

		if !ignoreConfigFlags {
			generatedConfig = flags.GenerateConfigFile()
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

			defer op.Execute("rm -rf " + dir)

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

			if !ignoreConfigFlags {
				info("Uploading generated Consul configuration...")
				err = op.Upload(strings.NewReader(generatedConfig), dir+"/config/"+configFile, "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul configuration: %s", err)
				}

				files = []string{}

				if flags.EnableTLS() {
					files = []string{flags.CertFile, flags.CertFile, flags.KeyFile}
				}
			} else {
				info(fmt.Sprintf("Uploading %s as consul.hcl...", configFile))
				err = op.UploadFile(expandPath(configFile), dir+"/config/consul.hcl", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul configuration: %s", err)
				}
			}

			for _, s := range files {
				info(fmt.Sprintf("Uploading %s...", s))
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
			if len(flags.RetryJoin) == 0 {
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
