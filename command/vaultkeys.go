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
func (vk *VaultKeys) Write(s store.Store, c cipher.Cipher) (int, error) {
	v := &VaultKeys{vk.RootToken, vk.MasterKeys}
	// encode vault keys into json
	data, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}

	if c == nil {
		return s.Write(data)
	}

	encData, err := c.Encrypt(data)
	if err != nil {
		return 0, err
	}

	return s.Write(encData)
}

// Read reads vault keys from store and returns them decrypted
// It modifies the keys of the receiver
func (vk *VaultKeys) Read(s store.Store, c cipher.Cipher) error {
	data, err := ioutil.ReadAll(s)
	if err != nil {
		return err
	}

	keys := data
	if c != nil {
		keys, err = c.Decrypt(data)
		if err != nil {
			return err
		}
	}

	v := new(VaultKeys)
	if err := json.Unmarshal(keys, v); err != nil {
		return err
	}
	vk.RootToken = v.RootToken
	vk.MasterKeys = v.MasterKeys

	return nil
}
