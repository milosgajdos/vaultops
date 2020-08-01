# vaultops

[![GoDoc](https://godoc.org/github.com/milosgajdos/vaultops?status.svg)](https://godoc.org/github.com/milosgajdos/vaultops)
[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Travis CI](https://travis-ci.org/milosgajdos/vaultops.svg?branch=master)](https://travis-ci.org/milosgajdos/vaultops)
[![Go Report Card](https://goreportcard.com/badge/milosgajdos/vaultops)](https://goreportcard.com/report/github.com/milosgajdos/vaultops)

`vaultops` is a command line utility which aims to simplify [vault](https://www.vaultproject.io/) server setup. At the moment it support automatic initialization and unsealing.

# Motivation

Typical `vault` setup usually requires taking several steps before the server can be used:

- initializing `vault` server
- unsealing `vault` server(s)
- mounting `vault` backends
- creating `vault` backend roles
- configuring `vault` policies

The above listed tasks are usually performed using the `vault` command line client which interacts with `vault` servers. Using the `vault` command line client requires writing a lot of `shell` scripts which often grow into unmanageable full fledged monsters which are often hard to debug and maintain.

`vaultops` attempts to address this problem by providing a simple manifest file which can be used to specify all the tasks required to perform `vault` setup once the `vault` server/cluster is running. `vaultops` reads the manifest file and performs all the actions requsted by user.

# Quick start

Get the project:

```console
$ go get -u github.com/milosgajdos/vaultops
```

Run the tests
```console
$ cd $GOPATH/src/github.com/milosgajdos/vaultops
$ make test
for pkg in github.com/milosgajdos/vaultops github.com/milosgajdos/vaultops/cipher github.com/milosgajdos/vaultops/cloud/aws github.com/milosgajdos/vaultops/cloud/gcp github.com/milosgajdos/vaultops/command github.com/milosgajdos/vaultops/manifest github.com/milosgajdos/vaultops/store github.com/milosgajdos/vaultops/store/local; do \
		go test -coverprofile="../../../$pkg/coverage.txt" -covermode=atomic $pkg || exit; \
	done
?   	github.com/milosgajdos/vaultops	[no test files]
?   	github.com/milosgajdos/vaultops/cipher	[no test files]
ok  	github.com/milosgajdos/vaultops/cloud/aws	0.022s	coverage: 93.5% of statements
ok  	github.com/milosgajdos/vaultops/cloud/gcp	0.024s	coverage: 28.1% of statements
ok  	github.com/milosgajdos/vaultops/command	1.478s	coverage: 15.4% of statements
ok  	github.com/milosgajdos/vaultops/manifest	0.022s	coverage: 100.0% of statements
ok  	github.com/milosgajdos/vaultops/store	0.020s	coverage: 100.0% of statements
ok  	github.com/milosgajdos/vaultops/store/local	0.020s	coverage: 88.9% of statements
```

Build the binary:
```console
$ make build
mkdir -p ./_build
go build -ldflags="-s -w" -o "./_build/vaultops"
```

Once you have pulled in all the project dependencies you can also build the binary running the familiar command:
```console
$ go build -ldflags="-s -w"
```

## Usage

`vaultops` provides various commands following the same UX as `vault` command line utility. You can see the currently available commands below:

```
Usage: vaultops [--version] [--help] <command> [<args>]

Available commands are:
    init      Initialize Vault cluster or server
    unseal    Unseal a Vault server
```

`vaultops` reads **the same environment variables** as `vault` utility, so you can rely on the familiar `$VAULT_` environment variables when specifying the `vault` server URLs and tokens.

At the moment only `init` and `unseal` commands are implemented. The plan is to add a few more.

## vaultops init

`vaultops init` initializes the vault server. Besides providing the familiar `vault` command line utility options to connect to the `vault` server, it adds a few flags which allow to encrypt the `vault` master keys and root token and store them either on the workstation filesystem or encrypted in a remote storage:

```console
./vaultops init -help
Usage: vaultops init [options]

    Initialize a new Vault server or cluster.

    This command connects to a Vault server and initializes it for the first time.
    It sets up initial set of master keys and secret store.
    Unless overridden init stores vault root token and keys on the local filesystem.

    When init is called on already initialized server it will return error.
...
...
...

  -redact=true 		  Redacts sensitive information when printing into stdout
  -kms-provider 	  KMS provider (aws, gcp)
  -aws-kms-id		  AWS KMS ID. KMS keys with given ID will be used to encrypt vault keys
  -gcp-kms-crypto-key	  GCP KMS crypto key ID
  -gcp-kms-key-ring       GCP KMS keyring
  -gcp-kms-region     	  GCP region (eg. 'global', 'europe-west1')
  -gcp-kms-project  	  GCP project name
  -storage-bucket         Cloud storage bucket
  -storage-key            Cloud storage key
  -key-store=local	  Type of store where to loook up vault keys (default: local)
    			  Local store is ./.local/vault.json
  -key-local-path         Path to locally stored keys

init Options:

  -status 			Don't initialize the server, only check the init status
  -key-shares=5 		Number of key shares to split the master key into
  -key-threshold=3		Number of key shares required to reconstruct the master key
  -config			Path to a config file which contains a list of vault servers
```

When run with the default options, `init` command will store the `vault` keys **UNENCRYPTED** on your local filesystem in `.local` directory of your **current working directory** in a predefied `json` format which looks as follows:

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

**This is not what you should do when initialising your vault servers! This option is provided as convenience for local development!**

For real life setup, `vaultops` allows to encrypt the `vault` keys using the encryption keys provided by publc cloud providers such as [AWS KMS](https://aws.amazon.com/kms/) or [CGP Cloud KMS](https://cloud.google.com/security-key-management). These options are available via `-kms-provider` flag. Here is an example how to use the GCP Cloud KMS to encrypt the vault keys when initialising a new vault server:

```console
$ # export VAULT_ADDR environment variable
$ export VAULT_ADDR="http://10.100.21.161:8200"
$ # initialize vault server and encrypt the vault keys
$ ./vaultops init -kms-provider="gcp" \
 		  -gcp-kms-project="kube-blog" \
		  -gcp-kms-region="europe-west1" \
		  -gcp-kms-key-ring="vaultops" \
		  -gcp-kms-crypto-key="vaultops"

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

Running the above `init` command will store the `vault` master keys and initial root token locally in `./.local/vault.json` file but now they're encrypted using the GCP Cloud KMS keys so if you try to inspect the local file you'll get a "garbage" pile of randomly generated bytes. If you want to see the actual unencrypted keys you need to decrypt the file using the same KMS keys you used when encrypting them.

**At the moment only AWS KMS and GCP KMS are supported**

#### Vault Keys redacting

By default `vaultops` tries to "redact" all sensitive information printed to `stdout`. You can disable this behavior via `-redact` command line switch by setting it to `false`. This will print the master keys and initial root token into `stdout` in plaintext.

### Vault Key storage

`vaultops` allows you to store `vault` keys remotely either in [AWS S3](https://aws.amazon.com/s3/), [Google Cloud Storage](https://cloud.google.com/storage/) or [kubernetes secrets](https://kubernetes.io/docs/concepts/configuration/secret/). You can choose the appropriate remote storage option via `-key-store` flag. Here is an example how to initialize `vault` using AWS KMS and store the keys in AWS S3 bucket of your choice:

```console
$ export VAULT_ADDR="http://${HostIP}:8200"
$ ./vaultops init -key-store="s3" \
		  -storage-bucket="vaultops-kms" \
		  -storage-key="vault.json" \
		  -kms-provider="aws" \
		  -aws-kms-id="your-kms-id"
```

**NOTE:** when using kubernetes secrets storage, you can also specify a namespace for the secret; the default value is set to `default` namespace

## vaultops unseal

`vaultops unseal` unseals the vault cluster using the keys generated by `vault` during its initalisation. These keys can be stored encrypted or in plaintext either locally or remotely based on the command line switches you picked when you initialized the server. `unseal` command allows you to read these keys from whichever locatioon you stored them in during initialization and use them to unseal the `vault` server. See the available command line options listed below:

```console
$ ./vaultops unseal -help
Usage: vaultops unseal [options]

    Unseal the vault serve by entering master keys.

    This command connects to a Vault server and attempts to unseal it.
    first time. It sets up initial set of master keys and backend store.

    When init is called on already initialized server it will error

...
...
...

  -redact=true 		  Redacts sensitive information when printing into stdout
  -kms-provider 	  KMS provider (aws, gcp)
  -aws-kms-id		  AWS KMS ID. KMS keys with given ID will be used to encrypt vault keys
  -gcp-kms-crypto-key	  GCP KMS crypto key id
  -gcp-kms-key-ring       GCP KMS key ring
  -gcp-kms-region     	  GCP region (eg. 'global', 'europe-west1')
  -gcp-kms-project  	  GCP project name
  -storage-bucket         Cloud storage bucket
  -storage-key            Cloud storage key
  -key-store=local	  Type of store where to loook up vault keys (default: local)
    			  Local store is ./.local/vault.json
  -key-local-path         Path to locally stored keys

unseal Options:

    -status 		  Don't unseal the server, only check the seal status
    -config		  Path to a config file which contains a list of vault servers
```

Here is an example of how to unseal the vault server using the keys stored in AWS S3 which were encrypted using AWS KMS:

```console
$ export AWS_REGION="us-east-1"
$ export VAULT_ADDR="http://${HostIP}:8200"
$ # unseal the vault server
$ ./vaultops unseal -key-store="s3" \
		    -storage-bucket="vaultops-kms" \
		    -storage-key="vault.json" \
		    -kms-provider="aws" \
		    -aws-kms-id="your-kms-id"
```

Obviously, you can create all kinds of crazy combination of storages and encryption keys i.e. store the keys in AWS S3, but encrypt them using GCP Cloud KMS

# Manifest

`vaultops` allows you to create a manifest file which can be used when running `vaultops` commands. The manifest is a simple `YAML` (woo, hoo! more `YAML` ᕕ( ᐛ )ᕗ) file which specifies a list of `vault` hosts for initialization and unsealing.

Let's look at a simple example:

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

Hopefully the `YAML` snippet above is straightforward to understand: we've got a server to initialize and a cluster of servers to unseal.

You don't have to have a full manifest at hand if you want to run just one command. Say, you just want to initialize the server -- you can supply the following `YAML` snippet to `init` command via `-config` command line switch:

```yaml
hosts:
  # URL of vault server to use for initialization
  init:
    - "http://10.100.21.161:8200"
```

Run the `init` using the manifest above stored in some file on the local filesystem:

```console
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

Same rules apply when running `unseal` command.

# TODO

* bigger test coverage
* plenty of room for refactoring
* setting up secret backends
* setting up secret backend roles
* setting up policies
