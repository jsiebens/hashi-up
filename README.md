# hashi-up

hashi-up is a lightweight utility to install HashiCorp [Consul](https://www.consul.io/), [Nomad](https://www.nomadproject.io) or [Vault](https://www.vaultproject.io/) on any remote Linux host. All you need is `ssh` access and the binary `hashi-up` to build a Consul, Nomad or Vault cluster.

The tool is written in Go and is cross-compiled for Linux, Windows, MacOS and even on Raspberry Pi.

This project is heavily inspired on the work of [Alex Ellis](https://www.alexellis.io/) who created [k3sup](https://k3sup.dev/), a tool to to get from zero to KUBECONFIG with [k3s](https://k3s.io/)

[![Go Report Card](https://goreportcard.com/badge/github.com/jsiebens/hashi-up)](https://goreportcard.com/report/github.com/jsiebens/hashi-up)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![GitHub All Releases](https://img.shields.io/github/downloads/jsiebens/hashi-up/total)

## What's this for?

This tool uses `ssh` to install HashiCorp Consul, Nomad or Vault to a remote Linux host. You can also use it to join existing Linux hosts into a Consul, Nomad or Vault cluster. First, Consul, Nomad or Vault is installed using a utility script, along with a minimal configuration to run the agent as server or client.

`hashi-up` was developed to automate what can be a very manual and confusing process for many developers, who are already short on time. Once you've provisioned a VM with your favourite tooling, `hashi-up` means you are only 60 seconds away from running `nomad status` on your own computer.

## Installing

`hashi-up` is distributed as a static Go binary. 
You can use the installer on MacOS and Linux, or visit the Releases page to download the executable for Windows.

``` shell
curl -sLS https://get.hashi-up.dev | sh
sudo install hashi-up /usr/local/bin/

hashi-up version
```

## Usage

The `hashi-up` tool is a client application which you can run on your own computer. It uses SSH to connect to remote servers when installing HashiCorp Consul or Nomad. Binaries are provided for MacOS, Windows, and Linux (including ARM).

### Consul

#### Setup a single Consul server

Provision a new VM running a compatible operating system such as Ubuntu, Debian, Raspbian, or something else. Make sure that you opt-in to copy your registered SSH keys over to the new VM or host automatically.

> Note: You can copy ssh keys to a remote VM with `ssh-copy-id user@IP`.

Imagine the IP was `192.168.0.100` and the username was `ubuntu`, then you would run this:

```sh
export IP=192.168.0.100
hashi-up consul install --ssh-target-addr $IP --ssh-target-user ubuntu --server --client 0.0.0.0
```

When the command finishes, try to access Consul using the UI at http://192.168.100:8500 or with the cli:

```
consul members -http-addr=http://192.168.0.100:8500
```

Other additional flags for `install`:

- `--version string` -           version of Consul to install, default to latest available
- `--advertise string` -         sets the advertise address to use.
- `--bind string` -              sets the bind address for cluster communication.
- `--bootstrap-expect int` -     sets server to expect bootstrap mode. (default 1)
- `--client string` -            sets the address to bind for client access.
- `--dc string` -                specifies the data center of the local agent. (default "dc1")
- `--retry-join` -               address of an agent to join at start time with retries enabled. Can be specified multiple times.
- `--server` -                   switches agent to server mode.


#### Join some agents to your Consul server

Let's say you have a Consul server up and running, now you can join one or more client agents to the cluster:

```sh
export SERVER_IP=192.168.0.100
export AGENT_1_IP=192.168.0.105
export AGENT_2_IP=192.168.0.106

hashi-up consul install --ssh-target-addr $AGENT_1_IP --ssh-target-user ubuntu --client 0.0.0.0 --retry-join $SERVER_IP
hashi-up consul install --ssh-target-addr $AGENT_2_IP --ssh-target-user ubuntu --client 0.0.0.0 --retry-join $SERVER_IP
```

#### Create a multi-server (HA) setup

Prepare, for example, 3 nodes and let's say the have the following ip addresses:

- 192.168.0.100
- 192.168.0.101
- 192.168.0.102

With `hashi-up` it is quite easy to install 3 Consul servers which will form a cluster:

```sh
export SERVER_1_IP=192.168.0.100
export SERVER_2_IP=192.168.0.101
export SERVER_3_IP=192.168.0.102

hashi-up consul install --ssh-target-addr $SERVER_1_IP --ssh-target-user ubuntu --server --client 0.0.0.0 --bootstrap-expect 3 --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
hashi-up consul install --ssh-target-addr $SERVER_2_IP --ssh-target-user ubuntu --server --client 0.0.0.0 --bootstrap-expect 3 --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
hashi-up consul install --ssh-target-addr $SERVER_3_IP --ssh-target-user ubuntu --server --client 0.0.0.0 --bootstrap-expect 3 --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
```

And of course, joining client agents is the same as above:

```sh
export SERVER_1_IP=192.168.0.100
export SERVER_2_IP=192.168.0.101
export SERVER_3_IP=192.168.0.102
export AGENT_1_IP=192.168.0.105
export AGENT_2_IP=192.168.0.106

hashi-up consul install --ssh-target-addr $AGENT_1_IP --ssh-target-user ubuntu --client 0.0.0.0 --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
hashi-up consul install --ssh-target-addr $AGENT_2_IP --ssh-target-user ubuntu --client 0.0.0.0 --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
```

### Nomad

#### Setup a single Nomad server

Provision a new VM running a compatible operating system such as Ubuntu, Debian, Raspbian, or something else. Make sure that you opt-in to copy your registered SSH keys over to the new VM or host automatically.

> Note: You can copy ssh keys to a remote VM with `ssh-copy-id user@IP`.

Imagine the IP was `192.168.0.100` and the username was `ubuntu`, then you would run this:

```sh
export IP=192.168.0.100
hashi-up nomad install --ssh-target-addr $IP --ssh-target-user ubuntu --server
```

When the command finishes, try to access Nomad using the UI at http://192.168.100:4646 or with the cli:

```sh
nomad agent-info -address==http://192.168.0.100:4646
```

Other additional flags for `install`:

- `--version` -           version of Nomad to install, default to latest available
- `--address` -           the address the agent will bind to for all of its various network services.
- `--advertise` -         the address the agent will advertise to for all of its various network services.
- `--bootstrap-expect` -  sets server to expect bootstrap mode. (default 1)
- `--client` -            enables the client mode of the agent.
- `--dc` -                specifies the data center of the local agent. (default "dc1")
- `--retry-join` -        address of an agent to join at start time with retries enabled. Can be specified multiple times.
- `--server` -            enables the server mode of the agent.

#### Join some agents to your Nomad server

Let's say you have a Nomad server up and running, now you can join one or more client agents to the cluster:

```sh
export SERVER_IP=192.168.0.100
export AGENT_1_IP=192.168.0.105
export AGENT_2_IP=192.168.0.106

hashi-up nomad install --ssh-target-addr $AGENT_1_IP --ssh-target-user ubuntu --client --retry-join $SERVER_IP
hashi-up nomad install --ssh-target-addr $AGENT_2_IP --ssh-target-user ubuntu --client --retry-join $SERVER_IP
```

#### Create a multi-server (HA) setup

Prepare, for example, 3 nodes and let's say the have the following ip addresses:

- 192.168.0.100
- 192.168.0.101
- 192.168.0.102

With `hashi-up` it is quite easy to install 3 Nomad servers which will form a cluster:

```sh
export SERVER_1_IP=192.168.0.100
export SERVER_2_IP=192.168.0.101
export SERVER_3_IP=192.168.0.102

hashi-up nomad install --ssh-target-addr $SERVER_1_IP --ssh-target-user ubuntu --server --bootstrap-expect 3 --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
hashi-up nomad install --ssh-target-addr $SERVER_2_IP --ssh-target-user ubuntu --server --bootstrap-expect 3 --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
hashi-up nomad install --ssh-target-addr $SERVER_3_IP --ssh-target-user ubuntu --server --bootstrap-expect 3 --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
```

And of course, joining client agents is the same as above:

```sh
export SERVER_1_IP=192.168.0.100
export SERVER_2_IP=192.168.0.101
export SERVER_3_IP=192.168.0.102
export AGENT_1_IP=192.168.0.105
export AGENT_2_IP=192.168.0.106

hashi-up nomad install --ssh-target-addr $AGENT_1_IP --ssh-target-user ubuntu --client --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
hashi-up nomad install --ssh-target-addr $AGENT_2_IP --ssh-target-user ubuntu --client --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
```

#### Create a multi-server (HA) setup with Consul

If a Consul agent is already available on the Nomad nodes, Nomad can use Consul the automatically bootstrap the cluster.
So after installing a Consul cluster on all nodes, with `hashi-up` the cluster as explained above can be installed with the following commands:

```sh
export SERVER_1_IP=192.168.0.100
export SERVER_2_IP=192.168.0.101
export SERVER_3_IP=192.168.0.102
export AGENT_1_IP=192.168.0.105
export AGENT_2_IP=192.168.0.106

hashi-up nomad install --ssh-target-addr $SERVER_1_IP --ssh-target-user ubuntu --server --bootstrap-expect 3 
hashi-up nomad install --ssh-target-addr $SERVER_2_IP --ssh-target-user ubuntu --server --bootstrap-expect 3 
hashi-up nomad install --ssh-target-addr $SERVER_3_IP --ssh-target-user ubuntu --server --bootstrap-expect 3 
hashi-up nomad install --ssh-target-addr $AGENT_1_IP --ssh-target-user ubuntu --client
hashi-up nomad install --ssh-target-addr $AGENT_2_IP --ssh-target-user ubuntu --client 
```

## If your ssh-key is password-protected

If the ssh-key is encrypted the first step is to try to connect to the ssh-agent. If this works, it will be used to connect to the server.
If the ssh-agent is not running, the user will be prompted for the password of the ssh-key.

On most Linux systems and MacOS, ssh-agent is automatically configured and executed at login. No additional actions are required to use it.

To start the ssh-agent manually and add your key run the following commands:

```bash
eval `ssh-agent`
ssh-add ~/.ssh/id_rsa
```

You can now just run hashi-up as usual. No special parameters are necessary. 

## Resources

[Deploying a highly-available Nomad cluster with hashi-up!](https://johansiebens.dev/posts/2020/07/deploying-a-highly-available-nomad-cluster-with-hashi-up/)

[Building a Nomad cluster on Raspberry Pi running Ubuntu server](https://johansiebens.dev/posts/2020/08/building-a-nomad-cluster-on-raspberry-pi-running-ubuntu-server/)

[Installing HashiCorp Vault on DigitalOcean with hashi-up](https://johansiebens.dev/posts/2020/12/installing-hashicorp-vault-on-digitalocean-with-hashi-up/)
