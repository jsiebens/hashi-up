# Installing Consul with hashi-up

There are two ways to install Consul with `hashi-up`.

By default, `hashi-up` will generate a Consul configuration file with values based on the CLI flags.
Consul has a lot of configuration options, and `hashi-up` tries to provide the most convenient parts as CLI options. 
This allows you to install Consul without creating a configuration file.

For more advanced use cases, use the `--config-file` flags to upload an existing configuration files, together with additional resources like certificates.

## Installing Consul

### Set up a single Consul server

Provision a new VM running a compatible operating system such as Ubuntu, Debian, Raspbian, or something else. Make sure that you opt-in to copy your registered SSH keys over to the new VM or host automatically.

> Note: You can copy ssh keys to a remote VM with `ssh-copy-id user@IP`.

Imagine the IP was `192.168.0.100` and the username was `ubuntu`, then you would run this:

```sh
export IP=192.168.0.100

hashi-up consul install \
    --ssh-target-addr $IP \
    --ssh-target-user ubuntu \
    --server \
    --client-addr 0.0.0.0
```

When the command finishes, try to access Consul using the UI at http://192.168.100:8500 or with the cli:

```
consul members -http-addr=http://192.168.0.100:8500
```

### Join some agents to your Consul server

Now that you have a Consul server up and running, you can join one or more client agents to the cluster:

```sh
export SERVER_IP=192.168.0.100
export AGENT_1_IP=192.168.0.105
export AGENT_2_IP=192.168.0.106

hashi-up consul install \
    --ssh-target-addr $AGENT_1_IP \
    --ssh-target-user ubuntu \
    --retry-join $SERVER_IP

hashi-up consul install \
    --ssh-target-addr $AGENT_2_IP \
    --ssh-target-user ubuntu \
    --retry-join $SERVER_IP
```

### Create a multi-server (HA) setup

Prepare, for example, 3 nodes and let's say the have the following ip addresses:

- 192.168.0.100
- 192.168.0.101
- 192.168.0.102

With `hashi-up` it is quite easy to install 3 Consul servers which will form a cluster:

```sh
export SERVER_1_IP=192.168.0.100
export SERVER_2_IP=192.168.0.101
export SERVER_3_IP=192.168.0.102

hashi-up consul install \
  --ssh-target-addr $SERVER_1_IP \
  --ssh-target-user ubuntu \
  --server \
  --client-addr 0.0.0.0 \
  --bootstrap-expect 3 \
  --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
  
hashi-up consul install \
  --ssh-target-addr $SERVER_2_IP \
  --ssh-target-user ubuntu \
  --server \
  --client-addr 0.0.0.0 \
  --bootstrap-expect 3 \
  --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
  
hashi-up consul install \
  --ssh-target-addr $SERVER_3_IP \
  --ssh-target-user ubuntu \
  --server \
  --client-addr 0.0.0.0 \
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

hashi-up consul install \
  --ssh-target-addr $AGENT_1_IP \
  --ssh-target-user ubuntu \
  --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP

hashi-up consul install \
  --ssh-target-addr $AGENT_2_IP \
  --ssh-target-user ubuntu \
  --retry-join $SERVER_1_IP --retry-join $SERVER_2_IP --retry-join $SERVER_3_IP
```

### Install Consul with an existing configuration file

Perhaps you have already a configuration file available, or you want to configure Consul with some values which are not supported by the CLI options of `hashi-up`.
In this case, use the `--config-file` and the `--file` options to upload this config file and additional resources. When doing so, all other CLI options are ignored.

First create a config file and additional resources like certificates and keys, e.g. `server.hcl`:

```hcl
datacenter  = "dc1"
data_dir    = "/opt/consul"
client_addr = "0.0.0.0"

ports {
  grpc  = 8502
  https = 8501
  http  = -1
}

server                 = true
bootstrap_expect       = 3
ca_file                = "/etc/consul.d/ca.pem"
cert_file              = "/etc/consul.d/server.pem"
key_file               = "/etc/consul.d/server-key.pem"
verify_incoming_rpc    = true
verify_outgoing        = true
verify_server_hostname = true

connect {
  enabled = true
}
```

