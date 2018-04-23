# vaultops

[![GoDoc](https://godoc.org/github.com/milosgajdos83/vaultops?status.svg)](https://godoc.org/github.com/milosgajdos83/vaultops)
[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Travis CI](https://travis-ci.org/milosgajdos83/vaultops.svg?branch=master)](https://travis-ci.org/milosgajdos83/vaultops)
[![Go Report Card](https://goreportcard.com/badge/milosgajdos83/vaultops)](https://goreportcard.com/report/github.com/milosgajdos83/vaultops)

`vaultops` is a command line utility which aims to simplify [vault](https://www.vaultproject.io/) server setup.

# Motivation

Typical `vault` setup usually requires taking several steps before the server can be used:

- initializing `vault` server
- unsealing `vault` server(s)
- mounting `vault` backends
- creating `vault` backend roles
- generating SSL certificates
- configuring `vault` policies

The above listed tasks are usually performed using the `vault` cli tool which interacts with `vault` API. Using `vault` cli tool requires writing a lot of `shell` scripts which often grow into unmanageable full fledged monsters which are often hard to debug.

`vaultops` attempts to address this problem by providing a simple manifest file which can be used to specify all the tasks required to perform `vault` setup once the `vault` server/cluster is running. `vaultops` reads the manifest file and performs all the actions requsted by user by interacting interacts with `vault` REST API. The user can choose to perform a full setup or only selected setup actions.

# Quick start

Get the project:

```
$ go get -u github.com/milosgajdos83/vaultops
```

Get dependencies:
```
$ cd $GOPATH/github.com/milosgajdos83/vaultops && make dep
```

Run the tests

```
$ cd $GOPATH/src/github.com/milosgajdos83/vaultops
$ make test
for pkg in github.com/milosgajdos83/vaultops github.com/milosgajdos83/vaultops/command github.com/milosgajdos83/vaultops/manifest; do \
		go test -coverprofile="../../../$pkg/coverage.txt" -covermode=atomic $pkg || exit; \
	done
?   	github.com/milosgajdos83/vaultops	[no test files]
ok  	github.com/milosgajdos83/vaultops/command	0.554s	coverage: 17.7% of statements
ok  	github.com/milosgajdos83/vaultops/manifest	0.015s	coverage: 100.0% of statements
```

Build the binary:

```
$ make build
mkdir -p ./_build
go build -ldflags="-s -w" -o "./_build/vaultops"
```

Once you have pulled in all the project dependencies you can also build the binary by the familiar:

```
$ go build -ldflags="-s -w"
```

## Usage

As mentioned, `vaultop` provides various subcommands following the same UX as provided by `vault` cli utility. You can see the available commands below:

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

`vaultop` reads **the same environment variables** as `vault` cli tool, so you can rely on the familiar `$VAULT_` environment variables when specifying the `vault` server URLs and tokens.

# Manifest

`vaultops` allows you to define a manifest file which can be supplied to its commands. The manifest is a simple `YAML` file which specifies a list of `vault` hosts and `vault` resources that are requested to be setup in `vault`. A sample manifest file can be seen in [example](manifest/examples/example.yaml).

Let's look at a simple example:

`manifest.yaml` file:

```yaml
hosts:
  # URL of vault server to use for initialization
  init:
    - "http://10.100.21.161:8200"
  # URLs of all vault servers that should be unsealed
  unseal:
    - "http://10.100.21.161:8200"
    - "http://10.100.21.162:8200"
    - "http://10.100.21.163:8200"
# vault mounts
mounts:
  - path: "my-ca"
    type: "pki"
    max-lease-ttl: "876000h"
```

Hopefully the `YAML` snippet above is straightforward to understand :-)

Now that you've put together the manifest you can supply it to the `vaultops` via `-config` switch as seen below:

```
$ VAULT_ADDR="http://10.100.21.161:8200" ./vaultops setup -config manifest.yaml
[ info ] Attempting to initialize vault:
[ info ] 	10.100.21.161:8200
[ info ] Host: http://192.168.1.64:8200 initialized. Master keys:
[ info ] Key 1: XXX
[ info ] Key 2: XXX
[ info ] Key 3: XXX
[ info ] Key 4: XXX
[ info ] Key 5: XXX
[ info ] Initial Root Token: XXX
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
[ info ] Attempting to mount vault backends
[ info ] Attempting to mount pki backend in path: my-ca
[ info ] Successfully mounted pki in path: my-ca
[ info ] All requested vault backends successfully mounted
```

## vaultops commands

`vaultops` commands do not require using the complete setup manifest, but rather only a small subset of it.

For example, say you only want to mount some backends. Given the `vault` is initialized and unsealed, you can specify the following `vaultops` manifest and run the `mount` command as shown below:

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
[ info ] Finished mounting vault backends
```

Similarly, you can run `policy` command with only the policies part of the full manifest:

```yaml
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

Run the `policy` command:

```
$ VAULT_TOKEN="XXX" VAULT_ADDR="http://10.100.21.161:8200" ./vaultops policy -config policies.yaml
[ info ] Attempting to configure vault policies:
[ info ] 	k8s-ca
[ info ] Finished creating vault policies
```

You can find more examples in [examples](examples) subdirectory of the `vaultops` project.

### Security

By default `init` command, which is used to initialize the `vault` server stores **unencrypted** `vault` keys on the local filesystem in `.local` directory of your **current working directory** in a predefied `json` format which looks as follows:

```json
{
  "root_token": "XXX",
  "master_keys": [
    "XXX",
    "XXX",
    "XXX",
    "XXX",
    "XXX"
  ]
}
```

**This is not what you should do when unsealing your vault clusters unless you are unsealing a cluster for development pruposes! No sensitive data should ever touch the filesystem unencrypted! This option is available here for local development only!**

In non-local environment you can encrypt the vault keys using the keys provided by publc cloud providers such as [AWS KMS](https://aws.amazon.com/kms/) or [CGP cloudkms](https://cloud.google.com/kms/). These are available via `-kms-provider` cli switch. here is an example how to use GCP cloudkms to encrypt the vault keys:

```
$ VAULT_ADDR="http://10.100.21.161:8200" ./vaultops init -kms-provider="gcp" -gcp-kms-project="kube-blog" -gcp-kms-region="europe-west1" -gcp-kms-key-ring="vaultops" -gcp-kms-crypto-key="vaultops"
Cipher created[ info ] Attempting to initialize vault:
[ info ] 	http://10.100.21.161:8200
[ info ] Host: http://10.100.21.161:8200 initialized. Master keys:
[ info ] Key 1: XXX
[ info ] Key 2: XXX
[ info ] Key 3: XXX
[ info ] Key 4: XXX
[ info ] Key 5: XXX
[ info ] Initial Root Token: XXX
```

The keys are still stored locally in `./.local/vault.json` file but now they're encrypted using the `cloudkms` keys so if you try to inspect them you'll get a "garbage" pile of bytes. If you want to inspect the actual keys you need to decrypt them using the cloudkms keys.

**At the moment only AWS KMS and GCP KMS are supported**.

`vaultops` code can be extended to use your own KMS: PRs welcome :-)

As long as the `vault` key storage is concerned, `vaultops` currently only supports local storage, however extending `vaultops` for different storage should be fairly simple.

# TODO

**WE NEED WAY BIGGER TEST COVERAGE**

There is lots of redundant code so tonnes of refactoring up for grabs! 🤗

Only local key store is implemented at the moment, remote stores are up gor grabs! 🤗
