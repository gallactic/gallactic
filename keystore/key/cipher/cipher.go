package cipher

import "github.com/gallactic/gallactic/crypto"

const (
	CipherNone   = "none"
	CipherAesCtr = "aes128_ctr"
)

type Cipher interface {
	CipherType() string
	Decrypt(auth string) (crypto.PrivateKey, error)
	Encrypt(pv crypto.PrivateKey, auth string) ([]byte,error)
}

func New(cipherType string) Cipher {
	switch cipherType {
	case CipherNone:
		return &plainKeyJSON{}
	case CipherAesCtr:
		return &encryptedKeyJSONV3{}

	}
	return nil
}
