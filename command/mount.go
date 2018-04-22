package command

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/milosgajdos83/vaultops/manifest"
	"github.com/ryanuber/columnize"
)

// MountCommand implements vault mounts
// It fulfills cli.Command interface
type MountCommand struct {
	// meta flags contain vault client config
	Meta
}

// Run runs mount command which mounts a secret backend
// If mount command fails it returns non-zero integer
func (c *MountCommand) Run(args []string) int {
	var list bool
	var config string
	// create command flags
	flags := c.Meta.FlagSet("mount", FlagSetDefault)
	flags.Usage = func() { c.UI.Info(c.Help()) }
	flags.BoolVar(&list, "list", false, "")
	flags.StringVar(&config, "config", "", "")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	if list {
		return c.runMountList()
	}

	var mounts manifest.Mounts
	if config != "" {
		m, err := manifest.Parse(config)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to parse config %s: %s", config, err))
			return 1
		}
		// get mounts
		mounts = m.GetMounts()
	}

	c.UI.Info(fmt.Sprintf("Attempting to mount vault backends:"))
	for _, mount := range mounts {
		c.UI.Info(fmt.Sprintf("\tType: %s Path: %s TTL: %s", mount.Type, mount.Path, mount.MaxLeaseTTL))
	}

	return c.runMount(mounts)
}

// runMountList list mounted secret backends
func (c *MountCommand) runMountList() int {
	client, err := c.Client("", "")
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to fetch Vault client: %v", err))
		return 1
	}

	mounts, err := client.Sys().ListMounts()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to fetch Vault mounts: %v", err))
		return 1
	}

	// ripped off https://github.com/hashicorp/vault/blob/master/command/mounts.go#L39-L77
	paths := make([]string, 0, len(mounts))
	for path := range mounts {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	columns := []string{"Path | Type | Accessor | Plugin | Default TTL | Max TTL | Force No Cache | Replication Behavior | Description"}
	for _, path := range paths {
		mount := mounts[path]
		pluginName := "n/a"
		if mount.Config.PluginName != "" {
			pluginName = mount.Config.PluginName
		}
		defTTL := "system"
		switch {
		case mount.Type == "system":
			defTTL = "n/a"
		case mount.Type == "cubbyhole":
			defTTL = "n/a"
		case mount.Config.DefaultLeaseTTL != 0:
			defTTL = strconv.Itoa(mount.Config.DefaultLeaseTTL)
		}
		maxTTL := "system"
		switch {
		case mount.Type == "system":
			maxTTL = "n/a"
		case mount.Type == "cubbyhole":
			maxTTL = "n/a"
		case mount.Config.MaxLeaseTTL != 0:
			maxTTL = strconv.Itoa(mount.Config.MaxLeaseTTL)
		}
		replicatedBehavior := "replicated"
		if mount.Local {
			replicatedBehavior = "local"
		}
		columns = append(columns, fmt.Sprintf(
			"%s | %s | %s | %s | %s | %s | %v | %s | %s", path, mount.Type, mount.Accessor, pluginName, defTTL, maxTTL,
			mount.Config.ForceNoCache, replicatedBehavior, mount.Description))
	}

	c.UI.Info(columnize.SimpleFormat(columns))

	return 0
}

// runMount mounts secrets backend mounts
func (c *MountCommand) runMount(mounts manifest.Mounts) int {
	// more than 1 server requested
	for _, mount := range mounts {
		client, err := c.Client("", "")
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to fetch Vault client: %v", err))
			return 1
		}
		c.UI.Info(fmt.Sprintf("Attempting to mount %s backend in path: %s", mount.Type, mount.Path))
		apiMount := &api.MountInput{
			Type: mount.Type,
			Config: api.MountConfigInput{
				MaxLeaseTTL: mount.MaxLeaseTTL,
			},
		}
		if err := client.Sys().Mount(mount.Path, apiMount); err != nil {
			c.UI.Error(fmt.Sprintf("Failed to mount %s backend in path: %s", mount.Type, mount.Path))
		}
	}
	c.UI.Info(fmt.Sprintf("Finished mounting vault backends"))

	return 0
}

// Synopsis provides a simple command description
func (c *MountCommand) Synopsis() string {
	return "Mount a new vault secret backend"
}

// Help returns detailed command help
func (c *MountCommand) Help() string {
	helpText := `
Usage: vaultops mount [options]

    Mount a new vault secret backend

    This command connects to a Vault server and mounts a secret backend
    with requested options.

    When mount is called on already mounted backend it will modify it

General Options:
` + GeneralOptionsUsage() + `
mount Options:

    -list			Lists the mounted backends, their mount points and various vault mount information
    -config			Path to a config file which contains a list of mounts and mount options

`
	return strings.TrimSpace(helpText)
}
