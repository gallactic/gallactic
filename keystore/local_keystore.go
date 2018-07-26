package keystore

import (
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/keystore/key"
)

type localKeyStore struct {
}

func NewLocalKeyStore( /*params*/ ) KeyStore {
	return &localKeyStore{}
}

func (ks *localKeyStore) GetKey(addr crypto.Address, auth string) (*key.Key, error) {
	return nil, nil
}

func (ks *localKeyStore) StoreKey(k *key.Key, auth string) error {
	return nil
}
