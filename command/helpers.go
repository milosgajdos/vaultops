package command

import (
	"fmt"
	"path/filepath"

	"github.com/milosgajdos83/vaultops/kms"
	"github.com/milosgajdos83/vaultops/store"
)

// VaultKeyStore creates vault keys store
func VaultKeyStore(storeType string) (s Store, err error) {
	switch storeType {
	case "local":
		path := filepath.Join(localDir, localFile)
		s, err = store.NewLocal(path)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Unsupported store: %s", storeType)
	}

	return s, nil
}

// VaultKeyCipher returns KMS key handle to use for encrypting and decrypting keys
func VaultKeyCipher(m *Meta) (c Cipher, err error) {
	switch m.flagKMSProvider {
	case "aws":
		c, err = kms.NewAWS(m.flagAwsKmsID)
		if err != nil {
			return nil, err
		}
	case "gcp":
		c, err = kms.NewGCP(m.flagGcpKmsProject, m.flagGcpKmsRegion, m.flagGcpKmsKeyRing, m.flagGcpKmsCryptoKey)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Unsupported cipher provider: %s", m.flagKMSProvider)
	}

	return c, nil
}
