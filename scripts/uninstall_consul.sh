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
}

stop_and_disable_service() {
  info "Stopping and disabling Consul systemd service"
  $SUDO systemctl stop consul
  $SUDO systemctl disable consul
  $SUDO systemctl daemon-reload
}

clean_up() {
  info "Removing Consul installation"
  $SUDO rm -rf $CONSUL_CONFIG_DIR
  $SUDO rm -rf $CONSUL_DATA_DIR
  $SUDO rm -rf $CONSUL_SERVICE_FILE
  $SUDO rm -rf $BIN_DIR/consul
}

verify_system
setup_env
stop_and_disable_service
clean_up