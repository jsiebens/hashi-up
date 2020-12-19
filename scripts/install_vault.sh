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
    fatal 'Can not find systemd to use as a process supervisor for vault'
  fi
}

setup_env() {
  SUDO=sudo
  if [ "$(id -u)" -eq 0 ]; then
    SUDO=
  fi

  VAULT_DATA_DIR=/opt/vault
  VAULT_CONFIG_DIR=/etc/vault.d
  VAULT_SERVICE_FILE=/etc/systemd/system/vault.service

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

# --- get hashes of the current vault bin and service files
get_installed_hashes() {
  $SUDO sha256sum ${BIN_DIR}/vault ${VAULT_CONFIG_DIR}/* ${VAULT_SERVICE_FILE} 2>&1 || true
}

has_yum() {
  [ -n "$(command -v yum)" ]
}

has_apt_get() {
  [ -n "$(command -v apt-get)" ]
}

install_dependencies() {
  if [ ! -x "${TMP_DIR}/vault" ]; then
    if ! [ -x "$(command -v unzip)" ] || ! [ -x "$(command -v curl)" ]; then
      if $(has_apt_get); then
        $SUDO apt-get install -y curl unzip
      elif $(has_yum); then
        $SUDO yum install -y curl unzip
      else
        fatal "Could not find apt-get or yum. Cannot install dependencies on this OS."
        exit 1
      fi
    fi
  fi
}

download_and_install() {
  if [ -f "${TMP_DIR}/vault.zip" ]; then
    info "Installing uploaded Vault package"
    $SUDO unzip -qq -o "$TMP_DIR/vault.zip" -d $BIN_DIR
    $SUDO setcap cap_ipc_lock=+ep "${BIN_DIR}/vault"
  else
    if [ -x "${BIN_DIR}/vault" ] && [ "$(${BIN_DIR}/vault version | grep Vault | cut -d' ' -f2)" = "v${VAULT_VERSION}" ]; then
      info "Vault binary already installed in ${BIN_DIR}, skipping downloading and installing binary"
    else
      info "Downloading and unpacking vault_${VAULT_VERSION}_linux_${SUFFIX}.zip"
      curl -o "$TMP_DIR/vault.zip" -sfL "https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_linux_${SUFFIX}.zip"
      $SUDO unzip -qq -o "$TMP_DIR/vault.zip" -d $BIN_DIR
      $SUDO setcap cap_ipc_lock=+ep "${BIN_DIR}/vault"
    fi
  fi
}

create_user_and_config() {
  if $(id vault >/dev/null 2>&1); then
    info "User vault already exists. Will not create again."
  else
    info "Creating user named vault"
    $SUDO useradd --system --home ${VAULT_CONFIG_DIR} --shell /bin/false vault
  fi

  $SUDO mkdir --parents ${VAULT_DATA_DIR}
  $SUDO mkdir --parents ${VAULT_CONFIG_DIR}

  $SUDO cp ${TMP_DIR}/config/* ${VAULT_CONFIG_DIR}

  $SUDO chown --recursive vault:vault /opt/vault
  $SUDO chown --recursive vault:vault /etc/vault.d
}

# --- write systemd service file ---
create_systemd_service_file() {
  info "Creating service file ${VAULT_SERVICE_FILE}"
  $SUDO tee ${VAULT_SERVICE_FILE} >/dev/null <<EOF
[Unit]
Description="HashiCorp Vault - A tool for managing secrets"
Documentation=https://www.vaultproject.io/docs/
Requires=network-online.target
After=network-online.target
StartLimitIntervalSec=60
StartLimitBurst=3
[Service]
User=vault
Group=vault
ProtectSystem=full
ProtectHome=read-only
PrivateTmp=yes
PrivateDevices=yes
SecureBits=keep-caps
AmbientCapabilities=CAP_IPC_LOCK
Capabilities=CAP_IPC_LOCK+ep
CapabilityBoundingSet=CAP_SYSLOG CAP_IPC_LOCK
NoNewPrivileges=yes
ExecStart=/usr/local/bin/vault server -config=/etc/vault.d/
ExecReload=/bin/kill --signal HUP $MAINPID
KillMode=process
KillSignal=SIGINT
Restart=on-failure
RestartSec=5
TimeoutStopSec=30
StartLimitInterval=60
StartLimitIntervalSec=60
StartLimitBurst=3
LimitNOFILE=65536
LimitMEMLOCK=infinity
[Install]
WantedBy=multi-user.target
EOF
}

# --- startup systemd service ---
systemd_enable_and_start() {
  [ "${SKIP_ENABLE}" = true ] && return

  info "Enabling vault unit"
  $SUDO systemctl enable ${VAULT_SERVICE_FILE} >/dev/null
  $SUDO systemctl daemon-reload >/dev/null

  [ "${SKIP_START}" = true ] && return

  POST_INSTALL_HASHES=$(get_installed_hashes)
  if [ "${PRE_INSTALL_HASHES}" = "${POST_INSTALL_HASHES}" ]; then
    info 'No change detected so skipping service start'
    return
  fi

  info "Starting vault"
  $SUDO systemctl restart vault

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
