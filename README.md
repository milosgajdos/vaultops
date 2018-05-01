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

The above listed tasks are usually performed using the `vault` cli tool which interacts with `vault` API. Using `vault` cli tool requires writing a lot of `shell` scripts which often grow into unmanageable full fledged monsters which are often hard to debug and maintain.

`vaultops` attempts to address this problem by providing a simple manifest file which can be used to specify all the tasks required to perform `vault` setup once the `vault` server/cluster is running. `vaultops` reads the manifest file and performs all the actions requsted by user by interacting `vault` REST API.

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
for pkg in github.com/milosgajdos83/vaultops github.com/milosgajdos83/vaultops/cipher github.com/milosgajdos83/vaultops/cloud/aws github.com/milosgajdos83/vaultops/cloud/gcp github.com/milosgajdos83/vaultops/command github.com/milosgajdos83/vaultops/manifest github.com/milosgajdos83/vaultops/store github.com/milosgajdos83/vaultops/store/local; do \
		go test -coverprofile="../../../$pkg/coverage.txt" -covermode=atomic $pkg || exit; \
	done
?   	github.com/milosgajdos83/vaultops	[no test files]
?   	github.com/milosgajdos83/vaultops/cipher	[no test files]
ok  	github.com/milosgajdos83/vaultops/cloud/aws	0.022s	coverage: 93.5% of statements
ok  	github.com/milosgajdos83/vaultops/cloud/gcp	0.024s	coverage: 28.1% of statements
ok  	github.com/milosgajdos83/vaultops/command	1.478s	coverage: 15.4% of statements
ok  	github.com/milosgajdos83/vaultops/manifest	0.022s	coverage: 100.0% of statements
ok  	github.com/milosgajdos83/vaultops/store	0.020s	coverage: 100.0% of statements
ok  	github.com/milosgajdos83/vaultops/store/local	0.020s	coverage: 88.9% of statements
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

`vaultops` provides various subcommands following the same UX as `vault` cli utility. You can see the available commands below:

```
Usage: vaultops [--version] [--help] <command> [<args>]

Available commands are:
    init      Initialize Vault cluster or server
    unseal    Unseal a Vault server
```

`vaultops` reads **the same environment variables** as `vault` cli tool, so you can rely on the familiar `$VAULT_` environment variables when specifying the `vault` server URLs and tokens.

At the moment only `init` and `unseal` commands are implemented. The plan is to add many more like `secrets` etc.

# Manifest

`vaultops` allows you to create a manifest file which can be used when running `vaultops` commands. The manifest is a simple `YAML` file which specifies a list of `vault` hosts and `vault` resources that are requested to be setup in `vault`. A sample manifest file can be seen in [example](manifest/examples/example.yaml).

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
```

Hopefully the `YAML` snippet above is straightforward to understand :-)

## vaultops commands

`vaultops` commands do not require the complete setup manifest. Instead you can extract only a part of manifest needed by particular command you are trying to run.

For example, say you only want to initialis `vault`. You can extract the following manifest and supply it to `init` command via `-config` command line switch:

```yaml
hosts:
  # URL of vault server to use for initialization
  init:
    - "http://10.100.21.161:8200"
```

Run `init` command:

```
$ VAULT_ADDR="http://10.100.21.161:8200" ./vaultops init -config init.yaml
[INFO] Attempting to initialize vault:
[INFO] 	http://10.100.21.161:8200
[INFO] Host: http://10.100.21.161:8200 initialized. Master keys:
[INFO] Key 1: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
[INFO] Key 2: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
[INFO] Key 3: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
[INFO] Key 4: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
[INFO] Key 5: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
[INFO] Initial Root Token: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

You can find more examples in [examples](examples) subdirectory of the `vaultops` project.

### Security

When run with default options, `init` command, will store `vault` keys **UNENCRYPTED** on your local filesystem in `.local` directory of your **current working directory** in a predefied `json` format which looks as follows:

```json
{
  "root_token": "your-root-token",
  "master_keys": [
    "master-key-1",
    "master-key-2",
    "master-key-3",
    "master-key-4",
    "master-key-5"
  ]
}
```

**This is not what you should do when unsealing your vault clusters unless you are unsealing a cluster for development pruposes! No sensitive data should ever touch the filesystem unencrypted! This option is available here for a convenience when devloping locally!**

In non-local environment `vaultop` allows to encrypt the `vault` keys using encryption keys provided by publc cloud providers such as [AWS KMS](https://aws.amazon.com/kms/) or [CGP cloudkms](https://cloud.google.com/kms/). These options are available via `-kms-provider` cli switch. Here is an example how to use GCP cloudkms to encrypt the vault keys:

```
$ VAULT_ADDR="http://10.100.21.161:8200" ./vaultops init -kms-provider="gcp" -gcp-kms-project="kube-blog" -gcp-kms-region="europe-west1" -gcp-kms-key-ring="vaultops" -gcp-kms-crypto-key="vaultops"
[INFO] Attempting to initialize vault:
[INFO] 	http://10.100.21.161:8200
[INFO] Host: http://10.100.21.161:8200 initialized. Master keys:
[INFO] Key 1: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
[INFO] Key 2: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
[INFO] Key 3: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
[INFO] Key 4: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
[INFO] Key 5: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
[INFO] Initial Root Token: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

Running the above will store the `vault` keys locally in `./.local/vault.json` file but now they're encrypted using the KMS keys so if you try to inspect them you'll get a "garbage" pile of bytes. If you want to see the actual unencrypted keys you need to decrypt them using the same KMS keys you used when encrypting them.

**At the moment only AWS KMS and GCP KMS are supported**.

`vaultops` code can be extended to use your own KMS: PRs welcome :-)

#### Vault Keys redacting

By default `vaultops` redacts all sensitive information printed in `stdout`. You can disable this behavior via `-redact` command line switch.

### Vault Key storage

`vaultops` allows you to store `vault` keys remotely either in [AWS S3](https://aws.amazon.com/s3/) or [Google Cloud Storage](https://cloud.google.com/storage/). You can choose the storage option via `-key-store` command line switch. Here is an example how to initialize `vault` using AWS KMS and store the keys in AWS S3 bucket of your choice:

```
$ VAULT_ADDR="http://${HostIP}:8200" ./vaultops init -key-store="s3" -storage-bucket="vaultops-kms" -storage-key="vault.json" -kms-provider="aws" -aws-kms-id="your-kms-id"
```

# TODO

**WE NEED WAY BIGGER TEST COVERAGE**

There is lots of redundant code so tonnes of refactoring up for grabs! 🤗
