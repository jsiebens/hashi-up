package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jsiebens/hashi-up/pkg/config"
	"github.com/jsiebens/hashi-up/pkg/operator"
	"github.com/jsiebens/hashi-up/scripts"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/thanhpk/randstr"
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

	var target = Target{}
	target.prepareCommand(command)

	command.Flags().BoolVar(&skipEnable, "skip-enable", false, "If set to true will not enable or start Consul service")
	command.Flags().BoolVar(&skipStart, "skip-start", false, "If set to true will not start Consul service")
	command.Flags().StringVar(&binary, "package", "", "Upload and use this Consul package instead of downloading")
	command.Flags().StringVarP(&version, "version", "v", "", "Version of Consul to install")

	command.Flags().StringVarP(&configFile, "config-file", "c", "", "Custom Consul configuration file to upload, setting this will disable config file generation meaning the other flags are ignored")
	command.Flags().StringSliceVarP(&files, "file", "f", []string{}, "Additional files, e.g. certificates, to upload")

	command.Flags().BoolVar(&flags.Server, "server", false, "Consul: switches agent to server mode. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.Datacenter, "datacenter", "dc1", "Consul: specifies the data center of the local agent. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.BindAddr, "bind-addr", "", "Consul: sets the bind address for cluster communication. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.AdvertiseAddr, "advertise-addr", "", "Consul: sets the advertise address to use. (see Consul documentation for more info)")

	command.Flags().StringVar(&flags.ClientAddr, "client-addr", "", "Consul: sets the address to bind for client access. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.DnsAddr, "dns-addr", "", "Consul: sets the address for the DNS server. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.HttpAddr, "http-addr", "", "Consul: sets the address for the HTTP API server. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.HttpsAddr, "https-addr", "", "Consul: sets the address for the HTTPS API server. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.GrpcAddr, "grpc-addr", "", "Consul: sets the address for the gRPC API server. (see Consul documentation for more info)")

	command.Flags().Int64Var(&flags.BootstrapExpect, "bootstrap-expect", 1, "Consul: sets server to expect bootstrap mode. 0 are less disables bootstrap mode. (see Consul documentation for more info)")
	command.Flags().StringSliceVar(&flags.RetryJoin, "retry-join", []string{}, "Consul: address of an agent to join at start time with retries enabled. Can be specified multiple times. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.Encrypt, "encrypt", "", "Consul: provides the gossip encryption key. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.CaFile, "ca-file", "", "Consul: the certificate authority used to check the authenticity of client and server connections. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.CertFile, "cert-file", "", "Consul: the certificate to verify the agent's authenticity. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.KeyFile, "key-file", "", "Consul: the key used with the certificate to verify the agent's authenticity. (see Consul documentation for more info)")
	command.Flags().BoolVar(&flags.AutoEncrypt, "auto-encrypt", false, "Consul: this option enables auto_encrypt and allows servers to automatically distribute certificates from the Connect CA to the clients. (see Consul documentation for more info)")
	command.Flags().BoolVar(&flags.HttpsOnly, "https-only", true, "Consul: if true, HTTP port is disabled on both clients and servers and to only accept HTTPS connections when TLS enabled.")
	command.Flags().BoolVar(&flags.EnableConnect, "connect", false, "Consul: enables the Connect feature on the agent. (see Consul documentation for more info)")
	command.Flags().BoolVar(&flags.EnableACL, "acl", false, "Consul: enables Consul ACL system. (see Consul documentation for more info)")
	command.Flags().StringVar(&flags.AgentToken, "agent-token", "", "Consul: the token that the agent will use for internal agent operations.. (see Consul documentation for more info)")

	command.Flags().StringVar(&flags.ClientAddr, "client", "", "")
	command.Flags().StringVar(&flags.BindAddr, "bind", "", "")
	command.Flags().StringVar(&flags.AdvertiseAddr, "advertise", "", "")
	_ = command.Flags().MarkDeprecated("client", "use the new flag client-addr")
	_ = command.Flags().MarkDeprecated("bind", "use the new flag bind-addr")
	_ = command.Flags().MarkDeprecated("advertise", "use the new flag advertise-addr")

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
			latest, err := config.GetLatestVersion("consul")

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
				info("Uploading Consul package ...")
				err = op.UploadFile(binary, dir+"/consul.zip", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload Consul package: %s", err)
				}
			}

			if !ignoreConfigFlags {
				info("Uploading generated Consul configuration ...")
				err = op.Upload(strings.NewReader(generatedConfig), dir+"/config/consul.hcl", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul configuration: %s", err)
				}

				files = []string{}

				if flags.EnableTLS() {
					files = []string{flags.CaFile, flags.CertFile, flags.KeyFile}
				}
			} else {
				info(fmt.Sprintf("Uploading %s as consul.hcl...", configFile))
				err = op.UploadFile(expandPath(configFile), dir+"/config/consul.hcl", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul configuration: %s", err)
				}
			}

			for _, s := range files {
				if len(s) != 0 {
					info(fmt.Sprintf("Uploading %s...", s))
					_, filename := filepath.Split(expandPath(s))
					err = op.UploadFile(expandPath(s), dir+"/config/"+filename, "0640")
					if err != nil {
						return fmt.Errorf("error received during upload consul ca file: %s", err)
					}
				}
			}

			installScript, err := scripts.Open("install_consul.sh")

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

			info("Installing Consul ...")
			sudoPass, err := target.sudoPass()
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}
			_, err = op.Execute(fmt.Sprintf("cat %s/install.sh | SUDO_PASS=\"%s\" TMP_DIR='%s' SERVICE_TYPE='%s' CONSUL_VERSION='%s' SKIP_ENABLE='%t' SKIP_START='%t' sh -\n", dir, sudoPass, dir, serviceType, version, skipEnable, skipStart))
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
