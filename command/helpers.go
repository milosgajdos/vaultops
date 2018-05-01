package command

import (
	"fmt"

	"github.com/milosgajdos83/vaultops/cipher"
	"github.com/milosgajdos83/vaultops/cloud/aws"
	"github.com/milosgajdos83/vaultops/cloud/gcp"
	"github.com/milosgajdos83/vaultops/store"
	"github.com/milosgajdos83/vaultops/store/local"
)

// VaultKeyStore creates vault keys store
func VaultKeyStore(storeType string, m *Meta) (s store.Store, err error) {
	switch storeType {
	case "local":
		s, err = local.NewStore(m.flagKeyLocalPath)
		if err != nil {
			return nil, err
		}
	case "s3":
		s, err = aws.NewS3(m.flagStorageBucket, m.flagStorageKey)
		if err != nil {
			return nil, err
		}
	case "gcs":
		s, err = gcp.NewGCS(m.flagStorageBucket, m.flagStorageKey)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Unsupported store: %s", storeType)
	}

	return s, nil
}

// VaultKeyCipher returns KMS key handle to use for encrypting and decrypting keys
func VaultKeyCipher(m *Meta) (c cipher.Cipher, err error) {
	switch m.flagKMSProvider {
	case "aws":
		c, err = aws.NewKMS(m.flagAwsKmsID)
		if err != nil {
			return nil, err
		}
	case "gcp":
		c, err = gcp.NewKMS(m.flagGcpKmsProject, m.flagGcpKmsRegion, m.flagGcpKmsKeyRing, m.flagGcpKmsCryptoKey)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Unsupported cipher provider: %s", m.flagKMSProvider)
	}

	return c, nil
}

// Redact returns string of characters ch of length long
func Redact(ch rune, length int) string {
	data := make([]rune, length)
	for i := 0; i < length; i++ {
		data[i] = ch
	}

	return string(data)
}
