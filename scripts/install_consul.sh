#!/bin/bash
set -e

info() {
  echo '[INFO] ->' "$@"
}

fatal() {
  echo '[ERROR] ->' "$@"
  exit 1
}

verify_system() {
  if ! [ -d /run/systemd ]; then
    fatal 'Can not find systemd to use as a process supervisor for Consul'
  fi
}

setup_env() {
  SUDO=sudo
  if [ "$(id -u)" -eq 0 ]; then
    SUDO=
  else
    if [ ! -z "$SUDO_PASS" ]; then
      echo $SUDO_PASS | sudo -S true
      echo ""
    fi
  fi

  CONSUL_DATA_DIR=/opt/consul
  CONSUL_CONFIG_DIR=/etc/consul.d
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
  $SUDO sha256sum ${BIN_DIR}/consul ${CONSUL_CONFIG_DIR}/* ${CONSUL_SERVICE_FILE} 2>&1 || true
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
        $SUDO apt-get install -y curl unzip
      elif $(has_yum); then
        $SUDO yum install -y curl unzip
      else
        fatal "Could not find apt-get or yum. Cannot install dependencies on this OS"
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
      info "Downloading consul_${CONSUL_VERSION}_linux_${SUFFIX}.zip"
      curl -o "$TMP_DIR/consul_${CONSUL_VERSION}_linux_${SUFFIX}.zip" -sfL "https://releases.hashicorp.com/consul/${CONSUL_VERSION}/consul_${CONSUL_VERSION}_linux_${SUFFIX}.zip"

      info "Downloading consul_${CONSUL_VERSION}_SHA256SUMS"
      curl -o "$TMP_DIR/consul_${CONSUL_VERSION}_SHA256SUMS" -sfL "https://releases.hashicorp.com/consul/${CONSUL_VERSION}/consul_${CONSUL_VERSION}_SHA256SUMS"
      info "Verifying downloaded consul_${CONSUL_VERSION}_linux_${SUFFIX}.zip"
      sed -ni '/linux_'"${SUFFIX}"'.zip/p' "$TMP_DIR/consul_${CONSUL_VERSION}_SHA256SUMS"
      sha256sum -c "$TMP_DIR/consul_${CONSUL_VERSION}_SHA256SUMS"

      info "Unpacking consul_${CONSUL_VERSION}_linux_${SUFFIX}.zip"
      $SUDO unzip -qq -o "$TMP_DIR/consul_${CONSUL_VERSION}_linux_${SUFFIX}.zip" -d $BIN_DIR
    fi
  fi
}

create_user_and_config() {
  if $(id consul >/dev/null 2>&1); then
    info "User 'consul' already exists, will not create again"
  else
    info "Creating user named 'consul'"
    $SUDO useradd --system --home ${CONSUL_CONFIG_DIR} --shell /bin/false consul
  fi

  $SUDO mkdir --parents ${CONSUL_DATA_DIR}
  $SUDO mkdir --parents ${CONSUL_CONFIG_DIR}/config

  $SUDO cp ${TMP_DIR}/config/* ${CONSUL_CONFIG_DIR}
  $SUDO chown --recursive consul:consul /opt/consul
  $SUDO chown --recursive consul:consul /etc/consul.d
}

# --- write systemd service file ---
create_systemd_service_file() {
  info "Adding systemd service file ${CONSUL_SERVICE_FILE}"
  $SUDO tee ${CONSUL_SERVICE_FILE} >/dev/null <<EOF
[Unit]
Description="HashiCorp Consul - A service mesh solution"
Documentation=https://www.consul.io/
Requires=network-online.target
After=network-online.target

[Service]
Type=${SERVICE_TYPE}
User=consul
Group=consul
ExecStart=${BIN_DIR}/consul agent -config-file=${CONSUL_CONFIG_DIR}/consul.hcl -config-dir=${CONSUL_CONFIG_DIR}/config
ExecReload=/bin/kill --signal HUP $MAINPID
KillMode=process
KillSignal=SIGTERM
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
}

# --- startup systemd service ---
systemd_enable_and_start() {
  [ "${SKIP_ENABLE}" = true ] && return

  info "Enabling systemd service"
  $SUDO systemctl enable ${CONSUL_SERVICE_FILE} >/dev/null
  $SUDO systemctl daemon-reload >/dev/null

  [ "${SKIP_START}" = true ] && return

  POST_INSTALL_HASHES=$(get_installed_hashes)
  if [ "${PRE_INSTALL_HASHES}" = "${POST_INSTALL_HASHES}" ]; then
    info "No change detected so skipping service start"
    return
  fi

  info "Starting systemd service"
  $SUDO systemctl restart consul

  return 0
}

cd $TMP_DIR

setup_env
setup_verify_arch
verify_system
install_dependencies
create_user_and_config
download_and_install
create_systemd_service_file
systemd_enable_and_start
