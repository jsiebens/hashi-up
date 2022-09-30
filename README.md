# hashi-up

hashi-up is a lightweight utility to install HashiCorp [Consul](https://www.consul.io/), [Nomad](https://www.nomadproject.io) or [Vault](https://www.vaultproject.io/) on any remote Linux host. All you need is `ssh` access and the binary `hashi-up` to build a Consul, Nomad or Vault cluster.

The tool is written in Go and is cross-compiled for Linux, Windows, MacOS and even on Raspberry Pi.

This project is heavily inspired on the work of [Alex Ellis](https://www.alexellis.io/) who created [k3sup](https://k3sup.dev/), a tool to to get from zero to KUBECONFIG with [k3s](https://k3s.io/)

[![Go Report Card](https://goreportcard.com/badge/github.com/jsiebens/hashi-up)](https://goreportcard.com/report/github.com/jsiebens/hashi-up)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![GitHub All Releases](https://img.shields.io/github/downloads/jsiebens/hashi-up/total)
![matomo](https://matomo.nosceon.io/matomo.php?idsite=2&rec=1&action_name=gh-home)

## What's this for?

This tool uses `ssh` to install HashiCorp Consul, Nomad or Vault to a remote Linux host. You can also use it to join existing Linux hosts into a Consul, Nomad, Vault or Boundary cluster. First, Consul, Nomad or Vault is installed using a utility script, along with a minimal configuration to run the agent as server or client.

`hashi-up` was developed to automate what can be a very manual and confusing process for many developers, who are already short on time. Once you've provisioned a VM with your favourite tooling, `hashi-up` means you are only 60 seconds away from running `nomad status` on your own computer.

## Download `hashi-up`

`hashi-up` is distributed as a static Go binary. 
You can use the installer on MacOS and Linux, or visit the Releases page to download the executable for Windows.

``` shell
curl -sLS https://get.hashi-up.dev | sh
sudo install hashi-up /usr/local/bin/

hashi-up version
```

## Usage

The `hashi-up` tool is a client application which you can run on your own computer. It uses SSH to connect to remote servers when installing HashiCorp Consul or Nomad. Binaries are provided for MacOS, Windows, and Linux (including ARM).

### SSH credentials

By default, `hashi-up` talks to an SSH agent on your host via the SSH agent protocol. This saves you from typing a passphrase for an encrypted private key every time you connect to a server.
The `ssh-agent` that comes with OpenSSH is commonly used, but other agents, like gpg-agent or yubikey-agent are supported by setting the `SSH_AUTH_SOCK` environment variable to the Unix domain socket of the agent.

The `--ssh-target-key` flag can be used when no agent is available or when a specific private key is preferred.

The `--ssh-target-user` and `--ssh-target-password` flags allow you to authenticate using a username and a password.

### Guides

- [Installing Consul](docs/consul.md)
- [Installing Nomad](docs/nomad.md)
- [Installing Vault](docs/vault.md)
- [Installing Boundary](docs/boundary.md)

## Resources

[Deploying a highly-available Nomad cluster with hashi-up!](https://johansiebens.dev/posts/2020/07/deploying-a-highly-available-nomad-cluster-with-hashi-up/)

[Building a Nomad cluster on Raspberry Pi running Ubuntu server](https://johansiebens.dev/posts/2020/08/building-a-nomad-cluster-on-raspberry-pi-running-ubuntu-server/)

[Installing HashiCorp Vault on DigitalOcean with hashi-up](https://johansiebens.dev/posts/2020/12/installing-hashicorp-vault-on-digitalocean-with-hashi-up/)
