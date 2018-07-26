package cipher

import "github.com/gallactic/gallactic/crypto"

/// TODO: rename it later
type encryptedKeyJSONV3 struct {
	Crypto  cryptoJSON `json:"crypto"`
	Version int        `json:"version"`
}

type cryptoJSON struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams cipherparamsJSON       `json:"cipherparams"`
	KDF          string                 `json:"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

type cipherparamsJSON struct {
	IV string `json:"iv"`
}

func (c *encryptedKeyJSONV3) CipherType() string {
	return CipherAesCtr
}
func (c *encryptedKeyJSONV3) Decrypt(auth string) (crypto.PrivateKey, error) {
	return crypto.PrivateKey{}, nil
}

func (c *encryptedKeyJSONV3) Encrypt(pv crypto.PrivateKey, auth string) ([]byte, error) {
	return nil, nil
}
