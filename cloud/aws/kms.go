package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

// KMS is AWS KMS client
type KMS struct {
	client interface {
		Encrypt(input *kms.EncryptInput) (*kms.EncryptOutput, error)
		Decrypt(input *kms.DecryptInput) (*kms.DecryptOutput, error)
	}
	keyID string
}

// NewKMSWithSession creates new AWS KMS client with session sess
// It returns error if the keyID is invalid AWS KMS key id
func NewKMSWithSession(sess *session.Session, keyID string) (*KMS, error) {
	if keyID == "" {
		return nil, fmt.Errorf("Invalid AWS KMS key ID: %v", keyID)
	}

	return &KMS{kms.New(sess), keyID}, nil
}

// NewKMS returns new AWS KMS client
// It returns error if the keyID is invalid AWS KMS key id
func NewKMS(keyID string) (*KMS, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return NewKMSWithSession(sess, keyID)
}

// Encrypt encrypts plainText data and returns it
func (k *KMS) Encrypt(plainText []byte) ([]byte, error) {
	out, err := k.client.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String(k.keyID),
		Plaintext: plainText,
		EncryptionContext: map[string]*string{
			"Tool": aws.String("vaultops"),
		},
		GrantTokens: []*string{},
	})

	return out.CiphertextBlob, err
}

// Decrypt decrypts cipherText data and returns it
func (k *KMS) Decrypt(cipherText []byte) ([]byte, error) {
	out, err := k.client.Decrypt(&kms.DecryptInput{
		CiphertextBlob: cipherText,
		EncryptionContext: map[string]*string{
			"Tool": aws.String("vaultops"),
		},
		GrantTokens: []*string{},
	})

	return out.Plaintext, err
}
