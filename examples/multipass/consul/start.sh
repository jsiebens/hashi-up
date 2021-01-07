#!/usr/bin/env bash

PUBLIC_KEY=$(cat ${1:-~/.ssh/id_rsa.pub})

create_cloud_init() {
  cat <<EOF
#cloud-config
ssh_authorized_keys:
  - $PUBLIC_KEY
EOF
}

create_cloud_init | multipass launch --cpus 1 --mem 1G --disk 5G --name consul-server --cloud-init -
create_cloud_init | multipass launch --cpus 1 --mem 1G --disk 5G --name consul-client-01 --cloud-init -
create_cloud_init | multipass launch --cpus 1 --mem 1G --disk 5G --name consul-client-02 --cloud-init -
create_cloud_init | multipass launch --cpus 1 --mem 1G --disk 5G --name consul-client-03 --cloud-init -

SERVER_IP=$(multipass info consul-server | grep 'IPv4' | awk '{print $2}')
CLIENT_1_IP=$(multipass info consul-client-01 | grep 'IPv4' | awk '{print $2}')
CLIENT_2_IP=$(multipass info consul-client-02 | grep 'IPv4' | awk '{print $2}')
CLIENT_3_IP=$(multipass info consul-client-03 | grep 'IPv4' | awk '{print $2}')

hashi-up consul install --ssh-target-addr "$SERVER_IP" --ssh-target-user ubuntu --server --client 0.0.0.0 &
hashi-up consul install --ssh-target-addr "$CLIENT_1_IP" --ssh-target-user ubuntu --retry-join "$SERVER_IP" &
hashi-up consul install --ssh-target-addr "$CLIENT_2_IP" --ssh-target-user ubuntu --retry-join "$SERVER_IP" &
hashi-up consul install --ssh-target-addr "$CLIENT_3_IP" --ssh-target-user ubuntu --retry-join "$SERVER_IP" &

wait
