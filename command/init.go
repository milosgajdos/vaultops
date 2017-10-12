package command

import (
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

// InitCommand implements vault initialization
// It fulfills cli.Command interface
type InitCommand struct {
	// meta flags contain vault client config
	Meta
}

// Run runs init command which initializes vault server
// If init command fails it returns non-zero integer
func (c *InitCommand) Run(args []string) int {
	var status bool
	var threshold, shares int
	var config string
	// create command flags
	flags := c.Meta.FlagSet("init", FlagSetDefault)
	flags.Usage = func() { c.UI.Info(c.Help()) }
	flags.BoolVar(&status, "status", false, "")
	flags.IntVar(&shares, "key-shares", 5, "")
	flags.IntVar(&threshold, "key-threshold", 3, "")
	flags.StringVar(&config, "config", "", "")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	// get hosts against which we want to run init command
	hosts, err := c.getRunHosts(config)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to read vault hosts: %v", err))
		return 1
	}

	// if status is requested
	if status {
		return c.runInitStatus(hosts)
	}

	// init request options
	req := &api.InitRequest{
		SecretShares:      shares,
		SecretThreshold:   threshold,
		RecoveryShares:    shares,
		RecoveryThreshold: threshold,
	}

	c.UI.Info(fmt.Sprintf("Attempting to initialize vault:"))
	for _, host := range hosts {
		c.UI.Info(fmt.Sprintf("\t%s", host))
	}

	return c.runInit(req)
}

// runHosts retrieves a list of hosts agsints which the Init cmd should be run from configuration and returns it
func (c *InitCommand) getRunHosts(config string) ([]string, error) {
	if config != "" {
		hosts, err := getVaultHosts(config)
		if err != nil {
			return nil, err
		}
		return hosts, nil
	}

	// if no config is supplied read environment
	cfg, err := c.Config("")
	if err != nil {
		return nil, err
	}

	return []string{cfg.Address}, nil
}

// runInitStatus checks init status of vault server
func (c *InitCommand) runInitStatus(hosts []string) int {
	// more than 1 server requested
	type res struct {
		host   string
		status bool
		err    error
	}
	statChan := make(chan *res, 1)
	// check status of each host concurrently
	for _, host := range hosts {
		v, err := c.Client(host, "")
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to fetch Vault client: %v", err))
			return 1
		}
		go func(h string) {
			c.UI.Info(fmt.Sprintf("Reading init status for host: %s", h))
			status, err := v.Sys().InitStatus()
			statChan <- &res{host: h, status: status, err: err}
		}(host)
	}
	// collect the results
	var errStatus bool
	for i := 0; i < len(hosts); i++ {
		status := <-statChan
		if status.err != nil {
			c.UI.Error(fmt.Sprintf("Failed to read init status of: %s: %v", status.host, status.err))
			errStatus = true
			continue
		}
		c.UI.Info(fmt.Sprintf("Host: %s Initialized: %v", status.host, status.status))
	}
	if errStatus {
		return 1
	}

	return 0
}

// runInit initializes vault server and returns 0 if successful
func (c *InitCommand) runInit(req *api.InitRequest) int {
	// create client
	v, err := c.Client("", "")
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to fetch vault client: %v", err))
		return 1
	}

	// initialize vault server
	resp, err := v.Sys().Init(req)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to initialize %s: %v", v.Address(), err))
		return 1
	}

	c.UI.Info(fmt.Sprintf("Vault successfully initialized"))
	for i, key := range resp.Keys {
		c.UI.Info(fmt.Sprintf("Key %d: %s", i+1, key))
	}
	c.UI.Info(fmt.Sprintf("Root Token: %s", resp.RootToken))

	// write the retrieved vault keys into .local/vault.json
	vk := &VaultKeys{RootToken: resp.RootToken, MasterKeys: resp.Keys}
	if err := writeVaultKeys(localDir, localFile, vk); err != nil {
		c.UI.Error(fmt.Sprintf("Failed to store vault keys: %v", err))
		return 1
	}

	return 0
}

// Synopsis provides a simple command description
func (c *InitCommand) Synopsis() string {
	return "Initialize Vault cluster or server"
}

// Help returns detailed command help
func (c *InitCommand) Help() string {
	helpText := `
Usage: cam-vault init [options]

    Initialize a new Vault server or cluster.

    This command connects to a Vault server and initializes it for the
    first time. It sets up initial set of master keys and backend store.

    When init is called on already initialized server it will return error

General Options:
` + GeneralOptionsUsage() + `
init Options:

    -status 			Don't initialize the server, only check the init status
    -key-shares=5 		Number of key shares to split the master key into
    -key-threshold=3		Number of key shares required to reconstruct the master key
    -config			Path to a config file which contains a list of vault servers

`
	return strings.TrimSpace(helpText)
}
