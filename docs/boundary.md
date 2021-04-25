# Installing Boundary with hashi-up

There are two ways to install Boundary with `hashi-up`.

By default, `hashi-up` will generate a Boundary configuration file with values based on the CLI flags.
Boundary has a lot of configuration options, and `hashi-up` tries to provide the most convenient parts as CLI options.
This allows you to install Boundary without creating a configuration file.

Especially regarding the KMS block, Boundary has support for many different implementations like AWS KMS, GCP Cloud KMS,...
Using the command-line flags, hashi-up only supports the AEAD implementation as the keys can be provided as CLI arguments.

For more advanced use cases, use the `--config-file` flags to upload an existing configuration files, together with additional resources like certificates.

## Installing Boundary

### Set up a single Boundary server

Provision a new VM running a compatible operating system such as Ubuntu, Debian, Raspbian, or something else. 
Make sure that you opt-in to copy your registered SSH keys over to the new VM or host automatically.

> Note: You can copy ssh keys to a remote VM with `ssh-copy-id user@IP`.

Imagine the IP was `192.168.0.100` and the username was `ubuntu`, then you would run this:

> The following steps assume that a Postgres server and database is available on the target host

```sh
export IP=192.168.0.100
export ROOT_KEY=JuZetUPcWEbsO4cTP7/E93O8Zrl6CQnYYb25CLFVAJM=
export WORKER_AUTH_KEY=uDuFypRqVPq7DM0Ujgrl5PoRRgqqy4eFdW7ORNIHHhU=
export RECOVERY_KEY=eixF4zdao9C6Soe+5jzBTtCnLUGMsP72NW5x30QP6So=

# initialize the database
hashi-up boundary init-database \
    --ssh-target-addr $IP \
    --db-url "postgresql://boundary:boundary123@localhost:5432/boundary?sslmode=disable" \
    --root-key $ROOT_KEY

# install Boundary running a controller and a worker
hashi-up boundary install \
    --ssh-target-addr $IP \
    --controller-name boundary-controller \
    --worker-name boundary-worker \
    --db-url "postgresql://boundary:boundary123@localhost:5432/boundary?sslmode=disable" \
    --cluster-addr 127.0.0.1 \
    --public-addr $IP \
    --root-key $ROOT_KEY \
    --worker-auth-key $WORKER_AUTH_KEY \
    --recovery-key $RECOVERY_KEY \
    --controller 127.0.0.1
```

The first command will initialize the database and prints out the detailed information to get you started with a default user and default targets.

When the commands finish, try to access Boundary using the UI at http://192.168.100:9200 or with the cli

### Create a multi-server (HA) setup

