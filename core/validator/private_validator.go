package validator

import (
	"encoding/hex"

	"github.com/gallactic/gallactic/crypto"
)

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
	signer := crypto.NewValidatorSigner(pv)

	return signer, nil
}
