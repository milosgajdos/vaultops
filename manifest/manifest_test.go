package manifest

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// makeTestFile creates a temporary test file and writes data into it
// It returns full path to newly created path  or error if the file fails to be created
func makeTestFile(data []byte) (string, error) {
	// create temp file for testing
	f, err := ioutil.TempFile("", "test")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(data); err != nil {
		return "", err
	}

	return f.Name(), nil
}

func TestParse(t *testing.T) {
	// non-existen path
	m, err := Parse("foobar.yaml")
	assert.Error(t, err)
	assert.Nil(t, m)

	// invalid config
	invalid := `
hosts:
  init:
    foo: bar
`
	invPath, err := makeTestFile([]byte(invalid))
	defer os.Remove(invPath)
	assert.NoError(t, err)
	m, err = Parse(invPath)
	assert.Error(t, err)
	assert.Nil(t, m)

	// valid config
	valid := `
hosts:
  init:
    - one
    - two
mounts:
  - path: vPath
    type: pki
`
	vPath, err := makeTestFile([]byte(valid))
	defer os.Remove(vPath)
	assert.NoError(t, err)
	m, err = Parse(vPath)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.EqualValues(t, m.Hosts.Init, []string{"one", "two"})
	assert.EqualValues(t, m.Mounts[0].Path, "vPath")
	assert.EqualValues(t, m.Mounts[0].Type, "pki")
}

func TestGetHosts(t *testing.T) {
	vaultHosts := []string{
		"http://192.168.1.101:8200",
		"http://192.168.1.102:8200",
	}

	// init hosts raw data
	data := `hosts:
  init:
    - ` + vaultHosts[0] + `
    - ` + vaultHosts[1]

	initPath, err := makeTestFile([]byte(data))
	defer os.Remove(initPath)
	assert.NoError(t, err)
	m, err := Parse(initPath)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	hosts, err := m.GetHosts("init")
	assert.NoError(t, err)

	for i := 0; i < len(hosts); i++ {
		assert.Equal(t, hosts[i], vaultHosts[i])
	}

	// unseal hosts raw data
	data = `hosts:
  unseal:
    - ` + vaultHosts[0] + `
    - ` + vaultHosts[1]

	unsealPath, err := makeTestFile([]byte(data))
	defer os.Remove(unsealPath)
	assert.NoError(t, err)
	m, err = Parse(unsealPath)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	hosts, err = m.GetHosts("unseal")
	assert.NoError(t, err)

	for i := 0; i < len(hosts); i++ {
		assert.Equal(t, hosts[i], vaultHosts[i])
	}

	// unsupported action causes error
	hosts, err = m.GetHosts("foobar")
	assert.Error(t, err)
}

func TestGetMounts(t *testing.T) {
	vaultMounts := []struct {
		Path        string
		Type        string
		MaxLeaseTTL string
	}{
		{"pki-path", "pki", "876000h"},
		{"generic-path", "generic", "876000h"},
	}
	// test data
	data := `mounts:
  - path: "` + vaultMounts[0].Path + `"
    type: "` + vaultMounts[0].Type + `"
    max_lease_ttl: "` + vaultMounts[0].MaxLeaseTTL + `"
  - path: "` + vaultMounts[1].Path + `"
    type: "` + vaultMounts[1].Type + `"
    max_lease_ttl: "` + vaultMounts[1].MaxLeaseTTL + `"
`
	f, err := makeTestFile([]byte(data))
	defer os.Remove(f)
	assert.NoError(t, err)
	m, err := Parse(f)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	mounts := m.GetMounts()
	assert.NoError(t, err)

	for i := 0; i < len(mounts); i++ {
		assert.Equal(t, vaultMounts[i].Path, mounts[i].Path)
		assert.Equal(t, vaultMounts[i].Type, mounts[i].Type)
		assert.Equal(t, vaultMounts[i].MaxLeaseTTL, mounts[i].MaxLeaseTTL)
	}
}

func TestGetBackends(t *testing.T) {
	vaultRoles := []struct {
		Name             string
		AllowedDomains   string
		AllowBareDomains bool
		AllowAnyName     bool
		EnforceHostnames bool
		Organization     string
	}{
		{"api", "kubernetes", true, true, false, "org1"},
	}

	vaultCerts := []struct {
		Action     string
		Name       string
		Type       string
		Kind       string
		Role       string
		CommonName string
		TTL        string
		IPSans     string
	}{
		{"generate", "k8s-1", "internal", "root", "", "comm-name", "100h", ""},
		{"generate", "k8s-2", "", "intermediate", "foorole", "foo-name", "87h", "10.1.1.1"},
	}
	vaultBackends := []string{"k8s-1", "k8s-2"}

	// test data
	data := `backends:
  - name: "` + vaultBackends[0] + `"
    roles:
      - name: "` + vaultRoles[0].Name + `"
        allowed_domains: "` + vaultRoles[0].AllowedDomains + `"
        allow_bare_domains: ` + fmt.Sprintf("%t", vaultRoles[0].AllowBareDomains) + `
        allow_any_name: ` + fmt.Sprintf("%t", vaultRoles[0].AllowAnyName) + `
        enforce_hostnames: ` + fmt.Sprintf("%t", vaultRoles[0].EnforceHostnames) + `
        organization: "` + vaultRoles[0].Organization + `"
    certificates:
      - name: "` + vaultCerts[0].Name + `"
        action: "` + vaultCerts[0].Action + `"
        kind: ` + vaultCerts[0].Kind + `
        common_name: "` + vaultCerts[0].CommonName + `"
        ttl: "` + vaultCerts[0].TTL + `"
        type: "` + vaultCerts[0].Type + `"
  - name: "` + vaultBackends[1] + `"
    certificates:
      - name: "` + vaultCerts[1].Name + `"
        action: "` + vaultCerts[1].Action + `"
        kind: ` + vaultCerts[1].Kind + `
        common_name: "` + vaultCerts[1].CommonName + `"
        ttl: "` + vaultCerts[1].TTL + `"
        ip_sans: "` + vaultCerts[1].IPSans + `"
        role: "` + vaultCerts[1].Role + `"
`
	f, err := makeTestFile([]byte(data))
	defer os.Remove(f)
	assert.NoError(t, err)
	m, err := Parse(f)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	backends := m.GetBackends()
	assert.NotNil(t, backends)
}

func TestGetPolicies(t *testing.T) {
	vaultPolicies := []struct {
		Name  string
		Rules string
	}{
		{"pki-path", "path \"secret/k8s-ca/*\" {policy = \"write\"}"},
	}
	// test data
	data := `policies:
- name: "` + vaultPolicies[0].Name + `"
  rules: |
    ` + vaultPolicies[0].Rules

	f, err := makeTestFile([]byte(data))
	defer os.Remove(f)
	assert.NoError(t, err)
	m, err := Parse(f)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	policies := m.GetPolicies()
	assert.NoError(t, err)

	for i := 0; i < len(policies); i++ {
		assert.Equal(t, vaultPolicies[i].Name, policies[i].Name)
		assert.Equal(t, vaultPolicies[i].Rules, policies[i].Rules)
	}
}
