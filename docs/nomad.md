# `hashi-up` and Nomad

There are two ways to install Nomad with `hashi-up`.

By default, `hashi-up` will generate a Nomad configuration file with values based on the CLI flags.
Nomad has a lot of configuration options, and `hashi-up` tries to provide the most convenient parts as CLI options.
This allows you to install Nomad without creating a configuration file.

For more advanced use cases, use the `--config-file` flags to upload an existing configuration files, together with additional resources like certificates.

## Installing Nomad

### Set up a single Nomad server

Provision a new VM running a compatible operating system such as Ubuntu, Debian, Raspbian, or something else. Make sure that you opt-in to copy your registered SSH keys over to the new VM or host automatically.

> Note: You can copy ssh keys to a remote VM with `ssh-copy-id user@IP`.

Imagine the IP was `192.168.0.100` and the username was `ubuntu`, then you would run this:

```sh
export IP=192.168.0.100
hashi-up nomad install \
  --ssh-target-addr $IP \
  --ssh-target-user ubuntu \
  --server
```

When the command finishes, try to access Nomad using the UI at http://192.168.100:4646 or with the cli:

```sh
nomad agent-info -address==http://192.168.0.100:4646
```

### Join some agents to your Nomad server

Let's say you have a Nomad server up and running, now you can join one or more client agents to the cluster:

```sh
export SERVER_IP=192.168.0.100
export AGENT_1_IP=192.168.0.105
export AGENT_2_IP=192.168.0.106

hashi-up nomad install \
  --ssh-target-addr $AGENT_1_IP \
  --ssh-target-user ubuntu \
  --client \
  --retry-join $SERVER_IP
  
hashi-up nomad install \
  --ssh-target-addr $AGENT_1_IP \
  --ssh-target-user ubuntu \
  --client \
  --retry-join $SERVER_IP
```

### Create a multi-server (HA) setup

Prepare, for example, 3 nodes and let's say the have the following ip addresses:

- 192.168.0.100
- 192.168.0.101
- 192.168.0.102

With `hashi-up` it is quite easy to install 3 Nomad servers which will form a cluster:

```sh
export SERVER_1_IP=192.168.0.100
export SERVER_2_IP=192.168.0.101
export SERVER_3_IP=192.168.0.102

hashi-up nomad install \
  --ssh-target-addr $SERVER_1_IP \
  --ssh-target-user ubuntu \
  --server \
  --bootstrap-expect 3 \
  --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
  
hashi-up nomad install \
  --ssh-target-addr $SERVER_2_IP \
  --ssh-target-user ubuntu \
  --server \
  --bootstrap-expect 3 \
  --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
  
hashi-up nomad install \
  --ssh-target-addr $SERVER_3_IP \
  --ssh-target-user ubuntu \
  --server \
  --bootstrap-expect 3 \
  --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
```

And of course, joining client agents is the same as above:

```sh
export SERVER_1_IP=192.168.0.100
export SERVER_2_IP=192.168.0.101
export SERVER_3_IP=192.168.0.102
export AGENT_1_IP=192.168.0.105
export AGENT_2_IP=192.168.0.106

hashi-up nomad install \
  --ssh-target-addr $AGENT_1_IP \
  --ssh-target-user ubuntu \
  --client \
  --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
  
hashi-up nomad install \
  --ssh-target-addr $AGENT_2_IP \
  --ssh-target-user ubuntu \
  --client \
  --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
```

### Create a multi-server (HA) setup with Consul

If a Consul agent is already available on the Nomad nodes, Nomad can use Consul the automatically bootstrap the cluster.
So after installing a Consul cluster on all nodes, with `hashi-up` the cluster as explained above can be installed with the following commands:

```sh
export SERVER_1_IP=192.168.0.100
export SERVER_2_IP=192.168.0.101
export SERVER_3_IP=192.168.0.102
export AGENT_1_IP=192.168.0.105
export AGENT_2_IP=192.168.0.106

hashi-up nomad install \
  --ssh-target-addr $SERVER_1_IP \
  --ssh-target-user ubuntu \
  --server \
  --bootstrap-expect 3
   
hashi-up nomad install \
  --ssh-target-addr $SERVER_2_IP \
  --ssh-target-user ubuntu \
  --server \
  --bootstrap-expect 3

hashi-up nomad install \
  --ssh-target-addr $SERVER_3_IP \
  --ssh-target-user ubuntu \
  --server \
  --bootstrap-expect 3
    
hashi-up nomad install \
  --ssh-target-addr $AGENT_1_IP \
  --ssh-target-user ubuntu \
  --client
  
hashi-up nomad install \
  --ssh-target-addr $AGENT_2_IP \
  --ssh-target-user ubuntu \
  --client 
```

