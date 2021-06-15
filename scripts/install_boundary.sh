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
    fatal 'Can not find systemd to use as a process supervisor for Boundary'
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

  BOUNDARY_DATA_DIR=/opt/boundary
  BOUNDARY_CONFIG_DIR=/etc/boundary.d
  BOUNDARY_SERVICE_FILE=/etc/systemd/system/boundary.service

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

# --- get hashes of the current boundary bin and service files
get_installed_hashes() {
  $SUDO sha256sum ${BIN_DIR}/boundary ${BOUNDARY_CONFIG_DIR}/* ${BOUNDARY_SERVICE_FILE} 2>&1 || true
}

has_yum() {
  [ -n "$(command -v yum)" ]
}

has_apt_get() {
  [ -n "$(command -v apt-get)" ]
}

install_dependencies() {
  if [ ! -x "${TMP_DIR}/boundary" ]; then
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
  if [ -f "${TMP_DIR}/boundary.zip" ]; then
    info "Installing uploaded Boundary package"
    $SUDO unzip -qq -o "$TMP_DIR/boundary.zip" -d $BIN_DIR
  else
    if [ -x "${BIN_DIR}/boundary" ] && [ "$(${BIN_DIR}/boundary version | grep "Version Number" | tr -s ' ' | cut -d' ' -f4)" = "${BOUNDARY_VERSION}" ]; then
      info "Boundary binary already installed in ${BIN_DIR}, skipping downloading and installing binary"
    else
      info "Downloading boundary_${BOUNDARY_VERSION}_linux_${SUFFIX}.zip"
      curl -o "$TMP_DIR/boundary_${BOUNDARY_VERSION}_linux_${SUFFIX}.zip" -sfL "https://releases.hashicorp.com/boundary/${BOUNDARY_VERSION}/boundary_${BOUNDARY_VERSION}_linux_${SUFFIX}.zip"

      info "Downloading boundary_${BOUNDARY_VERSION}_SHA256SUMS"
      curl -o "$TMP_DIR/boundary_${BOUNDARY_VERSION}_SHA256SUMS" -sfL "https://releases.hashicorp.com/boundary/${BOUNDARY_VERSION}/boundary_${BOUNDARY_VERSION}_SHA256SUMS"
      info "Verifying downloaded boundary_${BOUNDARY_VERSION}_linux_${SUFFIX}.zip"
      sed -ni '/linux_'"${SUFFIX}"'.zip/p' "$TMP_DIR/boundary_${BOUNDARY_VERSION}_SHA256SUMS"
      sha256sum -c "$TMP_DIR/boundary_${BOUNDARY_VERSION}_SHA256SUMS"

      info "Unpacking boundary_${BOUNDARY_VERSION}_linux_${SUFFIX}.zip"
      $SUDO unzip -qq -o "$TMP_DIR/boundary_${BOUNDARY_VERSION}_linux_${SUFFIX}.zip" -d $BIN_DIR
    fi
  fi
}

create_user_and_config() {
  if $(id boundary >/dev/null 2>&1); then
    info "User 'boundary' already exists, will not create again"
  else
    info "Creating user named 'boundary'"
    $SUDO useradd --system --home ${BOUNDARY_CONFIG_DIR} --shell /bin/false boundary
  fi

  $SUDO mkdir --parents ${BOUNDARY_DATA_DIR}
  $SUDO mkdir --parents ${BOUNDARY_CONFIG_DIR}

  $SUDO cp ${TMP_DIR}/config/* ${BOUNDARY_CONFIG_DIR}
  $SUDO chown --recursive boundary:boundary /opt/boundary
  $SUDO chown --recursive boundary:boundary /etc/boundary.d
}

# --- write systemd service file ---
create_systemd_service_file() {
  info "Adding system service file ${BOUNDARY_SERVICE_FILE}"
  $SUDO tee ${BOUNDARY_SERVICE_FILE} >/dev/null <<EOF
[Unit]
Description=Boundary
Documentation=https://boundaryproject.io/docs/
Wants=network-online.target
After=network-online.target

[Service]
ExecStart=${BIN_DIR}/boundary server -config ${BOUNDARY_CONFIG_DIR}/boundary.hcl
ExecReload=/bin/kill -s HUP \$MAINPID
User=boundary
Group=boundary
LimitMEMLOCK=infinity
Capabilities=CAP_IPC_LOCK+ep
CapabilityBoundingSet=CAP_SYSLOG CAP_IPC_LOCK

[Install]
WantedBy=multi-user.target
EOF
}

# --- startup systemd service ---
systemd_enable_and_start() {
  [ "${SKIP_ENABLE}" = true ] && return

  info "Enabling systemd service"
  $SUDO systemctl enable ${BOUNDARY_SERVICE_FILE} >/dev/null
  $SUDO systemctl daemon-reload >/dev/null

  [ "${SKIP_START}" = true ] && return

  POST_INSTALL_HASHES=$(get_installed_hashes)
  if [ "${PRE_INSTALL_HASHES}" = "${POST_INSTALL_HASHES}" ]; then
    info "No change detected so skipping service start"
    return
  fi

  info "Starting systemd service"
  $SUDO systemctl restart boundary

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
