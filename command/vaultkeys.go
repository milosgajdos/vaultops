package command

import (
	"encoding/json"
	"io/ioutil"

	"github.com/milosgajdos/vaultops/cipher"
	"github.com/milosgajdos/vaultops/store"
)

// VaultKeys stores vault root token and master keys
type VaultKeys struct {
	// RootToken is vault root token
	RootToken string `json:"root_token,omitempty"`
	// MasterKeys are vault master keys used to unseal vault servers
	MasterKeys []string `json:"master_keys,omitempty"`
}

// Write writes vault keys in store and encrypts them with cipher c
func (v *VaultKeys) Write(s store.Store, c cipher.Cipher) (int, error) {
	k := &VaultKeys{v.RootToken, v.MasterKeys}
	// encode vault keys into json
	data, err := json.Marshal(k)
	if err != nil {
		return 0, err
	}

	if c == nil {
		return s.Write(data)
	}

	enc, err := c.Encrypt(data)
	if err != nil {
		return 0, err
	}

	return s.Write(enc)
}

// Read reads vault keys from store, decrypts them and stores them in
// its fields i.e. it modifies the keys stored in the receiver.
func (v *VaultKeys) Read(s store.Store, c cipher.Cipher) (int, error) {
	data, err := ioutil.ReadAll(s)
	if err != nil {
		return 0, err
	}

	keys := data
	if c != nil {
		keys, err = c.Decrypt(data)
		if err != nil {
			return 0, err
		}
	}

	k := new(VaultKeys)
	if err := json.Unmarshal(keys, k); err != nil {
		return 0, err
	}
	v.RootToken = k.RootToken
	v.MasterKeys = k.MasterKeys

	return len(data), nil
}
