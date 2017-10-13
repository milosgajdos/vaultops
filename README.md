**[NOTE] This utility is a PoC! Use it at your own risk!**

# vaultops

`vaultops` is a simple command line utility which aims to simplify complex [vault](https://www.vaultproject.io/) server setups. The tool provides a few subcommands which allows the user to performa various setup tasks (creating vault policies, mounting backends etc.) using a single command.

# Motivation

`vault` setup typically requires a lot of manual tasks:
- initializing vault server
- unsealing vault server or multiple servers in the same clsuter
- mounting backends of different types
- creating backend roles
- generating SSL certificates
- configuring policies

All of these tasks are usually performed via the `vault` cli utility which requires writing a lot of `shell` scripts that often grow into unmanageable full fledged monster `shell` programs which have no type checking or proper error handling.

`vaultops` utility addresses this problem by providing a simple manifest file which can be used to specify all the tasks required to perform a full `vault` setup once the `vault` is running. `vaultops` reads the manifest file and performs all the actions requsted by user. `vaultops` interacts directly with `vault` API and performs the requested tasks concurrently whenever possible. The user can choose to perform a full set up or only particular setup actions.`vaultops` provides a few subcommands.

# Quick start

Install the project and run the tests:
```
$ go get github.com/milosgajdos83/vaultops
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

`vaultop` reads the same environment as `vault` utility so you can rely on the familiar `$VAULT_` environment variables when spcifying your `vault` server URL and authentication token.

# Manifest

`vaultops` allows you to define a manifest file which can be supplied to particular subcommand. The manifest is a simple `YAML` file which specifies a list of `vault` hosts or `vault` resources that are requested to be created in `vault`. A skeleton of a sample manifest file can be seen below:

```yaml
# vault hosts
hosts:
  unseal:
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
# vault policies
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

**READ CAREFULLY**

As you may have noticed, when running the tool, `vaultops` initializes the `vault` and the prints the `vault` keys into . `standard output`. Furthermore, purely for convenience `init` command stores the `vault` keys in `.local` directory of your corrent working directory in a predefined `json` file:

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

**This is not what you should do in real life!. No secrets should touch your filesystem!**. In real life you would store the `vault` keys in secure location. The reason why the keys are stored locally is because currently they are read by `unseal` subcommand if no other key file is specified via its `-key-file` switch. `vaultops` attempts to unseal every single `vault` server specified in the manifest using the keys specified via the earlier mentioned cli flag.

**Storing the keys on the local filesystem will be optional in the future release!**

## vaultops commands

Most of the `vaultops` commands do not require using full manifest. User can simply pick only some extracted parts and run the tool just against those. For example, say you only want to mount some backends. Given the `vault` is initialized and unsealed, you can specify the following `vaultops` manifest and run the `mount` command as below:

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

Similarly, you run `policy` command with only the policies manifest. You can find some examples in [examples]() subdirectory of the `vaultops` project source tree.

# TODO

**WE NEED WAY BIGGER TEST COVERAGE**

There is lots of redundant code so tonnes of refactoring up for grabs!
