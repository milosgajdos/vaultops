package command

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/vault/api"
)

// SetupCommand sets up vault cluster
// It fulfills cli.Command interface
type SetupCommand struct {
	// meta flags contain vault client config
	Meta
}

// Run runs setup command which performs at minimum init an dunseal actions
// If setup command fails it returns non-zero integer
func (c *SetupCommand) Run(args []string) int {
	var config string
	var keyFile string
	// create command flags
	flags := c.Meta.FlagSet("setup", FlagSetDefault)
	flags.Usage = func() { c.UI.Info(c.Help()) }
	flags.StringVar(&config, "config", "", "")
	flags.StringVar(&keyFile, "key-file", "", "")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	return c.runSetup(config, keyFile)
}

// runsetup setupializes vault server and returns 0 if successful
func (c *SetupCommand) runSetup(config, keyFile string) int {
	initCmd := &InitCommand{
		Meta: c.Meta,
	}
	// get hosts against which we want to run unseal command
	hosts, err := getVaultHosts(config, "init")
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to read vault init hosts: %v", err))
		return 1
	}
	// init request options
	req := &api.InitRequest{
		SecretShares:      5,
		SecretThreshold:   3,
		RecoveryShares:    5,
		RecoveryThreshold: 3,
	}
	// runt Vault Init
	c.UI.Info(fmt.Sprintf("Attempting to initialize vault cluster"))
	if ret := initCmd.runInit(hosts, req); ret != 0 {
		return ret
	}

	unsealCmd := &UnsealCommand{
		Meta: c.Meta,
	}
	// get hosts against which we want to run unseal command
	hosts, err = getVaultHosts(config, "unseal")
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to read vault unseal hosts: %v", err))
		return 1
	}
	// if keyFile is not provided try to read the loca keys
	if keyFile == "" {
		keyFile = filepath.Join(localDir, localFile)
	}
	// get keys
	keys, err := readVaultKeys(keyFile)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to parse  %s key file: %v", keyFile, err))
		return 1
	}
	// unsearl vault cluster
	c.UI.Info(fmt.Sprintf("Attempting to unseal vault cluster"))
	if ret := unsealCmd.runUnseal(hosts, keys); ret != 0 {
		return ret
	}

	mountCmd := &MountCommand{
		Meta: c.Meta,
	}
	// get mounts
	mounts, err := getVaultMounts(config)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to read vault mounts: %v", err))
		return 1
	}
	// run Vault Mount
	c.UI.Info(fmt.Sprintf("Attempting to mount vault backends"))
	if ret := mountCmd.runMount(mounts); ret != 0 {
		return ret
	}

	beCmd := &BackendCommand{
		Meta: c.Meta,
	}
	// get backends
	backends, err := getVaultBackends(config)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to read vault backends: %v", err))
		return 1
	}
	// run vault backend setup
	c.UI.Info(fmt.Sprintf("Attempting to configure vault backends"))
	if ret := beCmd.runBackend(backends); ret != 0 {
		return ret
	}

	policyCmd := &PolicyCommand{
		Meta: c.Meta,
	}
	// get policies
	policies, err := getVaultPolicies(config)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to read vault policies: %v", err))
		return 1
	}
	// run vault policy setup:
	c.UI.Info(fmt.Sprintf("Attempting to configure vault policies"))
	if ret := policyCmd.runPolicy(policies); ret != 0 {
		return ret
	}

	return 0
}

// Synopsis provides a simple command description
func (c *SetupCommand) Synopsis() string {
	return "Setup a new Vault server"
}

// Help returns detailed command help
func (c *SetupCommand) Help() string {
	helpText := `
Usage: cam-vault setup [options]

    setup sets up vault cluster as per configuration

    At minimum, set up will attempt to initialize and unseal the vault cluster.
    Additionally it will perform further setup actions specified in the configuration file

    If at any point setup fails with error it will return non-zero error code

General Options:
` + GeneralOptionsUsage() + `
setup Options:

    -config			Path to a config file which contains a list of vault servers and setup actions

`
	return strings.TrimSpace(helpText)
}
