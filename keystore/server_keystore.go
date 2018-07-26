package keystore

import (
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/keystore/key"
)

type serverKeyStore struct {
}

func NewServerKeyStore( /*params*/ ) KeyStore {
	return &serverKeyStore{}
}

func (ks *serverKeyStore) GetKey(addr crypto.Address, auth string) (*key.Key, error) {
	return nil, nil
}

func (ks *serverKeyStore) StoreKey(k *key.Key, auth string) error {
	return nil
}
