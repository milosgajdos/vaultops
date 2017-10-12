# vaultops

`vaultops` is a simple command line utility which aims to simplify complex vault server setups. It provides a few subcommands that allow you to set up vault policies, mounts, roles and generate SSL certificates for a given `pki` vault backend with a single command.

**[NOTE] This utility is a PoC. Use it at your own risk!**

# Motivation

`vault` setup often requires a lot of tasks: mounting different backend types, creating roles, generating [root] certificates and configuring policies. All of these tasks can be done via the brilliant `vault` cli utility, however using it requires writing a lot of `shell` scripts without proper error handling etc.

This utility provides a simple manifest file that can perform full `vault` setup running a single command or by running a few subcommands should you decide to do so. It iteracts directly with `vault` API and performs various tasks concurrently.

# Usage

See the output below:

```
Usage: vaultops [--version] [--help] <command> [<args>]

Available commands are:
    backend    Manage vault backends
    init       Initialize Vault cluster or server
    mount      Mount a new vault secret backend
    policy     Manage vault policies
    setup      Setup a new Vault server
    unseal     Unseal a Vault server
```

The utility reads the same environment variables as `vault` utility so you can rely on the familiar `$VAULT_` environment variables when interacting with your `vault` server.

## Manifest

`vaultops` allows you to define a manifest file which you can supply to particular subcommand. The manifest isa simple `YAML` file which contains a list of resources that are requested to be created in `vault`. A skeleton of full manifest file can be seen below:

```yaml
hosts:
  - "http://10.100.21.161:8200"
  - "http://10.100.21.162:8200"
  - "http://10.100.21.163:8200"
# vault mounts
mounts:
  - path: "k8s-ca"
    type: "pki"
    max-lease-ttl: "876000h"
# vault backends
backends:
  - name: "k8s-ca"
    # vault roles to create
    roles:
      - name: "api-server"
        allowed-domains: "api-server,kubernetes,kubernetes.default"
        allow-bare-domains: true
        allow-any-name: true
        organization: "system:control-plane"
    # certificates to generate
    certificates:
      - name: "root-ca"
        root: true
        common-name: "k8s-ca"
        ttl: "876000h"
        type: "internal"
      - name: "k8s-ca"
        common-name: "api-server"
        ttl: "875000h"
        ip-sans: "10.101.0.1,127.0.0.1"
        alt-names: "kubernetes.default.svc.cluster.local"
        role: "api-server"
        store: true
# vault policies to configure
policies:
  - name: "k8s-ca"
    rules: |
      path "secret/k8s-ca/*" {
         policy = "write"
      }
      path "k8s-ca/roles/*" {
         policy = "write"
      }
      path "k8s-ca/issue/*" {
        policy = "write"
      }
```

See below for a simple example of using the tool:

`manifest.yaml` file:

```yaml
hosts:
  - "http://10.100.21.161:8200"
  - "http://10.100.21.162:8200"
  - "http://10.100.21.163:8200"
# vault mounts
mounts:
  - path: "my-ca"
    type: "pki"
    max-lease-ttl: "876000h"
```

You can supply the above manifest to the `vaultops` as follows:

```
$ VAULT_ADDR="http://10.100.21.161:8200" ./vaultops setup -config manifest.yaml
[ info ] Attempting to initialize vault cluster
[ info ] Vault successfully initialized
[ info ] Key 1: XXX
[ info ] Key 2: XXX
[ info ] Key 3: XXX
[ info ] Key 4: XXX
[ info ] Key 5: XXX
[ info ] Root Token: XXX
[ info ]
[ info ] Attempting to unseal vault cluster
[ info ] Attempting to unseal host: "http://10.100.21.161:8200"
[ info ] Host http://10.100.21.161:8200:
	Sealed: false
	Key Shares: 5
	Key Threshold: 3
	Unseal Progress: 0
	Unseal Nonce:
[ info ] Host http://10.100.21.162:8200:
	Sealed: false
	Key Shares: 5
	Key Threshold: 3
	Unseal Progress: 0
	Unseal Nonce:
[ info ] Host http://10.100.21.163:8200:
	Sealed: false
	Key Shares: 5
	Key Threshold: 3
	Unseal Progress: 0
	Unseal Nonce:
[ info ] Vault successfully unsealed
[ info ]
[ info ] Attempting to mount vault backends
[ info ] Attempting to mount pki backend in path: my-ca
[ info ] Successfully mounted pki in path: my-ca
[ info ] All requested vault backends successfully mounted
```


### Security

As you may have noticed the tool unseals `vault` cluster by unsealing every `vault` server in the cluster. **This is not what you should do in real life**. `init` subcommand is here for a convenience you should normally avoid doing by all means to avoid compromising the server/cluster.

Futhermore, for convenience `vault init` stores the `vault` keys in `.local` directory in a predefined `json` file. This will become optional in the further releases. It's there at the moment out of pure convenience for operator.

## Subcommands

Most of the `vaultops` subcommands do not require using full manifest. You can simply pick particular resources and run the tool just against those. For example, say you want to mount some backends. You can specify the following manifest and use the `mount` subcommand as below.

`mounts.yaml` example:

```yaml
mounts:
  - path: "k8s-ca"
    type: "pki"
    max-lease-ttl: "876000h"
```

Run `mount` subcommand:

```
$ VAULT_ADDR="http://10.100.21.161:8200" ./vaultops mount -config mounts.yaml
[ info ] Attempting to mount vault backends:
[ info ] 	Type: pki Path: k8s-ca TTL: 876000h
[ info ] Attempting to mount pki backend in path: k8s-ca
[ info ] Successfully mounted k8s-ca in path pki
[ info ] All requested vault backends successfully mounted
```

# TODO

Dont get me started on this.... ;-)