Next, you can install Consul with those resources:

```sh
export SERVER_IP=192.168.0.100

hashi-up consul install \
    --ssh-target-addr $MASTER_NODE_IP \
    --ssh-target-user ubuntu \
    --config-file ./server.hcl \
    --file ./ca.pem \
    --file ./server.pem \
    --file ./server-key.pem
```

> Note that `hashi-up` will upload the additional resources to `/etc/consul.d`

## What happens during installation?

During installation the following steps are executed on the target host

- download the Consul distribution from https://releases.hashicorp.com and place the binary in `/usr/local/bin`
- create a `consul` user and directories, like `/etc/consul.d` and `/opt/consul`
- generate or upload the config file to `/etc/consul.d/consul.hcl`
- upload other resources, like certificates, to `/etc/consul.d`
- create a systemd service file for Consul
- enable and start this new systemd service

## CLI options

```text
$ hashi-up consul install --help
Install Consul on a server via SSH

Usage:
  hashi-up consul install [flags]

Flags:
      --acl                           Consul: enables Consul ACL system. (see Consul documentation for more info)
      --advertise-addr string         Consul: sets the advertise address to use. (see Consul documentation for more info)
      --agent-token string            Consul: the token that the agent will use for internal agent operations.. (see Consul documentation for more info)
      --auto-encrypt                  Consul: this option enables auto_encrypt and allows servers to automatically distribute certificates from the Connect CA to the clients. (see Consul documentation for more info)
      --bind-addr string              Consul: sets the bind address for cluster communication. (see Consul documentation for more info)
      --bootstrap-expect int          Consul: sets server to expect bootstrap mode. 0 are less disables bootstrap mode. (see Consul documentation for more info) (default 1)
      --ca-file string                Consul: the certificate authority used to check the authenticity of client and server connections. (see Consul documentation for more info)
      --cert-file string              Consul: the certificate to verify the agent's authenticity. (see Consul documentation for more info)
      --client-addr string            Consul: sets the address to bind for client access. (see Consul documentation for more info)
  -c, --config-file string            Custom Consul configuration file to upload, setting this will disable config file generation meaning the other flags are ignored
      --connect                       Consul: enables the Connect feature on the agent. (see Consul documentation for more info)
      --datacenter string             Consul: specifies the data center of the local agent. (see Consul documentation for more info) (default "dc1")
      --dns-addr string               Consul: sets the address for the DNS server. (see Consul documentation for more info)
      --encrypt string                Consul: provides the gossip encryption key. (see Consul documentation for more info)
  -f, --file strings                  Additional files, e.g. certificates, to upload
      --grpc-addr string              Consul: sets the address for the gRPC API server. (see Consul documentation for more info)
  -h, --help                          help for install
      --http-addr string              Consul: sets the address for the HTTP API server. (see Consul documentation for more info)
      --https-addr string             Consul: sets the address for the HTTPS API server. (see Consul documentation for more info)
      --https-only                    Consul: if true, HTTP port is disabled on both clients and servers and to only accept HTTPS connections when TLS enabled. (default true)
      --key-file string               Consul: the key used with the certificate to verify the agent's authenticity. (see Consul documentation for more info)
      --local                         Running the installation locally, without ssh
      --package string                Upload and use this Consul package instead of downloading
      --retry-join strings            Consul: address of an agent to join at start time with retries enabled. Can be specified multiple times. (see Consul documentation for more info)
      --server                        Consul: switches agent to server mode. (see Consul documentation for more info)
      --skip-enable                   If set to true will not enable or start Consul service
      --skip-start                    If set to true will not start Consul service
  -r, --ssh-target-addr string        Remote SSH target address (e.g. 127.0.0.1:22
  -k, --ssh-target-key string         The ssh key to use for SSH login
  -p, --ssh-target-password string    The ssh password to use for SSH login
  -s, --ssh-target-sudo-pass string   The ssh password to use for SSH login
  -u, --ssh-target-user string        Username for SSH login (default "root")
  -v, --version string                Version of Consul to install
```