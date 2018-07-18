package validator

import (
	"encoding/hex"

	"github.com/gallactic/gallactic/crypto"
)

type privateValidator struct {
	address    crypto.Address
	publicKey  crypto.PublicKey
	privateKey crypto.PrivateKey
}

func NewPrivateValidator(file, passphrase string) (crypto.Signer, error) {
	/*
		key, err := keystore.DecryptKeyFile(file, passphrase)
		if err != nil {
			return nil, err
		}


		return &privateValidator{
			address:    key.Address(),
			publicKey:  key.PublicKey(),
			privateKey: key.PrivateKey(),
		}, nil
	*/
	bs, _ := hex.DecodeString("6018F8B9C6EDB3F51FA847E2AADBCE42EE165658642EFF0B302FEA3343B21B83D67A7F69CFECFBACFA45046942191B310DC4FF1F9E8BF71DE565949FC72AF373")
	pv, _ := crypto.PrivateKeyFromRawBytes(bs)

	return &privateValidator{
		address:    pv.PublicKey().AccountAddress(),
		publicKey:  pv.PublicKey(),
		privateKey: pv,
	}, nil
}

func (pv *privateValidator) Address() crypto.Address {
	return pv.address
}

func (pv *privateValidator) PublicKey() crypto.PublicKey {
	return pv.privateKey.PublicKey()
}

func (pv *privateValidator) PrivateKey() crypto.PrivateKey {
	return pv.privateKey
}

func (pv *privateValidator) Sign(msg []byte) (crypto.Signature, error) {
	return pv.privateKey.Sign(msg)
}
