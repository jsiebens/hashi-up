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
    fatal 'Can not find systemd to use as a process supervisor for Vault'
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

  VAULT_DATA_DIR=/opt/vault
  VAULT_CONFIG_DIR=/etc/vault.d
  VAULT_SERVICE_FILE=/etc/systemd/system/vault.service
  BIN_DIR=/usr/local/bin
}

stop_and_disable_service() {
  info "Stopping and disabling Vault systemd service"
  $SUDO systemctl stop vault
  $SUDO systemctl disable vault
  $SUDO systemctl daemon-reload
}

clean_up() {
  info "Removing Vault installation"
  $SUDO rm -rf $VAULT_CONFIG_DIR
  $SUDO rm -rf $VAULT_DATA_DIR
  $SUDO rm -rf $VAULT_SERVICE_FILE
  $SUDO rm -rf $BIN_DIR/vault
}

verify_system
setup_env
stop_and_disable_service
clean_up