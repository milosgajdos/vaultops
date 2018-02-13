package command

import (
	"fmt"
	"strings"
)

// PolicyCommand allows to configure vault policies
// It satisfies cli.Command interface
type PolicyCommand struct {
	// meta flags contain vault client config
	Meta
}

// Run runs backedn command which configures vault policies
// If policy command fails it returns non-zero integer
func (c *PolicyCommand) Run(args []string) int {
	var list string
	var config string
	// create command flags
	flags := c.Meta.FlagSet("policy", FlagSetDefault)
	flags.Usage = func() { c.UI.Info(c.Help()) }
	flags.StringVar(&list, "list", "", "")
	flags.StringVar(&config, "config", "", "")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	// if status is requested
	if list != "" {
		return c.runPolicyList(list)
	}

	policies, err := getVaultPolicies(config)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to read vault policies: %v", err))
		return 1
	}

	c.UI.Info(fmt.Sprintf("Attempting to configure vault policies:"))
	for _, policy := range policies {
		c.UI.Info(fmt.Sprintf("\t%s", policy.Name))
	}

	return c.runPolicy(policies)
}

// runPolicyList list vault policy rules for given policy name
func (c *PolicyCommand) runPolicyList(policy string) int {
	client, err := c.Client("", "")
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to fetch Vault client: %v", err))
		return 1
	}

	switch {
	case strings.EqualFold(policy, "all"):
		policies, err := client.Sys().ListPolicies()
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to fetch Vault policies: %v", err))
			return 1
		}

		for i := range policies {
			c.UI.Info(fmt.Sprintf("\n%s", policies[i]))
		}
	default:
		p, err := client.Sys().GetPolicy(policy)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to fetch %s policy", policy))
			return 1
		}

		if p == "" {
			c.UI.Error(fmt.Sprintf("Could not find policy: %q", policy))
			return 1
		}
		c.UI.Info(fmt.Sprintf("%s", p))
	}

	return 0
}

// runPolicy creates vault policy rules for given vault policies
func (c *PolicyCommand) runPolicy(policies []*VaultPolicy) int {
	// more than 1 server requested
	pChan := make(chan error, 1)
	for _, policy := range policies {
		v, err := c.Client("", "")
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to fetch Vault client: %v", err))
			return 1
		}
		go func(p *VaultPolicy) {
			c.UI.Info(fmt.Sprintf("Attempting to configure policy: %s", p.Name))
			pChan <- v.Sys().PutPolicy(p.Name, p.Rules)
		}(policy)
	}
	// collect the results
	var errStatus bool
	for i := 0; i < len(policies); i++ {
		err := <-pChan
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to configure policy %s: %v", policies[i].Name, err))
			errStatus = true
			continue
		}
		c.UI.Info(fmt.Sprintf("Successfully configured vault policy: %s", policies[i].Name))
	}
	if errStatus {
		return 1
	}

	c.UI.Info(fmt.Sprintf("All requested policies successfully created"))

	return 0
}

// Synopsis provides a simple command description
func (c *PolicyCommand) Synopsis() string {
	return "Manage vault policies"
}

// Help returns detailed command help
func (c *PolicyCommand) Help() string {
	helpText := `
Usage: cam-vault policy [options]

    Manages vault policies

    This command connects to a Vault server and adds or modifies vault policies

    When policy is called on already existing policy it will modify it

General Options:
` + GeneralOptionsUsage() + `
init Options:

    -list 			List vault policy rules
    -config			Path to a config file which contains a list of vault policies to manage

`
	return strings.TrimSpace(helpText)
}
