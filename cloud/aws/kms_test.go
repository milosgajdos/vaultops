package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/stretchr/testify/assert"
)

func TestNewKMS(t *testing.T) {
	k, err := NewKMS("")
	assert.Nil(t, k)
	assert.Error(t, err)

	k, err = NewKMS("some-key")
	assert.NotNil(t, k)
	assert.NoError(t, err)
}

type mockKMS struct {
	EncryptFunc func(*kms.EncryptInput) (*kms.EncryptOutput, error)
	DecryptFunc func(*kms.DecryptInput) (*kms.DecryptOutput, error)
	keyID       string
}

func (m *mockKMS) Encrypt(input *kms.EncryptInput) (*kms.EncryptOutput, error) {
	return m.EncryptFunc(input)
}

func (m *mockKMS) Decrypt(input *kms.DecryptInput) (*kms.DecryptOutput, error) {
	return m.DecryptFunc(input)
}

func TestEncrypt(t *testing.T) {
	c := &mockKMS{keyID: "someID"}
	kmsClient := &KMS{client: c, keyID: c.keyID}
	plaintext := []byte("plaintext")

	// Success encrypting
	c.EncryptFunc = func(in *kms.EncryptInput) (*kms.EncryptOutput, error) {
		in.Plaintext = plaintext
		ciphertextBlob := []byte(strings.ToUpper(string(in.Plaintext)))
		return &kms.EncryptOutput{
			CiphertextBlob: ciphertextBlob,
			KeyId:          aws.String(c.keyID),
		}, nil
	}
	blob, err := kmsClient.Encrypt(plaintext)
	assert.NoError(t, err)
	assert.EqualValues(t, []byte("PLAINTEXT"), blob)

	// Error encrypting
	c.EncryptFunc = func(in *kms.EncryptInput) (*kms.EncryptOutput, error) {
		return &kms.EncryptOutput{CiphertextBlob: nil}, fmt.Errorf("Encrypt Error")
	}
	blob, err = kmsClient.Encrypt(plaintext)
	assert.Nil(t, blob)
	assert.EqualError(t, err, "Encrypt Error")
}

func TestDecrypt(t *testing.T) {
	c := &mockKMS{keyID: "someID"}
	kmsClient := &KMS{client: c, keyID: c.keyID}
	ciphertextBlob := []byte("PLAINTEXT")

	// Success decrypting
	c.DecryptFunc = func(in *kms.DecryptInput) (*kms.DecryptOutput, error) {
		in.CiphertextBlob = ciphertextBlob
		plaintext := []byte(strings.ToLower(string(in.CiphertextBlob)))
		return &kms.DecryptOutput{
			Plaintext: plaintext,
			KeyId:     aws.String(c.keyID),
		}, nil
	}
	blob, err := kmsClient.Decrypt(ciphertextBlob)
	assert.NoError(t, err)
	assert.EqualValues(t, []byte("plaintext"), blob)

	// Error decrypting
	c.DecryptFunc = func(in *kms.DecryptInput) (*kms.DecryptOutput, error) {
		return &kms.DecryptOutput{Plaintext: nil}, fmt.Errorf("Decrypt Error")
	}
	blob, err = kmsClient.Decrypt(ciphertextBlob)
	assert.Nil(t, blob)
	assert.EqualError(t, err, "Decrypt Error")
}
