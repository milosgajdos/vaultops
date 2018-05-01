package manifest

import (
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
