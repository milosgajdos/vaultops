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
	c, err := Parse("foobar.yaml")
	assert.Error(t, err)
	assert.Nil(t, c)

	// invalid config
	invalid := `
hosts:
  init:
    foo: bar
`
	invPath, err := makeTestFile([]byte(invalid))
	defer os.Remove(invPath)
	c, err = Parse(invPath)
	assert.Error(t, err)
	assert.Nil(t, c)

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
	c, err = Parse(vPath)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.EqualValues(t, c.Hosts.Init, []string{"one", "two"})
	assert.EqualValues(t, c.Mounts[0].Path, "vPath")
	assert.EqualValues(t, c.Mounts[0].Type, "pki")
}
