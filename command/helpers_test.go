package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	// write data to temp file
	if _, err := f.Write(data); err != nil {
		return "", err
	}

	return f.Name(), nil
}

func Test_readVaultKeys(t *testing.T) {
	vaultKeys := []string{
		"key1",
		"key2",
	}
	data := []byte(` {
		"root_token": "root-token",
		"token": "vault-token",
		"master_keys": [
			"` + vaultKeys[0] + `",
			"` + vaultKeys[1] + `"
		]
	}`)

	f, err := makeTestFile([]byte(data))
	defer os.Remove(f)
	assert.NoError(t, err)

	keys, err := readVaultKeys(f)
	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.Equal(t, keys.RootToken, "root-token")
	assert.Equal(t, keys.Token, "vault-token")
	for i := 0; i < len(vaultKeys); i++ {
		assert.Equal(t, keys.MasterKeys[i], vaultKeys[i])
	}

	// garbage file path causes error
	keys, err = readVaultKeys("foobar/dfd")
	assert.NoError(t, err)
	assert.Equal(t, keys.RootToken, "")
	assert.Equal(t, keys.Token, "")
	assert.True(t, len(keys.MasterKeys) == 0)

	// garbage data
	data = []byte(` {sdfsdfdff_dsf}`)
	f2, err := makeTestFile([]byte(data))
	defer os.Remove(f2)
	assert.NoError(t, err)
	keys, err = readVaultKeys(f2)
	assert.Nil(t, keys)
	assert.Error(t, err)
}

func Test_writeVaultKeys(t *testing.T) {
	dir := os.TempDir()
	file := "vault.json"
	path := filepath.Join(dir, file)

	rootToken := "root-token"
	token := "vault-token"
	keys := []string{"key1", "key2"}
	vk := &VaultKeys{RootToken: rootToken, Token: token, MasterKeys: keys}
	err := writeVaultKeys(dir, file, vk)
	assert.NoError(t, err)
	defer os.Remove(path)

	// can't create dir
	err = writeVaultKeys("/etc/foo", "foo", vk)
	assert.Error(t, err)
}

func Test_getVaultHosts(t *testing.T) {
	vaultHosts := []string{
		"http://192.168.1.101:8200",
		"http://192.168.1.102:8200",
	}
	// raw data
	data := `hosts:
  - ` + vaultHosts[0] + `
  - ` + vaultHosts[1]

	f, err := makeTestFile([]byte(data))
	defer os.Remove(f)
	assert.NoError(t, err)
	hosts, err := getVaultHosts(f)
	assert.NoError(t, err)

	for i := 0; i < len(hosts); i++ {
		assert.Equal(t, hosts[i], vaultHosts[i])
	}
	// garbage file path causes error
	hosts, err = getVaultHosts("foobar/dfd")
	assert.Error(t, err)
	// no parsed hosts returns error
	data = `foo:
  - bar: foobar
`
	f2, err := makeTestFile([]byte(data))
	defer os.Remove(f2)
	assert.NoError(t, err)
	hosts, err = getVaultHosts(f2)
	assert.Error(t, err)
}

func Test_getVaultMounts(t *testing.T) {
	vaultMounts := []struct {
		Path string
		Type string
	}{
		{"pki-path", "pki"},
		{"generic-path", "generic"},
	}
	// test data
	data := `mounts:
  - path: "` + vaultMounts[0].Path + `"
    type: "` + vaultMounts[0].Type + `"
    max-lease-ttl: "876000h"
  - path: "` + vaultMounts[1].Path + `"
    type: "` + vaultMounts[1].Type + `"
    max-lease-ttl: "876000h"
`
	f, err := makeTestFile([]byte(data))
	defer os.Remove(f)
	assert.NoError(t, err)
	mounts, err := getVaultMounts(f)
	assert.NoError(t, err)

	for i := 0; i < len(mounts); i++ {
		assert.Equal(t, vaultMounts[i].Path, mounts[i].Path)
		assert.Equal(t, vaultMounts[i].Type, mounts[i].MountInput.Type)
	}

	// random file path causes error
	mounts, err = getVaultMounts("foobar/dfd")
	assert.Error(t, err)
	// no parsed hosts returns error
	data = `foo:
  - bar: foobar
`
	f2, err := makeTestFile([]byte(data))
	defer os.Remove(f2)
	assert.NoError(t, err)
	mounts, err = getVaultMounts(f2)
	assert.Error(t, err)
}

