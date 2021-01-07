#!/usr/bin/env bash

multipass delete nomad-server
multipass delete nomad-client-01
multipass delete nomad-client-02
multipass delete nomad-client-03
multipass purge
