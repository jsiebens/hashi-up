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
}

stop_and_disable_service() {
  info "Stopping and disabling Boundary systemd service"
  $SUDO systemctl stop boundary
  $SUDO systemctl disable boundary
  $SUDO systemctl daemon-reload
}

clean_up() {
  info "Removing Boundary installation"
  $SUDO rm -rf $BOUNDARY_CONFIG_DIR
  $SUDO rm -rf $BOUNDARY_DATA_DIR
  $SUDO rm -rf $BOUNDARY_SERVICE_FILE
  $SUDO rm -rf $BIN_DIR/boundary
}

verify_system
setup_env
stop_and_disable_service
clean_up