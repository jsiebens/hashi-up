#!/usr/bin/env bash

multipass delete consul-server
multipass delete consul-client-01
multipass delete consul-client-02
multipass delete consul-client-03
multipass purge
