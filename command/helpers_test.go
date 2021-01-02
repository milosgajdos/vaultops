package command

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
