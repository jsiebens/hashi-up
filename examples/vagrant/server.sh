#! /bin/bash

apt-get update
apt-get install curl unzip -y

curl -sL get.hashi-up.dev | sh

hashi-up consul install \
    --local \
    --server \
    --connect \
    --advertise-addr "{{ GetInterfaceIP \"eth1\" }}" \
    --http-addr 0.0.0.0

hashi-up nomad install \
    --local \
    --advertise "{{ GetInterfaceIP \"eth1\" }}" \
    --server

hashi-up vault install \
    --local \
    --storage consul