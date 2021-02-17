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
  else
    if [ ! -z "$SUDO_PASS" ]; then
      echo $SUDO_PASS | sudo -S true
      echo ""
    fi
  fi

  NOMAD_DATA_DIR=/opt/nomad
  NOMAD_CONFIG_DIR=/etc/nomad.d
  NOMAD_SERVICE_FILE=/etc/systemd/system/nomad.service
  BIN_DIR=/usr/local/bin
}

stop_and_disable_service() {
  info "Stopping and disabling Nomad systemd service"
  $SUDO systemctl stop nomad
  $SUDO systemctl disable nomad
  $SUDO systemctl daemon-reload
}

clean_up() {
  info "Removing Nomad installation"
  $SUDO rm -rf $NOMAD_CONFIG_DIR
  $SUDO rm -rf $NOMAD_DATA_DIR
  $SUDO rm -rf $NOMAD_SERVICE_FILE
  $SUDO rm -rf $BIN_DIR/nomad
}

verify_system
setup_env
stop_and_disable_service
clean_up