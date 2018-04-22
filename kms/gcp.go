package kms

import (
	"encoding/base64"
	"fmt"

	"context"

	"golang.org/x/oauth2/google"
	gcp "google.golang.org/api/cloudkms/v1"
)

const (
	keyPath = "projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s"
)

// GcpKMS is GCP KMS client
type GcpKMS struct {
	client  *gcp.Service
	keyPath string
}

// NewGCP returns new GCP KMS client
// It returns error if either Google OAuth client or CloudKMS client failed to be created
func NewGCP(project, location, keyring, cryptoKey string) (*GcpKMS, error) {
	ctx := context.Background()
	oauth, err := google.DefaultClient(ctx, gcp.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Google OAuth client: %s", err.Error())
	}

	client, err := gcp.New(oauth)
	if err != nil {
		return nil, fmt.Errorf("Failed to create GCP CloudKMS client: %s", err.Error())
	}

	return &GcpKMS{
		client:  client,
		keyPath: fmt.Sprintf(keyPath, project, location, keyring, cryptoKey),
	}, nil
}

// Encrypt encrypts plainText data and returns it
func (k *GcpKMS) Encrypt(plainText []byte) ([]byte, error) {
	resp, err := k.client.Projects.Locations.KeyRings.CryptoKeys.Encrypt(k.keyPath, &gcp.EncryptRequest{
		Plaintext: base64.StdEncoding.EncodeToString(plainText),
	}).Do()

	if err != nil {
		return nil, fmt.Errorf("Failed to encrypt data: %s", err.Error())
	}

	return base64.StdEncoding.DecodeString(resp.Ciphertext)
}

// Decrypt decrypts cipherText data and returns it
func (k *GcpKMS) Decrypt(cipherText []byte) ([]byte, error) {
	resp, err := k.client.Projects.Locations.KeyRings.CryptoKeys.Decrypt(k.keyPath, &gcp.DecryptRequest{
		Ciphertext: base64.StdEncoding.EncodeToString(cipherText),
	}).Do()

	if err != nil {
		return nil, fmt.Errorf("Failed to decrypt data: %s", err.Error())
	}

	return base64.StdEncoding.DecodeString(resp.Plaintext)
}
