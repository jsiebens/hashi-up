# `hashi-up` and Vault

There are two ways to install Vault with `hashi-up`.

By default, `hashi-up` will generate a Vault configuration file with values based on the CLI flags.
Vault has a lot of configuration options, and `hashi-up` tries to provide the most convenient parts as CLI options.
This allows you to install Vault without creating a configuration file.

For more advanced use cases, use the `--config-file` flags to upload an existing configuration files, together with additional resources like certificates.

## Installing Vault

### Set up a single Vault server with the Filesystem Storage Backend

Provision a new VM running a compatible operating system such as Ubuntu, Debian, Raspbian, or something else. Make sure that you opt-in to copy your registered SSH keys over to the new VM or host automatically.

> Note: You can copy ssh keys to a remote VM with `ssh-copy-id user@IP`.

Imagine the IP was `192.168.0.100` and the username was `ubuntu`, then you would run this:

```sh
export IP=192.168.0.100
hashi-up vault install \
  --ssh-target-addr $IP \
  --ssh-target-user ubuntu \
  --storage file
```

When the command finishes, try to access Vault using the UI at http://192.168.100:8200 or with the cli:

```sh
vault status -address=http://192.168.100:8200
```

### Create a multi-server (HA) setup

Vault supports a multi-server mode for high availability. At this moment, `hashi-up` only support creating such a HA setup with the Consul Storage Backend.
If you want to use other storage backends, you can always install Vault with a more advanced configuration file. 

Prepare, for example, 3 nodes and let's say the have the following ip addresses:

- 192.168.0.100
- 192.168.0.101
- 192.168.0.102

First install a Consul cluster on the nodes, as explained [here](consul.md).

Next, with `hashi-up` it is quite easy to install 3 Vault servers which will form a cluster:

```sh
export SERVER_1_IP=192.168.0.100
export SERVER_2_IP=192.168.0.101
export SERVER_3_IP=192.168.0.102

hashi-up vault install \
    --ssh-target-addr $SERVER_1_IP \
    --ssh-target-user ubuntu \
    --storage consul \
    --api-addr http://$SERVER_1_IP:8200

hashi-up vault install \
    --ssh-target-addr $SERVER_2_IP \
    --ssh-target-user ubuntu \
    --storage consul \
    --api-addr http://$SERVER_2_IP:8200

hashi-up vault install \
    --ssh-target-addr $SERVER_3_IP \
    --ssh-target-user ubuntu \
    --storage consul \
    --api-addr http://$SERVER_3_IP:8200
```

### Install Vault with an existing configuration file

Perhaps you have already a configuration file available, or you want to configure Vault with some values which are not supported by the CLI options of `hashi-up`.
In this case, use the `--config-file` and the `--file` options to upload this config file and additional resources. When doing so, all other CLI options are ignored.

First create a config file and additional resources like certificates and keys, e.g. `server.hcl`:

```hcl
ui = true
listener "tcp" {
  address       = "0.0.0.0:8200"
  tls_cert_file = "/etc/vault.d/server.pem"
  tls_key_file  = "/etc/vault.d/server-key.pem"
}

api_addr = "https://vault-leader.my-company.internal"

storage "gcs" {
  bucket        = "mycompany-vault-data"
  ha_enabled    = "true"
}
```

Next, you can install Vault with those resources:

```sh
export SERVER_IP=192.168.0.100

hashi-up vault install \
    --ssh-target-addr $MASTER_NODE_IP \
    --ssh-target-user ubuntu \
    --config-file ./server.hcl \
    --file ./server.pem \
    --file ./server-key.pem
```

> Note that `hashi-up` will upload the additional resources to `/etc/vault.d`

## What happens during installation?

During installation the following steps are executed on the target host

- download the Vault distribution from https://releases.hashicorp.com and place the binary in `/usr/local/bin`
- create a `vault` user and directories, like `/etc/vault.d` and `/opt/vault`
- generate or upload the config file to `/etc/vault.d/vault.hcl`
- upload other resources, like certificates, to `/etc/vault.d`
- create a systemd service file for Vault
- enable and start this new systemd service

## CLI options

```text
$ hashi-up vault install --help
Usage:
  hashi-up vault install [flags]

Flags:
      --address stringArray           Vault: the address to bind to for listening. (see Vault documentation for more info) (default [0.0.0.0:8200])
      --api-addr string               Vault: the address (full URL) to advertise to other Vault servers in the cluster for client redirection. (see Vault documentation for more info)
      --cert-file string              Vault: the certificate for TLS. (see Vault documentation for more info)
      --cluster-addr string           Vault: the address to advertise to other Vault servers in the cluster for request forwarding. (see Vault documentation for more info)
  -c, --config-file string            Custom Vault configuration file to upload, setting this will disable config file generation meaning the other flags are ignored
      --consul-addr string            Vault: the address of the Consul agent to communicate with. (see Vault documentation for more info) (default "127.0.0.1:8500")
      --consul-path string            Vault: the path in Consul's key-value store where Vault data will be stored. (see Vault documentation for more info) (default "vault/")
      --consul-tls-ca-file string     Vault: the path to the CA certificate used for Consul communication. (see Vault documentation for more info)
      --consul-tls-cert-file string   Vault: the path to the certificate for Consul communication. (see Vault documentation for more info)
      --consul-tls-key-file string    Vault: the path to the private key for Consul communication. (see Vault documentation for more info)
      --consul-token string           Vault: the Consul ACL token with permission to read and write from the path in Consul's key-value store. (see Vault documentation for more info)
  -f, --file stringArray              Additional files, e.g. certificates, to upload
  -h, --help                          help for install
      --key-file string               Vault: the private key for the certificate. (see Vault documentation for more info)
  -p, --package string                Upload and use this Vault package instead of downloading
      --skip-enable                   If set to true will not enable or start Vault service
      --skip-start                    If set to true will not start Vault service
      --storage string                Vault: the type of storage backend. Currently only "file" of "consul" is supported. (see Vault documentation for more info) (default "file")
  -v, --version string                Version of Vault to install

Global Flags:
      --local                    Running the installation locally, without ssh
      --ssh-target-addr string   Remote SSH target address (e.g. 127.0.0.1:22
      --ssh-target-key string    The ssh key to use for SSH login
      --ssh-target-user string   Username for SSH login (default "root")
```