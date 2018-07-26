package key

import (
	"encoding/json"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/keystore/key/cipher"
)

type keyJSON struct {
	Address crypto.Address  `json:"address"`
	Cipher  string          `json:"cipher"`
	Crypto  json.RawMessage `json:"crypto"`
}

// DecryptKeyFile returns an instance of Key object
func DecryptKeyFile(file, auth string) (*Key, error) {

	/// read the file
	var data []byte //// TONIYA

	kj := new(keyJSON)
	if err := json.Unmarshal(data, kj); err != nil {
		return nil, err
	}

	c := cipher.New(kj.Cipher)
	if err := json.Unmarshal(kj.Crypto, c); err != nil {
		return nil, err
	}
	///

	pv, err := c.Decrypt(auth)
	if err != nil {
		return nil, err
	}

	key := NewKey(kj.Address, pv)
	return key, nil
}

// EncryptKeyFile encrypts a key and return encrypted byte array
func EncryptKeyFile(key *Key, auth, cipherType, file string) error {
	c := cipher.New(cipherType)
	bs, err := c.Encrypt(key.PrivateKey(), auth)
	if err != nil {
		return err
	}

	kj := &keyJSON{
		Address: key.Address(),
		Cipher:  cipherType,
		Crypto:  bs,
	}
	bs, err = json.Marshal(kj)
	if err != nil {
		return err
	}
	//// save to file

	return nil
}
