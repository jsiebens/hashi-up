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
    fatal "Can not find systemd to use as a process supervisor"
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

  DATA_DIR="/opt/$SERVICE"
  CONFIG_DIR="/etc/$SERVICE.d"
  SERVICE_FILE="/etc/systemd/system/$SERVICE.service"
  BIN_DIR=/usr/local/bin
}

stop_and_disable_service() {
  info "Stopping and disabling systemd service"
  $SUDO systemctl stop $SERVICE
  $SUDO systemctl disable $SERVICE
  $SUDO systemctl daemon-reload
}

clean_up() {
  info "Removing installation"
  $SUDO rm -rf $CONFIG_DIR
  $SUDO rm -rf $DATA_DIR
  $SUDO rm -rf $SERVICE_FILE
  $SUDO rm -rf $BIN_DIR/$SERVICE
}

verify_system
setup_env
stop_and_disable_service
clean_up