Boundary supports a multi-server mode for [high availability](https://www.boundaryproject.io/docs/installing/high-availability).
In this example we will create a single controller and two separate worker nodes.

Prepare, for example, 3 nodes and let's say they have the following ip addresses:

- 192.168.0.100
- 192.168.0.101
- 192.168.0.102

> The following steps assume that a Postgres server and database is available on the target host where we will install the Boundary Controller

First initialize the Boundary database:

``` sh
export CONTROLLER_IP=192.168.0.100
export ROOT_KEY=JuZetUPcWEbsO4cTP7/E93O8Zrl6CQnYYb25CLFVAJM=

hashi-up boundary init-database \
    --ssh-target-addr $CONTROLLER_IP \
    --db-url "postgresql://boundary:boundary123@localhost:5432/boundary?sslmode=disable" \
    --root-key $ROOT_KEY
```

Next, install a Boundary Controller on the first node:

``` sh
export CONTROLLER_IP=192.168.0.100
export ROOT_KEY=JuZetUPcWEbsO4cTP7/E93O8Zrl6CQnYYb25CLFVAJM=
export WORKER_AUTH_KEY=uDuFypRqVPq7DM0Ujgrl5PoRRgqqy4eFdW7ORNIHHhU=
export RECOVERY_KEY=eixF4zdao9C6Soe+5jzBTtCnLUGMsP72NW5x30QP6So=

hashi-up boundary install \
    --ssh-target-addr $CONTROLLER_IP \
    --controller-name boundary-controller \
    --db-url "postgresql://boundary:boundary123@localhost:5432/boundary?sslmode=disable" \
    --root-key $ROOT_KEY \
    --worker-auth-key $WORKER_AUTH_KEY \
    --cluster-addr $CONTROLLER_IP \
    --recovery-key $RECOVERY_KEY
```

And finally, install two Boundary Workers on the remaining nodes:

``` sh
export CONTROLLER_IP=192.168.0.100
export WORKER_01_IP=192.168.0.101
export WORKER_02_IP=192.168.0.102
export ROOT_KEY=JuZetUPcWEbsO4cTP7/E93O8Zrl6CQnYYb25CLFVAJM=
export WORKER_AUTH_KEY=uDuFypRqVPq7DM0Ujgrl5PoRRgqqy4eFdW7ORNIHHhU=
export RECOVERY_KEY=eixF4zdao9C6Soe+5jzBTtCnLUGMsP72NW5x30QP6So=

hashi-up boundary install \
    --ssh-target-addr $WORKER_01_IP \
    --worker-name boundary-worker-01 \
    --public-addr $WORKER_01_IP \
    --worker-auth-key $WORKER_AUTH_KEY \
    --controller $CONTROLLER_IP

hashi-up boundary install \
    --ssh-target-addr $WORKER2_PUBLIC_IP \
    --worker-name boundary-worker-01 \
    --public-addr $WORKER_02_IP \
    --worker-auth-key $WORKER_AUTH_KEY \
    --controller $CONTROLLER_IP
```

## What happens during installation?

During installation the following steps are executed on the target host

- download the Boundary distribution from https://releases.hashicorp.com and place the binary in `/usr/local/bin`
- create a `boundary` user and directories, like `/etc/boundary.d` and `/opt/boundary`
- generate or upload the config file to `/etc/boundary.d/boundary.hcl`
- upload other resources, like certificates, to `/etc/boundary.d`
- create a systemd service file for Boundary
- enable and start this new systemd service

## CLI options

```text
$ hashi-up boundary install --help
Usage:
  hashi-up boundary install [flags]

Flags:
      --api-addr string               Boundary: address for the API listener (default "0.0.0.0")
      --api-cert-file string          Boundary: specifies the path to the certificate for TLS.
      --api-key-file string           Boundary: specifies the path to the private key for the certificate.
      --cluster-addr string           Boundary: address for the Cluster listener (default "127.0.0.1")
      --cluster-cert-file string      Boundary: specifies the path to the certificate for TLS.
      --cluster-key-file string       Boundary: specifies the path to the private key for the certificate.
  -c, --config-file string            Custom Boundary configuration file to upload
      --controller stringArray        Boundary: a list of hosts/IP addresses and optionally ports for reaching controllers. (default [127.0.0.1])
      --controller-name string        Boundary: specifies a unique name of this controller within the Boundary controller cluster.
      --db-url string                 Boundary: configures the URL for connecting to Postgres
  -f, --file stringArray              Additional files, e.g. certificates, to upload
  -h, --help                          help for install
      --local                         Running the installation locally, without ssh
      --package string                Upload and use this Boundary package instead of downloading
      --proxy-addr string             Boundary: address for the Proxy listener (default "0.0.0.0")
      --proxy-cert-file string        Boundary: specifies the path to the certificate for TLS.
      --proxy-key-file string         Boundary: specifies the path to the private key for the certificate.
      --public-addr string            Boundary: specifies the public host or IP address (and optionally port) at which the worker can be reached by clients for proxying.
      --public-cluster-addr string    Boundary: specifies the public host or IP address (and optionally port) at which the controller can be reached by workers.
      --recovery-key string           Boundary: KMS key is used for rescue/recovery operations that can be used by a client to authenticate almost any operation within Boundary.
      --root-key string               Boundary: a KEK (Key Encrypting Key) for the scope-specific KEKs (also referred to as the scope's root key).
      --skip-enable                   If set to true will not enable or start Boundary service
      --skip-start                    If set to true will not start Boundary service
  -r, --ssh-target-addr string        Remote SSH target address (e.g. 127.0.0.1:22
  -k, --ssh-target-key string         The ssh key to use for SSH login
  -p, --ssh-target-password string    The ssh password to use for SSH login
  -s, --ssh-target-sudo-pass string   The ssh password to use for SSH login
  -u, --ssh-target-user string        Username for SSH login (default "root")
  -v, --version string                Version of Boundary to install
      --worker-auth-key string        Boundary: KMS key shared by the Controller and Worker in order to authenticate a Worker to the Controller.
      --worker-name string            Boundary: specifies a unique name of this worker within the Boundary worker cluster.
```