func Test_getVaultBackends(t *testing.T) {
	vaultRoles := []struct {
		BEName           string
		Name             string
		AllowedDomains   string
		AllowBareDomains bool
		AllowAnyName     bool
		EnforceHostnames bool
		Organization     string
	}{
		{"k8s-1", "api", "kubernetes", true, true, false, "org1"},
	}

	vaultCerts := []struct {
		Name       string
		Type       string
		Root       bool
		Role       string
		Store      bool
		CommonName string
		TTL        string
		IPSans     string
	}{
		{"k8s-1", "internal", true, "", false, "comm-name", "100h", ""},
		{"k8s-2", "", false, "foorole", false, "foo-name", "87h", "10.1.1.1"},
	}
	vaultBackends := []string{"k8s-1", "k8s-2"}

	// test data
	data := `backends:
  - name: "` + vaultBackends[0] + `"
    roles:
      - name: "` + vaultRoles[0].Name + `"
        allowed-domains: "` + vaultRoles[0].AllowedDomains + `"
        allow-bare-domains: ` + fmt.Sprintf("%t", vaultRoles[0].AllowBareDomains) + `
        allow-any-name: ` + fmt.Sprintf("%t", vaultRoles[0].AllowAnyName) + `
        enforce-hostnames: ` + fmt.Sprintf("%t", vaultRoles[0].EnforceHostnames) + `
        organization: "` + vaultRoles[0].Organization + `"
    certificates:
      - name: "` + vaultCerts[0].Name + `"
        root: ` + fmt.Sprintf("%t", vaultCerts[0].Root) + `
        common-name: "` + vaultCerts[0].CommonName + `"
        ttl: "` + vaultCerts[0].TTL + `"
        type: "` + vaultCerts[0].Type + `"
  - name: "` + vaultBackends[1] + `"
    certificates:
      - name: "` + vaultCerts[1].Name + `"
        common-name: "` + vaultCerts[1].CommonName + `"
        ttl: "` + vaultCerts[1].TTL + `"
        ip-sans: "` + vaultCerts[1].IPSans + `"
        role: "` + vaultCerts[1].Role + `"
        store: ` + fmt.Sprintf("%t", vaultCerts[1].Store) + `
`
	f, err := makeTestFile([]byte(data))
	defer os.Remove(f)
	assert.NoError(t, err)
	backends, err := getVaultBackends(f)
	assert.NoError(t, err)
	assert.NotNil(t, backends)

	for i := 0; i < len(backends); i++ {
		assert.Equal(t, vaultBackends[i], backends[i].Name)
		assert.Equal(t, vaultBackends[i], backends[i].Certs[0].Backend)
	}

	// random file path causes error
	backends, err = getVaultBackends("foobar/dfd")
	assert.Error(t, err)
	// no parsed backends returns error
	data = `foo:
  - bar: foobar
`
	f2, err := makeTestFile([]byte(data))
	defer os.Remove(f2)
	assert.NoError(t, err)
	backends, err = getVaultBackends(f2)
	assert.Error(t, err)
}

func Test_getVaultPolicies(t *testing.T) {
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
	policies, err := getVaultPolicies(f)
	assert.NoError(t, err)

	for i := 0; i < len(policies); i++ {
		assert.Equal(t, vaultPolicies[i].Name, policies[i].Name)
		assert.Equal(t, vaultPolicies[i].Rules, policies[i].Rules)
	}

	// random file path causes error
	policies, err = getVaultPolicies("foobar/dfd")
	assert.Error(t, err)
	// no parsed policies returns error
	data = `foo:
  - bar: foobar
`
	f2, err := makeTestFile([]byte(data))
	defer os.Remove(f2)
	assert.NoError(t, err)
	policies, err = getVaultPolicies(f2)
	assert.Error(t, err)
}
