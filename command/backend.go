package command

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

// BackendCommand allows to configure vault backends
// It satisfies cli.Command interface
type BackendCommand struct {
	// meta flags contain vault client config
	Meta
}

// Run runs backedn command which configures vault backends
// If backend command fails it returns non-zero integer
func (c *BackendCommand) Run(args []string) int {
	var list string
	var config string
	// create command flags
	flags := c.Meta.FlagSet("backend", FlagSetDefault)
	flags.Usage = func() { c.UI.Info(c.Help()) }
	flags.StringVar(&list, "list", "", "")
	flags.StringVar(&config, "config", "", "")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	// if status is requested
	if list != "" {
		return c.runBackendList(list)
	}

	backends, err := getVaultBackends(config)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to read vault backends: %v", err))
		return 1
	}

	c.UI.Info(fmt.Sprintf("Attempting to configure vault backends:"))
	for _, backend := range backends {
		c.UI.Info(fmt.Sprintf("\t%s", backend.Name))
	}

	return c.runBackend(backends)
}

// runBackendList list vault backends
func (c *BackendCommand) runBackendList(path string) int {
	client, err := c.Client("", "")
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to fetch Vault client: %v", err))
		return 1
	}

	secret, err := client.Logical().List(path)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to read data in %s: %s", path, err))
		return 1
	}

	if secret != nil {
		c.UI.Info(fmt.Sprintf("%v", secret))
	} else {
		c.UI.Info(fmt.Sprintf("No data found in path: %s", path))
	}

	return 0
}

func (c *BackendCommand) writeRoles(v *api.Client, roles []*VaultRole) error {
	for _, role := range roles {
		c.UI.Info(fmt.Sprintf("Attempting to create role: %s", role.Name))
		if _, err := v.Logical().Write(role.Backend+"/roles/"+role.Name, role.Config); err != nil {
			return err
		}
	}

	return nil
}

func (c *BackendCommand) writeCerts(v *api.Client, certs []*VaultCert) error {
	var path string
	var certInfo string
	for _, cert := range certs {
		if cert.Root {
			path = cert.Backend + "/root/generate/" + cert.Type
			certInfo = fmt.Sprintf("%s root SSL certificate", cert.Type)
		} else {
			path = cert.Backend + "/issue/" + cert.Role
			certInfo = fmt.Sprintf("SSL certificate with role %s", cert.Role)
		}
		// issue certificates
		c.UI.Info(fmt.Sprintf("Attempting to generate %s for backend: %s", certInfo, cert.Backend))
		c, err := v.Logical().Write(path, cert.Config)
		if err != nil {
			return err
		}

		if cert.Store {
			s, err := json.Marshal(c.Data)
			if err != nil {
				return err
			}
			// unmarshal into interface
			var params interface{}
			err = json.Unmarshal(s, &params)
			if err != nil {
				return err
			}
			path := "secret/" + cert.Backend + "/" + cert.Role
			if _, err := v.Logical().Write(path, params.(map[string]interface{})); err != nil {
				return err
			}
		}
	}

	return nil
}

// runBackend creates vault backends
func (c *BackendCommand) runBackend(backends []*VaultBackend) int {
	// more than 1 server requested
	beChan := make(chan error, 1)
	for _, backend := range backends {
		v, err := c.Client("", "")
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to fetch Vault client: %v", err))
			return 1
		}
		go func(b *VaultBackend) {
			// create PKI roles
			if err := c.writeRoles(v, b.Roles); err != nil {
				beChan <- err
				return
			}
			// generate SSL certs
			if err := c.writeCerts(v, b.Certs); err != nil {
				beChan <- err
				return
			}
			beChan <- nil
		}(backend)
	}
	// collect the results
	var errStatus bool
	for i := 0; i < len(backends); i++ {
		err := <-beChan
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to configure backend %s : %v", backends[i].Name, err))
			errStatus = true
			continue
		}
		c.UI.Info(fmt.Sprintf("Successfully configured backend %s", backends[i].Name))
	}
	if errStatus {
		return 1
	}

	c.UI.Info(fmt.Sprintf("All requested vault backends successfully configured"))

	return 0
}

// Synopsis provides a simple command description
func (c *BackendCommand) Synopsis() string {
	return "Manage vault backends"
}

// Help returns detailed command help
func (c *BackendCommand) Help() string {
	helpText := `
Usage: cam-vault backend [options]

    Manages vault backends

    This command connects to a Vault server and configures vault backends

    When backend is called on already existing backend it will modify it

General Options:
` + GeneralOptionsUsage() + `
init Options:

    -list 			List vault backend path
    -config			Path to a config file which contains a list of vault backends to manage

`
	return strings.TrimSpace(helpText)
}
