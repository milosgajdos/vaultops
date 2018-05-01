package manifest

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Hosts are vault server hosts to initialize and unseal
type Hosts struct {
	// Init is a slice of vault servers to initialize
	Init []string `yaml:"init,omitempty"`
	// Unseal is a slice of vault servers to unseal
	Unseal []string `yaml:"unseal,omitempty"`
}

// Manifest holds vault setup configuration
type Manifest struct {
	Hosts `yaml:"hosts,omitempty"`
}

// GetHosts returns hosts for given command
func (m *Manifest) GetHosts(cmd string) ([]string, error) {
	var hosts []string
	// if no hosts found, return erro
	switch cmd {
	case "init":
		hosts = append(hosts, m.Hosts.Init...)
	case "unseal":
		hosts = append(hosts, m.Hosts.Unseal...)
	default:
		return nil, fmt.Errorf("Unsupported command: %s", cmd)
	}

	return hosts, nil
}

// Parse parses configuration file stored in path and returns pointer to Manifest
// It fails with error if the supplied configuration file can not be read or parsed as valid config
func Parse(path string) (*Manifest, error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m Manifest
	if err := yaml.Unmarshal(f, &m); err != nil {
		return nil, err
	}

	return &m, nil
}
