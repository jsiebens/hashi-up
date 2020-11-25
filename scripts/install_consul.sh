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

  CONSUL_DATA_DIR=/var/lib/consul
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

# --- get hashes of the current consul bin and service files
get_installed_hashes() {
  $SUDO sha256sum ${BIN_DIR}/consul /etc/consul.d/consul.hcl /etc/consul.d/consul-agent-ca.pem /etc/consul.d/consul-agent-cert.pem /etc/consul.d/consul-agent-key.pem ${FILE_CONSUL_SERVICE} 2>&1 || true
}

has_yum() {
  [ -n "$(command -v yum)" ]
}

has_apt_get() {
  [ -n "$(command -v apt-get)" ]
}

install_dependencies() {
  if [ ! -x "${TMP_DIR}/consul" ]; then
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
  fi
}

download_and_install() {
  if [ -f "${TMP_DIR}/consul.zip" ]; then
    info "Installing uploaded Consul package"
    $SUDO unzip -qq -o "$TMP_DIR/consul.zip" -d $BIN_DIR
  else
    if [ -x "${BIN_DIR}/consul" ] && [ "$(${BIN_DIR}/consul version | grep Consul | cut -d' ' -f2)" = "v${CONSUL_VERSION}" ]; then
      info "Consul binary already installed in ${BIN_DIR}, skipping downloading and installing binary"
    else
      info "Downloading and unpacking consul_${CONSUL_VERSION}_linux_${SUFFIX}.zip"
      curl -o "$TMP_DIR/consul.zip" -sfL "https://releases.hashicorp.com/consul/${CONSUL_VERSION}/consul_${CONSUL_VERSION}_linux_${SUFFIX}.zip"
      $SUDO unzip -qq -o "$TMP_DIR/consul.zip" -d $BIN_DIR
    fi
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
  if [ -f "${TMP_DIR}/consul-agent-ca.pem" ]; then
    $SUDO cp "${TMP_DIR}/consul-agent-ca.pem" /etc/consul.d/consul-agent-ca.pem
  fi
  if [ -f "${TMP_DIR}/consul-agent-cert.pem" ]; then
    $SUDO cp "${TMP_DIR}/consul-agent-cert.pem" /etc/consul.d/consul-agent-cert.pem
  fi
  if [ -f "${TMP_DIR}/consul-agent-key.pem" ]; then
    $SUDO cp "${TMP_DIR}/consul-agent-key.pem" /etc/consul.d/consul-agent-key.pem
  fi

  $SUDO chown --recursive consul:consul ${CONSUL_DATA_DIR}
  $SUDO chown --recursive consul:consul ${CONSUL_CONFIG_DIR}
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

# --- startup systemd service ---
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
