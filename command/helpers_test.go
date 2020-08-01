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

func makeTestMeta() *Meta {
	return &Meta{
		flagKMSProvider:     "aws",
		flagAwsKmsID:        "id",
		flagGcpKmsCryptoKey: "key",
		flagGcpKmsKeyRing:   "ring",
		flagGcpKmsRegion:    "region",
		flagGcpKmsProject:   "project",
		flagStorageBucket:   "bucket",
		flagStorageKey:      "somekey",
		flagKeyLocalPath:    filepath.Join(localDir, localFile),
	}
}

func TestVaultKeyStore(t *testing.T) {
	m := makeTestMeta()

	testCases := []struct {
		storeType string
		meta      *Meta
		result    error
	}{
		{"local", m, nil},
		{"s3", m, nil},
		{"foobar", m, fmt.Errorf("unsupported store: foobar")},
	}
	defer os.Remove(filepath.Join(localDir, localFile))

	for _, tc := range testCases {
		s, err := VaultKeyStore(tc.storeType, m)
		if tc.result == nil {
			assert.NotNil(t, s)
		} else {
			assert.Nil(t, s)
			assert.EqualError(t, err, tc.result.Error())
		}
	}
}

func TestVaultKeyCipher(t *testing.T) {
	m := makeTestMeta()

	m.flagKMSProvider = "aws"
	k, err := VaultKeyCipher(m)
	assert.NoError(t, err)
	assert.NotNil(t, k)

	m.flagKMSProvider = "foobar"
	k, err = VaultKeyCipher(m)
	assert.Error(t, err)
	assert.Nil(t, k)
}

func TestRedact(t *testing.T) {
	ch := rune('X')
	length := 5
	str := Redact(ch, length)
	assert.Equal(t, len(str), length)
}
