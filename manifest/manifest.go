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

// Mounts provides vault mount configurations
type Mounts []struct {
	// Type of the vault mount
	Type string `yaml:"type"`
	// Path is a vault mount path
	Path string `yaml:"path"`
	// DefaultLeaseTTL default lease duration, specified as a go string duration like "5s"
	DefaultLeaseTTL string `yaml:"default_lease_ttl"`
	// MaxLeaseTTL is max TTL lease
	MaxLeaseTTL string `yaml:"max_lease_ttl,omitempty"`
	// ForceNoCache disables caching
	ForceNoCache bool `yaml:"force_no_cache"`
}

// Roles allows to configura vault roles
type Roles []struct {
	// Name is a role name
	Name string `yaml:"name"`
	// MaxTTL is Maximum Time To Live
	MaxTTL string `yaml:"max_ttl"`
	// TTL Time To Live
	TTL string `yaml:"ttl"`
	// AllowLocalhost specifies if clients can request certificates for localhost
	AllowLocalhost bool `yaml:"allow_localhost"`
	// AllowedBaseDomain specifies domains of the role
	AllowedBaseDomain string `yaml:"allowed_base_domain"`
	// AllowedDomains specifies allowed_domains vault parameter
	AllowedDomains []string `yaml:"allowed_domains_list"`
	// AllowBareDomains specifies allow_bare_domains vault parameter
	AllowBareDomains bool `yaml:"allow_bare_domains"`
	// AllowSubdomains  specifies if clients can request certs with CNs that are subdomains of the CNs allowed
	AllowSubdomains bool `yaml:"allow_subdomains"`
	// AllowGlobDomains allows ames specified in allowed_domains to contain glob patterns
	AllowGlobDomains bool `yaml:"allow_glob_domains"`
	// AllowAnyName specifies if clients can request any common_name
	AllowAnyName bool `yaml:"allow_any_name"`
	// EnforceHostnames specifies enforce_hostnames vault parameter
	EnforceHostnames bool `yaml:"enforce_hostnames,omitempty"`
	// AllowIPSANs specifies if clients can request IP Subject Alternative Names
	AllowIPSANs bool `yaml:"allow_ip_sans"`
	//ServerFlag specifies if certificates are flagged for server use
	ServerFlag bool `yaml:"server_flag"`
	// ClientFlag specifies if certificates are flagged for client use
	ClientFlag bool `yaml:"client_flag"`
	// CodeSigningFlag  specifies if certificates are flagged for code signing use
	CodeSigningFlag bool `yaml:"code_signing_flag"`
	// EmailProtectionFlag specifies if certificates are flagged for email protection use
	EmailProtectionFlag bool `yaml:"email_protection_flag"`
	// UseCSRCommonName CSR will be used instead of taken from the yaml data for the common name
	UseCSRCommonName bool `yaml:"use_csr_common_name"`
	// UseCSRSANsaCSR will be used instead of taken from the yaml data for subject alternate names
	UseCSRSANs bool `yaml:"use_csr_sans"`
	// KeyType specifies the type of key to generate for generated private keys
	KeyType string `yaml:"key_type"`
	// KeyBits specifies the number of bits to use for the generated keys
	KeyBits int `yaml:"key_bits"`
	// OU specifies the OU (OrganizationalUnit) values in the subject field of issued certificates
	OU []string `yaml:"ou_list"`
	// Organization specifies the O (Organization) values in the subject field of issued certificates
	Organization string `yaml:"organization_list"`
	// GenerateLease specifies if certs issued against this role will have Vault leases attached to them
	GenerateLease *bool `yaml:"generate_lease,omitempty"`
	// NoStore if set, certificates issued/signed against this role will not be stored in the storage backend
	NoStore bool `yaml:"no_store"`
}

// Certs are SSL certificates
type Certs []struct {
	// Action is a kinda pf PKI backend action
	Action string `yaml:"action"`
	// Name is a name identificator of certificate
	Name string `yaml:"name"`
	// Role is vault role
	Role string `yaml:"role,omitempty"`
	// Kind is TLS cert kind: root, intermediate, client
	Kind string `yaml:"kind"`
	// Type defines type of SSL certificate (internal or exported)
	Type string `yaml:"type,omitempty"`
	// CommonName is SSL cert common name
	CommonName string `yaml:"common_name"`
	// AltNames provides a list of alternative names
	AltNames string `yaml:"alt_names,omitempty"`
	// IPSans is SSL ip_sanse list of IP addresses
	IPSans string `yaml:"ip_sans,omitempty"`
	// TTL is SSL certificate TTL
	TTL string `yaml:"ttl"`
	// Format specifies the format for returned data. Can be pem, der, or pem_bundle
	Format string `yaml:"format"`
	// PrivateKeyFormat specifies the format for marshaling the private key
	PrivateKeyFormat string `yaml:"private_key_format"`
	// KeyType specifies the type of key to generate for generated private keys
	KeyType string `yaml:"key_type"`
	// KeyBits specifies the number of bits to use for the generated keys
	KeyBits int `yaml:"key_bits"`
	// MaxPathLength specifies the maximum path length to encode in the generated cert
	MaxPathLength int `yaml:"max_path_length"`
	// ExcludeCNFromSANS if set, the given CN will not be included in DNS or Email Subject Alternate Names
	ExcludeCNFromSANS bool `yaml:"exclude_cn_from_sans"`
	// PermitDNSDomains comma separated string (or, string array) containing DNS domains
	PermitDNSDomains string `yaml:"permitted_dns_domains"`
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
	Hosts    `yaml:"hosts,omitempty"`
	Mounts   `yaml:"mounts,omitempty"`
	Backends `yaml:"backends,omitempty"`
	Policies `yaml:"policies,omitempty"`
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

// GetMounts returns manifest mounts
func (m *Manifest) GetMounts() Mounts {
	return m.Mounts
}

// GetBackends returns manifest backends
func (m *Manifest) GetBackends() Backends {
	return m.Backends
}

// GetPolicies returns manifest policies
func (m *Manifest) GetPolicies() Policies {
	return m.Policies
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
