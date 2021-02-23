#! /bin/bash

apt-get update
apt-get install curl unzip docker.io -y

mkdir -p /opt/cni/bin
curl -sSL https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz | tar -xvz -C /opt/cni/bin

curl -sL get.hashi-up.dev | sh

hashi-up consul install \
    --local \
    --connect \
    --advertise-addr "{{ GetInterfaceIP \"eth1\" }}" \
    --retry-join "10.100.0.10"

hashi-up nomad install \
    --local \
    --advertise "{{ GetInterfaceIP \"eth1\" }}" \
    --client