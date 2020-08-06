package cmd

import (
	"fmt"
	checkpoint "github.com/hashicorp/go-checkpoint"
	"github.com/jsiebens/hashi-up/pkg/config"
	operator "github.com/jsiebens/hashi-up/pkg/operator"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/thanhpk/randstr"
	"golang.org/x/crypto/ssh"
	"net"
)

func InstallNomadCommand() *cobra.Command {

	var ip net.IP
	var user string
	var sshKey string
	var sshPort int
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

	command.Flags().IPVar(&ip, "ip", net.ParseIP("127.0.0.1"), "Public IP of node")
	command.Flags().StringVar(&user, "user", "root", "Username for SSH login")
	command.Flags().StringVar(&sshKey, "ssh-key", "~/.ssh/id_rsa", "The ssh key to use for remote login")
	command.Flags().IntVar(&sshPort, "ssh-port", 22, "The port on which to connect for ssh")
	command.Flags().BoolVar(&show, "show", false, "Just show the generated config instead of deploying Consul")

	command.Flags().StringVar(&version, "version", "", "Version of Nomad to install, default to latest available")
	command.Flags().BoolVar(&server, "server", false, "Nomad: enables the server mode of the agent. (see Nomad documentation for more info)")
	command.Flags().BoolVar(&client, "client", false, "Nomad: enables the client mode of the agent. (see Nomad documentation for more info)")
	command.Flags().StringVar(&datacenter, "dc", "dc1", "Nomad: specifies the data center of the local agent. (see Nomad documentation for more info)")
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

		fmt.Println("Public IP: " + ip.String())

		sshKeyPath := expandPath(sshKey)
		fmt.Printf("ssh -i %s -p %d %s@%s\n", sshKeyPath, sshPort, user, ip.String())

		authMethod, closeSSHAgent, err := loadPublickey(sshKeyPath)
		if err != nil {
			return errors.Wrapf(err, "unable to load the ssh key with path %q", sshKeyPath)
		}

		defer closeSSHAgent()

		config := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				authMethod,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		address := fmt.Sprintf("%s:%d", ip.String(), sshPort)
		operator, err := operator.NewSSHOperator(address, config)

		if err != nil {
			return errors.Wrapf(err, "unable to connect to %s over ssh", address)
		}

		dir := "/tmp/nomad-installation." + randstr.String(6)

		defer operator.Close()
		defer operator.Execute("rm -rf " + dir)

		_, err = operator.Execute("mkdir " + dir)
		if err != nil {
			return fmt.Errorf("error received during installation: %s", err)
		}

		if enableTLS {
			err = operator.UploadFile(caFile, dir+"/nomad-agent-ca.pem", "0640")
			if err != nil {
				return fmt.Errorf("error received during upload nomad ca file: %s", err)
			}

			err = operator.UploadFile(certFile, dir+"/nomad-agent-cert.pem", "0640")
			if err != nil {
				return fmt.Errorf("error received during upload nomad cert file: %s", err)
			}

			err = operator.UploadFile(keyFile, dir+"/nomad-agent-key.pem", "0640")
			if err != nil {
				return fmt.Errorf("error received during upload nomad key file: %s", err)
			}
		}

		err = operator.Upload(nomadConfig, dir+"/nomad.hcl", "0640")
		if err != nil {
			return fmt.Errorf("error received during upload nomad configuration: %s", err)
		}

		err = operator.Upload(InstallNomadScript, dir+"/install.sh", "0755")
		if err != nil {
			return fmt.Errorf("error received during upload install script: %s", err)
		}

		_, err = operator.Execute(fmt.Sprintf("cat %s/install.sh | TMP_DIR='%s' NOMAD_VERSION='%s' sh -\n", dir, dir, version))
		if err != nil {
			return fmt.Errorf("error received during installation: %s", err)
		}

		return nil
	}

	return command
}

