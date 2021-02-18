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
    fatal 'Can not find systemd to use as a process supervisor for Nomad'
  fi
}

setup_env() {
  SUDO=sudo
  if [ "$(id -u)" -eq 0 ]; then
    SUDO=
  fi

  NOMAD_DATA_DIR=/opt/nomad
  NOMAD_CONFIG_DIR=/etc/nomad.d
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

# --- get hashes of the current nomad bin and service files
get_installed_hashes() {
  $SUDO sha256sum ${BIN_DIR}/nomad ${NOMAD_CONFIG_DIR}/* ${NOMAD_SERVICE_FILE} 2>&1 || true
}

has_yum() {
  [ -n "$(command -v yum)" ]
}

has_apt_get() {
  [ -n "$(command -v apt-get)" ]
}

install_dependencies() {
  if [ ! -x "${TMP_DIR}/nomad" ]; then
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
  if [ -f "${TMP_DIR}/nomad.zip" ]; then
    info "Installing uploaded Nomad package"
    $SUDO unzip -qq -o "$TMP_DIR/nomad.zip" -d $BIN_DIR
  else
    if [ -x "${BIN_DIR}/nomad" ] && [ "$(${BIN_DIR}/nomad version | grep Nomad | cut -d' ' -f2)" = "v${NOMAD_VERSION}" ]; then
      info "Nomad binary already installed in ${BIN_DIR}, skipping downloading and installing binary"
    else
      info "Downloading nomad_${NOMAD_VERSION}_linux_${SUFFIX}.zip"
      curl -o "$TMP_DIR/nomad_${NOMAD_VERSION}_linux_${SUFFIX}.zip" -sfL "https://releases.hashicorp.com/nomad/${NOMAD_VERSION}/nomad_${NOMAD_VERSION}_linux_${SUFFIX}.zip"

      info "Downloading nomad_${NOMAD_VERSION}_SHA256SUMS"
      curl -o "$TMP_DIR/nomad_${NOMAD_VERSION}_SHA256SUMS" -sfL "https://releases.hashicorp.com/nomad/${NOMAD_VERSION}/nomad_${NOMAD_VERSION}_SHA256SUMS"
      info "Verifying downloaded nomad_${NOMAD_VERSION}_linux_${SUFFIX}.zip"
      sed -ni '/linux_'"${SUFFIX}"'.zip/p' "$TMP_DIR/nomad_${NOMAD_VERSION}_SHA256SUMS"
      sha256sum -c "$TMP_DIR/nomad_${NOMAD_VERSION}_SHA256SUMS"

      info "Unpacking nomad_${NOMAD_VERSION}_linux_${SUFFIX}.zip"
      $SUDO unzip -qq -o "$TMP_DIR/nomad_${NOMAD_VERSION}_linux_${SUFFIX}.zip" -d $BIN_DIR
    fi
  fi
}

create_user_and_config() {
  $SUDO mkdir --parents ${NOMAD_DATA_DIR}
  $SUDO mkdir --parents ${NOMAD_CONFIG_DIR}/config

  $SUDO cp ${TMP_DIR}/config/* ${NOMAD_CONFIG_DIR}
}

# --- write systemd service file ---
create_systemd_service_file() {
  info "Adding systemd service file ${NOMAD_SERVICE_FILE}"
  $SUDO tee ${NOMAD_SERVICE_FILE} >/dev/null <<EOF
[Unit]
Description=Nomad
Documentation=https://nomadproject.io/docs/
Wants=network-online.target
After=network-online.target

[Service]
ExecReload=/bin/kill -HUP $MAINPID
ExecStart=${BIN_DIR}/nomad agent -config ${NOMAD_CONFIG_DIR}/nomad.hcl -config=${NOMAD_CONFIG_DIR}/config
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

# --- startup systemd service ---
systemd_enable_and_start() {
  [ "${SKIP_ENABLE}" = true ] && return

  info "Enabling systemd service"
  $SUDO systemctl enable ${NOMAD_SERVICE_FILE} >/dev/null
  $SUDO systemctl daemon-reload >/dev/null

  [ "${SKIP_START}" = true ] && return

  POST_INSTALL_HASHES=$(get_installed_hashes)
  if [ "${PRE_INSTALL_HASHES}" = "${POST_INSTALL_HASHES}" ]; then
    info "No change detected so skipping service start"
    return
  fi

  info "Starting systemd service"
  $SUDO systemctl restart nomad

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