package manifest

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Hosts are vault API servers
type Hosts []string

// Mounts provides vault mount configurations
type Mounts []struct {
	// Type of the vault mount
	Type string `yaml:"type"`
	// Path is a vault mount path
	Path string `yaml:"path"`
	// MaxLeaseTTL is max TTL lease
	MaxLeaseTTL string `yaml:"max-lease-ttl,omitempty"`
}

// Roles allows to configura vault roles
type Roles []struct {
	// Backend is vault backend
	Backend string `yaml:"backend,omitempty"`
	// Name is a role name
	Name string `yaml:"name"`
	// AllowedDomains specifies allowed_domains vault parameter
	AllowedDomains string `yaml:"allowed-domains,omitempty"`
	// AllowBareDomains specifies allow_bare_domains vault parameter
	AllowBareDomains bool `yaml:"allow-bare-domains,omitempty"`
	// AllowAnyName specifis if clients can request any common_name
	AllowAnyName bool `yaml:"allow-any-name"`
	// EnforceHostnames specifies enforce_hostnames vault parameter
	EnforceHostnames bool `yaml:"enforce-hostnames,omitempty"`
	// Organization is O (Organization) value in the subject field of issued SSL certificate
	Organization string `yaml:"organization,omitempty"`
}

// Certs are SSL certificates
type Certs []struct {
	// Backend is vault backend
	Backend string `yaml:"backend,omitempty"`
	// Name is a name identificator of certificate
	Name string `yaml:"name"`
	// Type defines type of SSL certificate (internal or exported)
	Type string `yaml:"type,omitempty"`
	// Root specifies is the certificate is a root certificate
	Root bool `yaml:"root,omitempty"`
	// CommonName is SSL cert common name
	CommonName string `yaml:"common-name"`
	// TTL is SSL certificate TTL
	TTL string `yaml:"ttl"`
	// IPSans is SSL ip_sanse list of IP addresses
	IPSans string `yaml:"ip-sans,omitempty"`
	// AltNames provides a list of alternative names
	AltNames string `yaml:"alt-names,omitempty"`
	// Role is vault role
	Role string `yaml:"role,omitempty"`
	// Store request the certificate to be stored in vault
	Store bool `yaml:"store,omitempty"`
}

// Backends are vault secret backends
type Backends []struct {
	// Name of vault backend
	Name string `yaml:"name,omitempty"`
	// Roles allows to configura vault roles
	Roles `yaml:"roles,omitempty"`
	// Certs are SSL certificates issued by vault
	Certs `yaml:"certificates,omitempty"`
}

// Policies provide vault ACL policies
type Policies []struct {
	// Name is a name of vault policiy
	Name string `yaml:"name"`
	// Rules stores ACL policy rules
	Rules string `yaml:"rules"`
}

// Manifest holds vault setup configuration
type Manifest struct {
	Hosts    `yaml:"hosts"`
	Mounts   `yaml:"mounts,omitempty"`
	Backends `yaml:"backends,omitempty"`
	Policies `yaml:"policies,omitempty"`
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
