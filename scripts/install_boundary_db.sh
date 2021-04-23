#!/bin/bash
set -e

info() {
  echo '[INFO] ->' "$@"
}

fatal() {
  echo '[ERROR] ->' "$@"
  exit 1
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

  BIN_DIR=/usr/local/bin
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

init_database() {
  $SUDO ${BIN_DIR}/boundary database init -config ${TMP_DIR}/config/boundary.hcl
}

cd $TMP_DIR

setup_env
setup_verify_arch
install_dependencies
download_and_install
init_database
