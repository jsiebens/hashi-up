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

func InstallBoundaryCommand() *cobra.Command {

	var skipEnable bool
	var skipStart bool
	var binary string
	var version string

	var configFile string
	var files []string

	var flags = config.BoundaryConfig{}

	var command = &cobra.Command{
		Use:          "install",
		Short:        "Install Boundary on a server via SSH",
		Long:         "Install Boundary on a server via SSH",
		SilenceUsage: true,
	}

	var target = Target{}
	target.prepareCommand(command)

	command.Flags().BoolVar(&skipEnable, "skip-enable", false, "If set to true will not enable or start Boundary service")
	command.Flags().BoolVar(&skipStart, "skip-start", false, "If set to true will not start Boundary service")
	command.Flags().StringVar(&binary, "package", "", "Upload and use this Boundary package instead of downloading")
	command.Flags().StringVarP(&version, "version", "v", "", "Version of Boundary to install")

	command.Flags().StringVarP(&configFile, "config-file", "c", "", "Custom Boundary configuration file to upload")
	command.Flags().StringArrayVarP(&files, "file", "f", []string{}, "Additional files, e.g. certificates, to upload")

	command.Flags().StringVar(&flags.ControllerName, "controller-name", "", "Boundary: specifies a unique name of this controller within the Boundary controller cluster.")
	command.Flags().StringVar(&flags.WorkerName, "worker-name", "", "Boundary: specifies a unique name of this worker within the Boundary worker cluster.")
	command.Flags().StringVar(&flags.DatabaseURL, "db-url", "", "Boundary: configures the URL for connecting to Postgres")
	command.Flags().StringVar(&flags.RootKey, "root-key", "", "Boundary: a KEK (Key Encrypting Key) for the scope-specific KEKs (also referred to as the scope's root key).")
	command.Flags().StringVar(&flags.WorkerAuthKey, "worker-auth-key", "", "Boundary: KMS key shared by the Controller and Worker in order to authenticate a Worker to the Controller.")
	command.Flags().StringVar(&flags.RecoveryKey, "recovery-key", "", "Boundary: KMS key is used for rescue/recovery operations that can be used by a client to authenticate almost any operation within Boundary.")
	command.Flags().StringVar(&flags.ApiAddress, "api-addr", "0.0.0.0", "Boundary: address for the API listener")
	command.Flags().StringVar(&flags.ApiKeyFile, "api-key-file", "", "Boundary: specifies the path to the private key for the certificate.")
	command.Flags().StringVar(&flags.ApiCertFile, "api-cert-file", "", "Boundary: specifies the path to the certificate for TLS.")
	command.Flags().StringVar(&flags.ClusterAddress, "cluster-addr", "127.0.0.1", "Boundary: address for the Cluster listener")
	command.Flags().StringVar(&flags.ClusterKeyFile, "cluster-key-file", "", "Boundary: specifies the path to the private key for the certificate.")
	command.Flags().StringVar(&flags.ClusterCertFile, "cluster-cert-file", "", "Boundary: specifies the path to the certificate for TLS.")
	command.Flags().StringVar(&flags.ProxyAddress, "proxy-addr", "0.0.0.0", "Boundary: address for the Proxy listener")
	command.Flags().StringVar(&flags.ProxyKeyFile, "proxy-key-file", "", "Boundary: specifies the path to the private key for the certificate.")
	command.Flags().StringVar(&flags.ProxyCertFile, "proxy-cert-file", "", "Boundary: specifies the path to the certificate for TLS.")
	command.Flags().StringVar(&flags.PublicClusterAddress, "public-cluster-addr", "", "Boundary: specifies the public host or IP address (and optionally port) at which the controller can be reached by workers.")
	command.Flags().StringVar(&flags.PublicAddress, "public-addr", "", "Boundary: specifies the public host or IP address (and optionally port) at which the worker can be reached by clients for proxying.")
	command.Flags().StringArrayVar(&flags.Controllers, "controller", []string{"127.0.0.1"}, "Boundary: a list of hosts/IP addresses and optionally ports for reaching controllers.")

	command.RunE = func(command *cobra.Command, args []string) error {
		if !target.Local && len(target.Addr) == 0 {
			return fmt.Errorf("required ssh-target-addr flag is missing")
		}

		ignoreConfigFlags := len(configFile) != 0

		var generatedConfig string

		if !ignoreConfigFlags {
			if !(flags.IsControllerEnabled() || flags.IsWorkerEnabled()) {
				return fmt.Errorf("a controller-name and/or a worker-name is required")
			}

			if flags.IsControllerEnabled() {
				if !flags.HasDatabaseURL() {
					return fmt.Errorf("a db-url is required when running a controller")
				}
				if !flags.HasAllRequiredControllerKeys() {
					return fmt.Errorf("a root-key, a worker-auth-key and a recovery-key are required when running a controller")
				}
			}

			if flags.IsWorkerEnabled() && !flags.HasAllRequiredWorkerKeys() {
				return fmt.Errorf("a worker-auth-key are required when running a worker")
			}

			if !flags.HasValidApiTLSSettings() {
				return fmt.Errorf("both api-key-file and api-cert-file are required to enable API TLS")
			}

			if !flags.HasValidClusterTLSSettings() {
				return fmt.Errorf("both cluster-key-file and cluster-cert-file are required to enable cluster TLS")
			}

			if !flags.HasValidProxyTLSSettings() {
				return fmt.Errorf("both proxy-key-file and proxy-cert-file are required to enable proxy TLS")
			}

			generatedConfig = flags.GenerateConfigFile()
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

			_, err := op.Execute("mkdir -p " + dir + "/config")
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

				files = []string{}

				if flags.ApiTLSEnabled() {
					files = []string{flags.ApiCertFile, flags.ApiKeyFile}
				}
				if flags.ClusterTLSEnabled() {
					files = []string{flags.ClusterKeyFile, flags.ClusterCertFile}
				}
				if flags.ProxyTLSEnabled() {
					files = []string{flags.ProxyKeyFile, flags.ProxyCertFile}
				}
			} else {
				info(fmt.Sprintf("Uploading %s as boundary.hcl...", configFile))
				err = op.UploadFile(expandPath(configFile), dir+"/config/boundary.hcl", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload boundary configuration: %s", err)
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

			installScript, err := scripts.Open("install_boundary.sh")

			if err != nil {
				return err
			}

			defer installScript.Close()

			err = op.Upload(installScript, dir+"/install.sh", "0755")
			if err != nil {
				return fmt.Errorf("error received during upload install script: %s", err)
			}

			info("Installing Boundary ...")
			_, err = op.Execute(fmt.Sprintf("cat %s/install.sh | TMP_DIR='%s' BOUNDARY_VERSION='%s' SKIP_ENABLE='%t' SKIP_START='%t' sh -\n", dir, dir, version, skipEnable, skipStart))
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
