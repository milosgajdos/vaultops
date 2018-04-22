package kms

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

// AwsKMS is AWS KMS client
type AwsKMS struct {
	client *kms.KMS
	keyID  string
}

// NewWithSession creates new AWS KMS client with session sess
// It returns error if the keyID is invalid AWS KMS key id
func NewWithSession(sess *session.Session, keyID string) (*AwsKMS, error) {
	if keyID == "" {
		return nil, fmt.Errorf("Invalid AWS KMS key ID: %v", keyID)
	}

	return &AwsKMS{kms.New(sess), keyID}, nil
}

// NewAWS returns new AWS KMS client
// It returns error if the keyID is invalid AWS KMS key id
func NewAWS(keyID string) (*AwsKMS, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return NewWithSession(sess, keyID)
}

// Encrypt encrypts plainText data and returns it
func (k *AwsKMS) Encrypt(plainText []byte) ([]byte, error) {
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
func (k *AwsKMS) Decrypt(cipherText []byte) ([]byte, error) {
	out, err := k.client.Decrypt(&kms.DecryptInput{
		CiphertextBlob: cipherText,
		EncryptionContext: map[string]*string{
			"Tool": aws.String("vaultops"),
		},
		GrantTokens: []*string{},
	})

	return out.Plaintext, err
}
