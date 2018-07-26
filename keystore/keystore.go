package keystore

import (
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/keystore/key"
)

type KeyStore interface {
	// Loads and decrypts the key from disk.
	GetKey(addr crypto.Address,  auth string) (*key.Key, error)
	// Writes and encrypts the key.
	StoreKey(k *key.Key, auth string) error
	// Joins filename with the key directory unless it is already absolute.
	///JoinPath(filename string) string
}


