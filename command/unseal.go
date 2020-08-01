package command

import (
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/milosgajdos/vaultops/cipher"
	"github.com/milosgajdos/vaultops/manifest"
)

// UnsealCommand implements vault unsealing
// It fulfills cli.Command interface
type UnsealCommand struct {
	// meta flags contain vault client config
	Meta
}

// Run runs unsearl command which unseals vault servers
// If unseal fails Run returns non-zero integer
func (c *UnsealCommand) Run(args []string) int {
	var status bool
	var config string
	// create command flags
	flags := c.Meta.FlagSet("unseal", FlagSetDefault)
	flags.Usage = func() { c.UI.Error(c.Help()) }
	flags.BoolVar(&status, "status", false, "")
	flags.StringVar(&config, "config", "", "")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	// get hosts against which we want to run unseal command
	hosts, err := c.getRunHosts(config)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to read vault hosts: %v", err))
		return 1
	}

	if status {
		return c.runSealStatus(hosts)
	}
	// create vault keys store handle
	s, err := VaultKeyStore(c.flagKeyStore, &c.Meta)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to create %s store: %v", c.flagKeyStore, err))
		return 1
	}
	// if kms provider not empty, initialize cipher
	var cphr cipher.Cipher
	if c.flagKMSProvider != "" {
		cphr, err = VaultKeyCipher(&c.Meta)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to create %s cipher: %v", c.flagKMSProvider, err))
			return 1
		}
	}
	// read vault keys
	vk := new(VaultKeys)
	if _, err := vk.Read(s, cphr); err != nil {
		c.UI.Error(fmt.Sprintf("Failed to read vault keys: %v", err))
		return 1
	}

	c.UI.Info(fmt.Sprintf("Attempting to unseal vault cluster:"))
	for _, host := range hosts {
		c.UI.Info(fmt.Sprintf("\t%s", host))
	}

	return c.runUnseal(hosts, vk)
}

// runHosts retrieves a list of hosts agsints which the Init cmd should be run from configuration and returns it
func (c *UnsealCommand) getRunHosts(config string) ([]string, error) {
	if config != "" {
		m, err := manifest.Parse(config)
		if err != nil {
			return nil, err
		}

		hosts, err := m.GetHosts("init")
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

// runUnsealStatus checks unseal status of vault server
func (c *UnsealCommand) runSealStatus(hosts []string) int {
	type res struct {
		host string
		resp *api.SealStatusResponse
		err  error
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
			// check status and send down the status channel
			c.UI.Info(fmt.Sprintf("Reading seal status of host: %s", h))
			resp, err := v.Sys().SealStatus()
			statChan <- &res{host: h, resp: resp, err: err}
		}(host)
	}
	// collect the results
	var errStatus bool
	for i := 0; i < len(hosts); i++ {
		status := <-statChan
		if status.err != nil {
			c.UI.Error(fmt.Sprintf("Failed to read seal status of %s: %v", status.host, status.err))
			errStatus = true
			continue
		}
		c.UI.Info(fmt.Sprintf("Host: %s Sealed: %v", status.host, status.resp.Sealed))
	}
	if errStatus {
		return 1
	}

	return 0
}

// runUnseal attempts to unseal vault hosts using the keys
// unseal action requires vault root token to be supplied via keys as well as unseal keys
func (c *UnsealCommand) runUnseal(hosts []string, vk *VaultKeys) int {
	if vk.RootToken == "" || vk.MasterKeys == nil {
		c.UI.Error(fmt.Sprintf("No vault keys providedd"))
		return 1
	}

	type res struct {
		host string
		resp *api.SealStatusResponse
		err  error
	}
	statChan := make(chan *res, 1)

	// check status of each host concurrently
	for _, host := range hosts {
		v, err := c.Client(host, vk.RootToken)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to fetch Vault client: %v", err))
			return 1
		}

		go func(h string) {
			// check status and send down the status channel
			resp, err := v.Sys().SealStatus()
			if err != nil {
				c.UI.Error(fmt.Sprintf("Failed to read seal status of %s: %v", h, err))
				statChan <- &res{host: h, resp: resp, err: err}
				return
			}
			// if the host is unsealed, don't do anything
			if !resp.Sealed {
				statChan <- &res{host: h, resp: resp, err: err}
				return
			}
			// if unseal threshold is bigger than the number of supplied master keys
			// we only attempt len(keys) unseals
			t := resp.T
			if t > len(vk.MasterKeys) {
				t = len(vk.MasterKeys)
			}
			// attempt to unseal vault nodes with all the keys; bail on error
			// otherwise we return the latest unseal response
			c.UI.Info(fmt.Sprintf("Attempting to unseal host: %s", h))
			for i := 0; i < t; i++ {
				resp, err = v.Sys().Unseal(vk.MasterKeys[i])
				if err != nil {
					break
				}
			}
			statChan <- &res{host: h, resp: resp, err: err}
		}(host)
	}
	// collect the results
	var errStatus bool
	for i := 0; i < len(hosts); i++ {
		status := <-statChan
		if status.err != nil {
			c.UI.Error(fmt.Sprintf("Failed to unseal %s: %v", status.host, status.err))
			errStatus = true
			continue
		}
		c.UI.Info(fmt.Sprintf(
			"Host %s: \n"+
				"\tSealed: %v\n"+
				"\tKey Shares: %d\n"+
				"\tKey Threshold: %d\n"+
				"\tUnseal Progress: %d\n"+
				"\tUnseal Nonce: %v",
			status.host,
			status.resp.Sealed,
			status.resp.N,
			status.resp.T,
			status.resp.Progress,
			status.resp.Nonce,
		))
	}
	// if at least one error encounctered we return non-zero
	if errStatus {
		return 1
	}

	c.UI.Info(fmt.Sprintf("Vault successfully unsealed"))

	return 0
}

// Synopsis provides a simple command description
func (c *UnsealCommand) Synopsis() string {
	return "Unseal a Vault server"
}

// Help returns detailed command help
func (c *UnsealCommand) Help() string {
	helpText := `
Usage: vaultops unseal [options]

    Unseal the vault serve by entering master keys.

    This command connects to a Vault server and attempts to unseal it.
    first time. It sets up initial set of master keys and backend store.

    When init is called on already initialized server it will error

General Options:
` + GeneralOptionsUsage() + `
unseal Options:

    -status 			Don't unseal the server, only check the seal status
    -config			Path to a config file which contains a list of vault servers
`
	return strings.TrimSpace(helpText)
}
