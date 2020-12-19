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

func InstallVaultCommand() *cobra.Command {

	var show bool
	var skipEnable bool
	var skipStart bool
	var binary string

	var certFile string
	var keyFile string
	var address []string
	var apiAddr string
	var clusterAddr string

	var storage string

	var consulAddr string
	var consulPath string
	var consulToken string
	var consulCaFile string
	var consulCertFile string
	var consulKeyFile string

	var version string
	var configFile string
	var additionalConfigFiles []string

	var command = &cobra.Command{
		Use:          "install",
		SilenceUsage: true,
	}

	command.Flags().BoolVar(&show, "show", false, "Just show the generated config instead of deploying Vault")
	command.Flags().StringVar(&binary, "package", "", "Upload and use this Vault package instead of downloading")
	command.Flags().BoolVar(&skipEnable, "skip-enable", false, "If set to true will not enable or start Vault service")
	command.Flags().BoolVar(&skipStart, "skip-start", false, "If set to true will not start Vault service")

	command.Flags().StringVar(&version, "version", "", "Version of Vault to install")
	command.Flags().StringVar(&certFile, "cert-file", "", "Vault: the certificate for TLS. (see Vault documentation for more info)")
	command.Flags().StringVar(&keyFile, "key-file", "", "Vault: the private key for the certificate. (see Vault documentation for more info)")
	command.Flags().StringArrayVar(&address, "address", []string{"0.0.0.0:8200"}, "Vault: the address to bind to for listening. (see Vault documentation for more info)")
	command.Flags().StringVar(&apiAddr, "api-addr", "", "Vault: the address (full URL) to advertise to other Vault servers in the cluster for client redirection. (see Vault documentation for more info)")
	command.Flags().StringVar(&clusterAddr, "cluster-addr", "", "Vault: the address to advertise to other Vault servers in the cluster for request forwarding. (see Vault documentation for more info)")

	command.Flags().StringVar(&storage, "storage", "file", "Vault: the type of storage backend. Currently only \"file\" of \"consul\" is supported. (see Vault documentation for more info)")

	command.Flags().StringVar(&consulAddr, "consul-addr", "127.0.0.1:8500", "Vault: the address of the Consul agent to communicate with. (see Vault documentation for more info)")
	command.Flags().StringVar(&consulPath, "consul-path", "vault/", "Vault: the path in Consul's key-value store where Vault data will be stored. (see Vault documentation for more info)")
	command.Flags().StringVar(&consulToken, "consul-token", "", "Vault: the Consul ACL token with permission to read and write from the path in Consul's key-value store. (see Vault documentation for more info)")

	command.Flags().StringVar(&consulCaFile, "consul-tls-ca-file", "", "Vault: the path to the CA certificate used for Consul communication. (see Vault documentation for more info)")
	command.Flags().StringVar(&consulCertFile, "consul-tls-cert-file", "", "Vault: the path to the certificate for Consul communication. (see Vault documentation for more info)")
	command.Flags().StringVar(&consulKeyFile, "consul-tls-key-file", "", "Vault: the path to the private key for Consul communication. (see Vault documentation for more info)")

	command.Flags().StringVar(&configFile, "config-file", "vault.hcl", "Name of the generated config file")
	command.Flags().StringArrayVar(&additionalConfigFiles, "additional-config-file", []string{}, "Additional configuration file to upload")

	command.RunE = func(command *cobra.Command, args []string) error {
		if !show && !runLocal && len(sshTargetAddr) == 0 {
			return fmt.Errorf("required ssh-target-addr flag is missing")
		}

		var enableTLS = false
		var enableConsulTLS = false

		if len(certFile) != 0 && len(keyFile) != 0 {
			enableTLS = true
		}

		if len(consulCaFile) != 0 && len(consulCertFile) != 0 && len(consulKeyFile) != 0 {
			enableConsulTLS = true
		}

		vaultConfig := config.NewVaultConfiguration(apiAddr, clusterAddr, address, enableTLS, storage, consulAddr, consulPath, consulToken, enableConsulTLS)

		if show {
			fmt.Println(vaultConfig)
			return nil
		}

		if len(configFile) == 0 {
			return fmt.Errorf("config-file cannot be empty")
		}

		if !strings.HasSuffix(configFile, ".hcl") {
			configFile = configFile + ".hcl"
		}

		if len(binary) == 0 && len(version) == 0 {
			latest, err := config.GetLatestVersion("vault")

			if err != nil {
				return errors.Wrapf(err, "unable to get latest version number, define a version manually with the --version flag")
			}

			version = latest
		}

		callback := func(op operator.CommandOperator) error {
			dir := "/tmp/vault-installation." + randstr.String(6)

			defer op.Execute("rm -rf " + dir)

			_, err := op.Execute("mkdir -p " + dir + "/config")
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			if len(binary) != 0 {
				info("Uploading Vault package...")
				err = op.UploadFile(binary, dir+"/vault.zip", "0644")
				if err != nil {
					return fmt.Errorf("error received during upload Vault package: %s", err)
				}
			}

			info("Uploading Vault configuration and certificates...")
			if enableTLS {
				err = op.UploadFile(certFile, dir+"/config/vault-cert.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload vault cert file: %s", err)
				}

				err = op.UploadFile(keyFile, dir+"/config/vault-key.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload vault key file: %s", err)
				}
			}

			if enableConsulTLS {
				err = op.UploadFile(consulCaFile, dir+"/config/consul-ca.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul ca file: %s", err)
				}
				err = op.UploadFile(consulCertFile, dir+"/config/consul-cert.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul cert file: %s", err)
				}
				err = op.UploadFile(consulKeyFile, dir+"/config/consul-key.pem", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload consul key file: %s", err)
				}
			}

			err = op.Upload(strings.NewReader(vaultConfig), dir+"/config/"+configFile, "0640")
			if err != nil {
				return fmt.Errorf("error received during upload vault configuration: %s", err)
			}

			for _, s := range additionalConfigFiles {
				_, filename := filepath.Split(expandPath(s))
				err = op.UploadFile(expandPath(s), dir+"/config/"+filename, "0640")
				if err != nil {
					return fmt.Errorf("error received during upload nomad config file: %s", err)
				}
			}

			installScript, err := pkger.Open("/scripts/install_vault.sh")

			if err != nil {
				return err
			}

			defer installScript.Close()

			err = op.Upload(installScript, dir+"/install.sh", "0755")
			if err != nil {
				return fmt.Errorf("error received during upload install script: %s", err)
			}

			info("Installing Vault...")
			_, err = op.Execute(fmt.Sprintf("cat %s/install.sh | TMP_DIR='%s' VAULT_VERSION='%s' SKIP_ENABLE='%t' SKIP_START='%t' sh -\n", dir, dir, version, skipEnable, skipStart))
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
