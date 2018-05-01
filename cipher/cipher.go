package cipher

// Cipher allows to encrypt and decrypt arbitrary data
type Cipher interface {
	// Encrypt encrypts data in io.Reader
	Encrypt([]byte) ([]byte, error)
	// Decrypt decrypts data in io.Reader
	Decrypt([]byte) ([]byte, error)
}
