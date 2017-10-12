package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/vault/api"
	"github.com/milosgajdos83/vaultops/manifest"
)

// readVaultKeys parses vault config file stored locally as json and returns *VaultKeys
// If the file could not be found it returns empty struct.
// If the file can not be parsed it returns error.
func readVaultKeys(path string) (*VaultKeys, error) {
	var v VaultKeys
	// open vault keys file
	keys, err := ioutil.ReadFile(path)
	if err != nil {
		// if the file does not exist return empty VaultKeys
		if os.IsNotExist(err) {
			return &v, nil
		}

		return nil, err
	}

	if err := json.Unmarshal(keys, &v); err != nil {
		return nil, err
	}

	return &v, nil
}

// writeVaultKeys writes vault keys in a file stored in path
func writeVaultKeys(dir, fileName string, v *VaultKeys) error {
	// create directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// encode vault keys into json
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(dir, fileName), data, 0600)
}

// getVaultHosts parses configguration file in path and returns a slice of hosts
// If either the file can not be parsed or there are no hosts found it returns error
func getVaultHosts(path string) ([]string, error) {
	var hosts []string
	m, err := manifest.Parse(path)
	if err != nil {
		return nil, err
	}
	// if no hosts found, return erro
	if len(m.Hosts) == 0 {
		return nil, fmt.Errorf("No vault hosts found in %s", path)
	}
	hosts = append(hosts, m.Hosts...)

	return hosts, nil
}

// getVaultMounts parses configuration file in path returns a list of vault mounts
// If either the file can not be parsed or there are no hosts found it returns error
func getVaultMounts(path string) ([]*VaultMount, error) {
	m, err := manifest.Parse(path)
	if err != nil {
		return nil, err
	}
	// if no hosts found, return erro
	if len(m.Mounts) == 0 {
		return nil, fmt.Errorf("No vault mounts found in %s", path)
	}

	var mounts []*VaultMount
	for _, m := range m.Mounts {
		mount := &VaultMount{
			MountInput: &api.MountInput{
				Type: m.Type,
				Config: api.MountConfigInput{
					MaxLeaseTTL: m.MaxLeaseTTL,
				},
			},
			Path: m.Path,
		}
		mounts = append(mounts, mount)
	}

	return mounts, nil
}

// getVaultBackends parses configuration file in path and returns a list of vault backends
// If either the file can not be parsed or there are no backends found it returns error
func getVaultBackends(path string) ([]*VaultBackend, error) {
	m, err := manifest.Parse(path)
	if err != nil {
		return nil, err
	}
	// if no hosts found, return erro
	if len(m.Backends) == 0 {
		return nil, fmt.Errorf("No vault backends found in %s", path)
	}

	var backends []*VaultBackend
	for _, b := range m.Backends {
		backend := &VaultBackend{
			Name: b.Name,
		}
		if len(b.Roles) != 0 {
			parseRoles(b.Roles, backend)
		}
		if len(b.Certs) != 0 {
			parseCerts(b.Certs, backend)
		}
		backends = append(backends, backend)
	}

	return backends, nil
}

func parseRoles(roles manifest.Roles, b *VaultBackend) {
	for _, r := range roles {
		role := &VaultRole{
			Backend: b.Name,
			Name:    r.Name,
		}
		role.Config = map[string]interface{}{
			"allowed_domains":    r.AllowedDomains,
			"allow_bare_domains": r.AllowBareDomains,
			"allow_any_name":     r.AllowAnyName,
			"enforce_hostnames":  r.EnforceHostnames,
			"organization":       r.Organization,
		}
		b.Roles = append(b.Roles, role)
	}
}

func parseCerts(certs manifest.Certs, b *VaultBackend) {
	for _, c := range certs {
		cert := &VaultCert{
			Backend: b.Name,
			Type:    c.Type,
			Root:    c.Root,
			Role:    c.Role,
			Store:   c.Store,
		}
		cert.Config = map[string]interface{}{
			"common_name": c.CommonName,
			"ttl":         c.TTL,
			"ip_sans":     c.IPSans,
			"alt_names":   c.AltNames,
		}
		b.Certs = append(b.Certs, cert)
	}
}

// getVaultPolicies parses configuration file in path and returns a list of vault policies
// If either the file can not be parsed or there are no policies found it returns error
func getVaultPolicies(path string) ([]*VaultPolicy, error) {
	m, err := manifest.Parse(path)
	if err != nil {
		return nil, err
	}
	// if no hosts found, return erro
	if len(m.Policies) == 0 {
		return nil, fmt.Errorf("No vault policies found in %s", path)
	}

	var policies []*VaultPolicy
	for _, p := range m.Policies {
		policy := &VaultPolicy{
			Name:  p.Name,
			Rules: p.Rules,
		}

		policies = append(policies, policy)
	}

	return policies, nil
}
