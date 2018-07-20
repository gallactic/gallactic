package keystore

import (
	"encoding/hex"
	"io/ioutil"

	"github.com/gallactic/gallactic/crypto"
)

type Key struct {
	data keyData
}

type keyData struct {
	Address    crypto.Address
	PublicKey  crypto.PublicKey
	PrivateKey crypto.PrivateKey
}

// DecryptKeyFile returns an instance of Key object
func DecryptKeyFile(file, passphrase string) (*Key, error) {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	pv, err := hex.DecodeString(string(bs))
	if err != nil {
		return nil, err
	}
	privKey, _ := crypto.PrivateKeyFromRawBytes(pv)

	return &Key{
		data: keyData{
			PrivateKey: privKey,
			PublicKey:  privKey.PublicKey(),
			Address:    privKey.PublicKey().ValidatorAddress(),
		},
	}, nil
}

func (k *Key) Address() crypto.Address {
	return k.data.Address
}

func (k *Key) PublicKey() crypto.PublicKey {
	return k.data.PublicKey
}

func (k *Key) PrivateKey() crypto.PrivateKey {
	return k.data.PrivateKey
}