const InstallNomadScript = `
#!/bin/bash
set -e

info() {
  echo '[INFO] ' "$@"
}

fatal() {
  echo '[ERROR] ' "$@"
  exit 1
}

verify_system() {
  if ! [ -d /run/systemd ]; then
    fatal 'Can not find systemd to use as a process supervisor for nomad'
  fi
}

setup_env() {
  SUDO=sudo
  if [ "$(id -u)" -eq 0 ]; then
    SUDO=
  fi

  NOMAD_DATA_DIR=/opt/nomad
  NOMAD_CONFIG_DIR=/etc/nomad.d
  NOMAD_CONFIG_FILE=/etc/nomad.d/nomad.hcl
  NOMAD_SERVICE_FILE=/etc/systemd/system/nomad.service  
  
  BIN_DIR=/usr/local/bin

  PRE_INSTALL_HASHES=$(get_installed_hashes)
}

# --- set arch and suffix, fatal if architecture not supported ---
setup_verify_arch() {
  if [ -z "$ARCH" ]; then
    ARCH=$(uname -m)
  fi
  case $ARCH in
  amd64)
    SUFFIX=amd64
    ;;
  x86_64)
    SUFFIX=amd64
    ;;
  arm64)
    SUFFIX=arm64
    ;;
  aarch64)
    SUFFIX=arm64
    ;;
  arm*)
    SUFFIX=arm
    ;;
  *)
    fatal "Unsupported architecture $ARCH"
    ;;
  esac
}

# --- get hashes of the current k3s bin and service files
get_installed_hashes() {
    $SUDO sha256sum ${BIN_DIR}/nomad /etc/nomad.d/nomad.hcl /etc/nomad.d/nomad-agent-ca.pem /etc/nomad.d/nomad-agent-cert.pem /etc/nomad.d/nomad-agent-key.pem ${FILE_CONSUL_SERVICE} 2>&1 || true
}

has_yum() {
  [ -n "$(command -v yum)" ]
}

has_apt_get() {
  [ -n "$(command -v apt-get)" ]
}

install_dependencies() {
  if ! [ -x "$(command -v unzip)" ] || ! [ -x "$(command -v curl)" ]; then
	  if $(has_apt_get); then
		$SUDO apt-get update -y
		$SUDO apt-get install -y curl unzip
	  elif $(has_yum); then
		$SUDO yum update -y
		$SUDO yum install -y curl unzip
	  else
		fatal "Could not find apt-get or yum. Cannot install dependencies on this OS."
		exit 1
	  fi
  fi
}

download_and_install() {
  if [ -x "${BIN_DIR}/nomad" ]; then
    info "Nomad binary already installed in ${BIN_DIR}, skipping downloading and installing binary"
  else
    info "Downloading and unpacking nomad_${NOMAD_VERSION}_linux_${SUFFIX}.zip"
	curl -o "$TMP_DIR/nomad.zip" -sfL "https://releases.hashicorp.com/nomad/${NOMAD_VERSION}/nomad_${NOMAD_VERSION}_linux_${SUFFIX}.zip"
    $SUDO unzip -qq "$TMP_DIR/nomad.zip" -d $BIN_DIR
  fi
}

create_user_and_config() {
  $SUDO mkdir --parents ${NOMAD_DATA_DIR}
  $SUDO mkdir --parents ${NOMAD_CONFIG_DIR}

  $SUDO cp "${TMP_DIR}/nomad.hcl" ${NOMAD_CONFIG_FILE}
  if [ -f "${TMP_DIR}/nomad-agent-ca.pem" ]; then
	$SUDO cp "${TMP_DIR}/nomad-agent-ca.pem" /etc/nomad.d/nomad-agent-ca.pem
  fi
  if [ -f "${TMP_DIR}/nomad-agent-cert.pem" ]; then
	$SUDO cp "${TMP_DIR}/nomad-agent-cert.pem" /etc/nomad.d/nomad-agent-cert.pem
  fi
  if [ -f "${TMP_DIR}/nomad-agent-key.pem" ]; then
	$SUDO cp "${TMP_DIR}/nomad-agent-key.pem" /etc/nomad.d/nomad-agent-key.pem
  fi
}

# --- write systemd service file ---
create_systemd_service_file() {
  info "Creating service file ${NOMAD_SERVICE_FILE}"
  $SUDO tee ${NOMAD_SERVICE_FILE} >/dev/null <<EOF
[Unit]
Description=Nomad
Documentation=https://nomadproject.io/docs/
Wants=network-online.target
After=network-online.target

[Service]
ExecReload=/bin/kill -HUP $MAINPID
ExecStart=${BIN_DIR}/nomad agent -config ${NOMAD_CONFIG_DIR}
KillMode=process
KillSignal=SIGINT
LimitNOFILE=infinity
LimitNPROC=infinity
Restart=on-failure
RestartSec=2
StartLimitBurst=3
StartLimitIntervalSec=10
TasksMax=infinity

[Install]
WantedBy=multi-user.target
EOF
}

# --- startup systemd or openrc service ---
systemd_enable_and_start() {
 	info "Enabling nomad unit"
  	$SUDO systemctl enable ${NOMAD_SERVICE_FILE} >/dev/null
  	$SUDO systemctl daemon-reload >/dev/null
    
	POST_INSTALL_HASHES=$(get_installed_hashes)
    if [ "${PRE_INSTALL_HASHES}" = "${POST_INSTALL_HASHES}" ]; then
        info 'No change detected so skipping service start'
        return
    fi

  	info "Starting nomad"
  	$SUDO systemctl restart nomad

    return 0
}

setup_env
setup_verify_arch
verify_system
install_dependencies
create_user_and_config
download_and_install
create_systemd_service_file
systemd_enable_and_start

`
