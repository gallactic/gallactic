package cipher

import "github.com/gallactic/gallactic/crypto"

type plainKeyJSON struct {
	PrivateKey string `json:"privatekey"`
	Version    int    `json:"version"`
}

func (c *plainKeyJSON) CipherType() string {
	return CipherNone
}

func (c *plainKeyJSON) Decrypt(auth string) (crypto.PrivateKey, error) {
	return crypto.PrivateKey{}, nil
}

func (c *plainKeyJSON) Encrypt(pv crypto.PrivateKey, auth string) ([]byte, error) {
	return nil, nil
}
