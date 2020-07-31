package cmd

import (
	"fmt"
	"github.com/hashicorp/go-checkpoint"
	"github.com/jsiebens/hashi-up/pkg/config"
	operator "github.com/jsiebens/hashi-up/pkg/operator"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/thanhpk/randstr"
	"golang.org/x/crypto/ssh"
	"net"
)

func InstallConsulCommand() *cobra.Command {

	var ip net.IP
	var user string
	var sshKey string
	var sshPort int

	var version string
	var datacenter string
	var bind string
	var advertise string
	var client string
	var server bool
	var boostrapExpect int64
	var retryJoin []string
	var encrypt string

	var command = &cobra.Command{
		Use:          "install",
		SilenceUsage: true,
	}

	command.Flags().IPVar(&ip, "ip", net.ParseIP("127.0.0.1"), "Public IP of node")
	command.Flags().StringVar(&user, "user", "root", "Username for SSH login")
	command.Flags().StringVar(&sshKey, "ssh-key", "~/.ssh/id_rsa", "The ssh key to use for remote login")
	command.Flags().IntVar(&sshPort, "ssh-port", 22, "The port on which to connect for ssh")

	command.Flags().StringVar(&version, "version", "", "Version of Consul to install, default to latest available")
	command.Flags().BoolVar(&server, "server", false, "Consul: switches agent to server mode. (see Consul documentation for more info)")
	command.Flags().StringVar(&datacenter, "dc", "dc1", "Consul: specifies the data center of the local agent. (see Consul documentation for more info)")
	command.Flags().StringVar(&bind, "bind", "", "Consul: sets the bind address for cluster communication. (see Consul documentation for more info)")
	command.Flags().StringVar(&advertise, "advertise", "", "Consul: sets the advertise address to use. (see Consul documentation for more info)")
	command.Flags().StringVar(&client, "client", "", "Consul: sets the address to bind for client access. (see Consul documentation for more info)")
	command.Flags().Int64Var(&boostrapExpect, "bootstrap-expect", 1, "Consul: sets server to expect bootstrap mode. (see Consul documentation for more info)")
	command.Flags().StringArrayVar(&retryJoin, "retry-join", []string{}, "Consul: address of an agent to join at start time with retries enabled. Can be specified multiple times. (see Consul documentation for more info)")
	command.Flags().StringVar(&encrypt, "encrypt", "", "Consul: provides the gossip encryption key. (see Consul documentation for more info)")

	command.RunE = func(command *cobra.Command, args []string) error {

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

		consulConfig := config.NewConsulConfiguration(datacenter, bind, advertise, client, server, boostrapExpect, retryJoin, encrypt)

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

		dir := "/tmp/consul-installation." + randstr.String(6)

		defer operator.Close()
		defer operator.Execute("rm -rf " + dir)

		_, err = operator.Execute("mkdir " + dir)
		if err != nil {
			return fmt.Errorf("error received during installation: %s", err)
		}

		err = operator.Upload(consulConfig, dir+"/consul.hcl", "0640")
		if err != nil {
			return fmt.Errorf("error received during upload consul configuration: %s", err)
		}

		err = operator.Upload(InstallConsulScript, dir+"/install.sh", "0755")
		if err != nil {
			return fmt.Errorf("error received during upload install script: %s", err)
		}

		var serviceType = "notify"
		if len(retryJoin) == 0 {
			serviceType = "exec"
		}

		_, err = operator.Execute(fmt.Sprintf("cat %s/install.sh | TMP_DIR='%s' SERVICE_TYPE='%s' CONSUL_VERSION='%s' sh -\n", dir, dir, serviceType, version))
		if err != nil {
			return fmt.Errorf("error received during installation: %s", err)
		}

		return nil
	}

	return command
}

const InstallConsulScript = `
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
    fatal 'Can not find systemd to use as a process supervisor for consul'
  fi
}

setup_env() {
  SUDO=sudo
  if [ "$(id -u)" -eq 0 ]; then
    SUDO=
  fi

  CONSUL_DATA_DIR=/opt/consul
  CONSUL_CONFIG_DIR=/etc/consul.d
  CONSUL_CONFIG_FILE=/etc/consul.d/consul.hcl
  CONSUL_SERVICE_FILE=/etc/systemd/system/consul.service  
  
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
    SUFFIX=armhfv6
    ;;
  *)
    fatal "Unsupported architecture $ARCH"
    ;;
  esac
}

# --- get hashes of the current k3s bin and service files
get_installed_hashes() {
    $SUDO sha256sum ${BIN_DIR}/consul ${CONSUL_CONFIG_FILE} ${CONSUL_SERVICE_FILE} 2>&1 || true
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
  if [ -x "${BIN_DIR}/consul" ]; then
    info "Consul binary already installed in ${BIN_DIR}, skipping downloading and installing binary"
  else
    info "Downloading and unpacking consul_${CONSUL_VERSION}_linux_${SUFFIX}.zip"
	curl -o "$TMP_DIR/consul.zip" -sfL "https://releases.hashicorp.com/consul/${CONSUL_VERSION}/consul_${CONSUL_VERSION}_linux_${SUFFIX}.zip"
    $SUDO unzip -qq "$TMP_DIR/consul.zip" -d $BIN_DIR
  fi
}

create_user_and_config() {
  if $(id consul >/dev/null 2>&1); then
    info "User consul already exists. Will not create again."
  else
    info "Creating user named consul"
    $SUDO useradd --system --home ${CONSUL_CONFIG_DIR} --shell /bin/false consul
  fi

  $SUDO mkdir --parents ${CONSUL_DATA_DIR}
  $SUDO mkdir --parents ${CONSUL_CONFIG_DIR}

  $SUDO cp "${TMP_DIR}/consul.hcl" ${CONSUL_CONFIG_FILE}

  $SUDO chown --recursive consul:consul /opt/consul
  $SUDO chown --recursive consul:consul /etc/consul.d
}

# --- write systemd service file ---
create_systemd_service_file() {
  info "Creating service file ${CONSUL_SERVICE_FILE}"
  $SUDO tee ${CONSUL_SERVICE_FILE} >/dev/null <<EOF
[Unit]
Description="HashiCorp Consul - A service mesh solution"
Documentation=https://www.consul.io/
Requires=network-online.target
After=network-online.target
ConditionFileNotEmpty=/etc/consul.d/consul.hcl

[Service]
Type=${SERVICE_TYPE}
User=consul
Group=consul
ExecStart=${BIN_DIR}/consul agent -config-dir=${CONSUL_CONFIG_DIR}
ExecReload=${BIN_DIR}/consul reload
ExecStop=${BIN_DIR}/consul leave
KillMode=process
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
}

# --- startup systemd or openrc service ---
systemd_enable_and_start() {
 	info "Enabling consul unit"
  	$SUDO systemctl enable ${CONSUL_SERVICE_FILE} >/dev/null
  	$SUDO systemctl daemon-reload >/dev/null
    
	POST_INSTALL_HASHES=$(get_installed_hashes)
    if [ "${PRE_INSTALL_HASHES}" = "${POST_INSTALL_HASHES}" ]; then
        info 'No change detected so skipping service start'
        return
    fi

  	info "Starting consul"
  	$SUDO systemctl restart consul

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