### Install Nomad with an existing configuration file

Perhaps you have already a configuration file available, or you want to configure Consul with some values which are not supported by the CLI options of `hashi-up`.
In this case, use the `--config-file` and the `--file` options to upload this config file and additional resources. When doing so, all other CLI options are ignored.

First create a config file and additional resources like certificates and keys, e.g. `server.hcl`:

```hcl
data_dir = "/op/nomad"

# Enable the server
server {
  enabled = true
  bootstrap_expect = 1
}

tls {
  http = true
  rpc  = true

  ca_file   = "/etc/nomad.d/nomad-ca.pem"
  cert_file = "/etc/nomad.d/server.pem"
  key_file  = "/etc/nomad.d/server-key.pem"

  verify_server_hostname = true
  verify_https_client    = true
}
```

Next, you can install Nomad with those resources:

```sh
export SERVER_IP=192.168.0.100

hashi-up nomad install \
    --ssh-target-addr $MASTER_NODE_IP \
    --ssh-target-user ubuntu \
    --config-file ./server.hcl \
    --file ./nomad-ca.pem \
    --file ./server.pem \
    --file ./server-key.pem
```

> Note that `hashi-up` will upload the additional resources to `/etc/nomad.d`

## What happens during installation?

During installation the following steps are executed on the target host

- download the Nomad distribution from https://releases.hashicorp.com and place the binary in `/usr/local/bin`
- create directories, like `/etc/nomad.d` and `/opt/nomad`
- generate or upload the config file to `/etc/nomad.d/nomad.hcl`
- upload other resources, like certificates, to `/etc/nomad.d`
- create a systemd service file for Nomad
- enable and start this new systemd service

## CLI options

```text
$ hashi-up nomad install --help
Usage:
  hashi-up nomad install [flags]

Flags:
      --acl                      Nomad: enables Nomad ACL system. (see Nomad documentation for more info)
      --address string           Nomad: the address the agent will bind to for all of its various network services. (see Nomad documentation for more info)
      --advertise string         Nomad: the address the agent will advertise to for all of its various network services. (see Nomad documentation for more info)
      --bootstrap-expect int     Nomad: sets server to expect bootstrap mode. (see Nomad documentation for more info) (default 1)
      --ca-file string           Nomad: the certificate authority used to check the authenticity of client and server connections. (see Nomad documentation for more info)
      --cert-file string         Nomad: the certificate to verify the agent's authenticity. (see Nomad documentation for more info)
      --client                   Nomad: enables the client mode of the agent. (see Nomad documentation for more info)
  -c, --config-file string       Custom Nomad configuration file to upload, setting this will disable config file generation meaning the other flags are ignored
      --datacenter string        Nomad: specifies the data center of the local agent. (see Nomad documentation for more info) (default "dc1")
      --encrypt string           Nomad: Provides the gossip encryption key. (see Nomad documentation for more info)
  -f, --file stringArray         Additional files, e.g. certificates, to upload
  -h, --help                     help for install
      --key-file string          Nomad: the key used with the certificate to verify the agent's authenticity. (see Nomad documentation for more info)
  -p, --package string           Upload and use this Nomad package instead of downloading
      --retry-join stringArray   Nomad: address of an agent to join at start time with retries enabled. Can be specified multiple times. (see Nomad documentation for more info)
      --server                   Nomad: enables the server mode of the agent. (see Nomad documentation for more info)
      --skip-enable              If set to true will not enable or start Nomad service
      --skip-start               If set to true will not start Nomad service
  -v, --version string           Version of Nomad to install

Global Flags:
      --local                    Running the installation locally, without ssh
      --ssh-target-addr string   Remote SSH target address (e.g. 127.0.0.1:22
      --ssh-target-key string    The ssh key to use for SSH login
      --ssh-target-user string   Username for SSH login (default "root")

```