#!/usr/bin/env bash

PUBLIC_KEY=$(cat ${1:-~/.ssh/id_rsa.pub})

create_cloud_init() {
  cat << EOF
#cloud-config
ssh_authorized_keys:
  - $PUBLIC_KEY
EOF
}

create_cloud_init | multipass launch --cpus 1 --mem 1G --disk 5G --name vault --cloud-init -

NODE_IP=$(multipass info vault | grep 'IPv4' | awk '{print $2}')

hashi-up vault install \
  --ssh-target-addr "$NODE_IP" \
  --ssh-target-user ubuntu \
  --storage file
