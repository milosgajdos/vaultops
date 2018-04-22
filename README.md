# vaultops

[![GoDoc](https://godoc.org/github.com/milosgajdos83/vaultops?status.svg)](https://godoc.org/github.com/milosgajdos83/vaultops)
[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Travis CI](https://travis-ci.org/milosgajdos83/vaultops.svg?branch=master)](https://travis-ci.org/milosgajdos83/vaultops)
[![Go Report Card](https://goreportcard.com/badge/milosgajdos83/vaultops)](https://goreportcard.com/report/github.com/milosgajdos83/vaultops)

`vaultops` is a simple command line utility which aims to simplify complex [vault](https://www.vaultproject.io/) server setups. The tool provides a few subcommands which allows the user to performa various setup tasks (creating vault policies, mounting backends etc.) using a single command.

# Motivation

`vault` setup typically requires a lot of manual tasks:
- initializing vault server
- unsealing vault server(s) in vault cluster
- mounting vault backends
- creating vault backend roles
- generating SSL certificates
- configuring vault policies

The above listed tasks are usually performed using the `vault` cli tool. This requires writing a lot of `shell` scripts which often grow into unmanageable full fledged monsters which have no type checking or proper error handling.

`vaultops` utility addresses this problem by providing a simple manifest file which can be used to specify all the tasks required to perform `vault` setup once the `vault` cluster is running. `vaultops` reads the manifest file and performs all the actions requsted by user. `vaultops` interacts with `vault` via REST API and performs the tasks concurrently whenever possible. The user can choose to perform a full setup or only selected setup actions.

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

## Usage

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

`vaultop` reads the same environment variables like `vault` cli tool, so you can rely on the familiar `$VAULT_` environment variables when specifying the `vault` server URLs and authentication tokens.

# Manifest

`vaultops` allows you to define a manifest file which can be supplied to its commands. The manifest is a simple `YAML` file which specifies a list of `vault` hosts and `vault` resources that are requested to be created in `vault`. A sample manifest file can be seen in [example](manifest/examples/example.yaml).

Let's look at a simple example:

`manifest.yaml` file:

```yaml
hosts:
  init:
    - "http://10.100.21.161:8200"
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

### Security

By default `init` command stores **unencrypted** `vault` keys on local filesystem in `.local` directory of your **current working directory** in a predefied `json` format which looks as follows:

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

**This is not what you should do when unsealing your vault clusters unless you are unsealing a cluster for development pruposes! Remember, no sensitive data should ever touch the filesystem! This option is available here for local development when using KMS or whatnot to decrypt the keys is too much hassle!**

In non-local environment you can encrypt the stored vault keys using keys provided by publc cloud providers such as [AWS KMS](https://aws.amazon.com/kms/) or [CGP cloudkms](https://cloud.google.com/kms/). These are available via `-kms-provider` cli option. Please consult the command line options. At the moment only AWS KMS and GCP KMS are supported.

`vaultop` code can be extended to use your own KMS: PRs welcome :-) Equally, as long as the key storage is concerned, the code currently only supports local storage, however extending it for different storage options should be very easy.

## vaultops commands

Most of the `vaultops` commands do not require using complete manifest. To run some of the commands you only need some parts of the complete manifest. For example, say you only want to mount some backends. Given the `vault` is initialized and unsealed, you can specify the following `vaultops` manifest and run the `mount` command as shown below:

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

Similarly, you can run `policy` command with only the policies part of the full manifest. You can find some examples in [examples](examples) subdirectory of the `vaultops` project.

# TODO

**WE NEED WAY BIGGER TEST COVERAGE**

There is lots of redundant code so tonnes of refactoring up for grabs! 🤗

Only local key store is implemented at the moment, remote stores are up gor grabs! 🤗
