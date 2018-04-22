package command

import (
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/milosgajdos83/vaultops/manifest"
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

	var backends manifest.Backends
	if config != "" {
		m, err := manifest.Parse(config)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to parse config %s: %s", config, err))
			return 1
		}
		// get mounts
		backends = m.GetBackends()
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

func (c *BackendCommand) writeRoles(v *api.Client, backend string, roles manifest.Roles) error {
	for _, role := range roles {
		c.UI.Info(fmt.Sprintf("Attempting to create role: %s", role.Name))
		// TODO: traverse cert yaml annotations to populate this config map
		config := map[string]interface{}{
			"allowed_domains":    role.AllowedDomains,
			"allow_bare_domains": role.AllowBareDomains,
			"allow_any_name":     role.AllowAnyName,
			"enforce_hostnames":  role.EnforceHostnames,
			"organization":       role.Organization,
		}
		if _, err := v.Logical().Write(backend+"/roles/"+role.Name, config); err != nil {
			return err
		}
	}

	return nil
}

func (c *BackendCommand) writeCerts(v *api.Client, backend string, certs manifest.Certs) error {
	var path string
	for _, cert := range certs {
		switch cert.Action {
		case "generate":
			path = fmt.Sprintf("%s/%s/generate/%s", backend, cert.Kind, cert.Type)
		case "issue":
			path = fmt.Sprintf("%s/issue/%s", backend, cert.Role)
		default:
			return fmt.Errorf("Invalid certificate action: %s", cert.Action)
		}
		// issue certificates
		c.UI.Info(fmt.Sprintf("Attempting to %s %s certificate", cert.Action, cert.Name))
		// TODO: traverse cert yaml annotations to populate this config map
		config := map[string]interface{}{
			"common_name": cert.CommonName,
			"ttl":         cert.TTL,
			"ip_sans":     cert.IPSans,
			"alt_names":   cert.AltNames,
		}

		resp, err := v.Logical().Write(path, config)
		if err != nil {
			return err
		}
		c.UI.Info(fmt.Sprintf("%v", resp.Data["certificate"].(string)))
		if cert.Action == "issue" {
			c.UI.Info(fmt.Sprintf("%v", resp.Data["private_key"].(string)))
		}
	}

	return nil
}

// runBackend creates vault backends
func (c *BackendCommand) runBackend(backends manifest.Backends) int {
	// more than 1 server requested
	for _, backend := range backends {
		v, err := c.Client("", "")
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to fetch Vault client: %v", err))
			return 1
		}
		// create PKI roles
		if err := c.writeRoles(v, backend.Name, backend.Roles); err != nil {
			c.UI.Error(fmt.Sprintf("Failed to create role: %s", err))
		}
		// generate SSL certs
		if err := c.writeCerts(v, backend.Name, backend.Certs); err != nil {
			c.UI.Error(fmt.Sprintf("Failed to create certificate: %s", err))
		}
	}
	// collect the results
	c.UI.Info(fmt.Sprintf("Finished configuring backends"))

	return 0
}

// Synopsis provides a simple command description
func (c *BackendCommand) Synopsis() string {
	return "Manage vault backends"
}

// Help returns detailed command help
func (c *BackendCommand) Help() string {
	helpText := `
Usage: vaultops backend [options]

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
