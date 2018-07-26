package key

import (
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

func GenAccountKey() *Key {
	pv := crypto.GeneratePrivateKey(nil)
	return &Key{
		data: keyData{
			PrivateKey: pv,
			PublicKey:  pv.PublicKey(),
			Address:    pv.PublicKey().AccountAddress(),
		},
	}
}

func GenValidatorKey() *Key {
	pv := crypto.GeneratePrivateKey(nil)
	return &Key{
		data: keyData{
			PrivateKey: pv,
			PublicKey:  pv.PublicKey(),
			Address:    pv.PublicKey().ValidatorAddress(),
		},
	}
}

func NewKey(addr crypto.Address, pv crypto.PrivateKey) *Key {
	return &Key{
		data: keyData{
			PrivateKey: pv,
			PublicKey:  pv.PublicKey(),
			Address:    addr,
		},
	}
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
