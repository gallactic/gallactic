package keystore

import (
	"io/ioutil"
	"log"

	"github.com/gallactic/gallactic/common"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/keystore/key"
)

type KeyData struct {
	Filename string
	Label    string
	Address  crypto.Address
	Key      *key.Key
}

type Keystore struct {
	Path string
	Keys []*KeyData
}

func Open(path string) *Keystore {
	ks := new(Keystore)
	ks.Path = path

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fname := path + file.Name()
		addr, lbl, v := key.VerifyFile(fname)

		if v == true {
			kd := &KeyData{fname, lbl, addr, nil}
			ks.Keys = append(ks.Keys, kd)
		}
	}

	return ks
}

func (ks *Keystore) New(auth, label string, valAddr bool) (*KeyData, error) {

	var keyObj *key.Key

	if valAddr {
		keyObj = key.GenValidatorKey()
	} else {
		keyObj = key.GenAccountKey()
	}

	// Encrypt key with passphrase.
	keyjson, err := key.EncryptKey(keyObj, auth, label)
	if err != nil {
		return nil, err
	}

	fname := ks.Path + keyObj.Address().String() + ".json"

	// Store the file to disk.
	if err := common.WriteFile(fname, keyjson); err != nil {
		return nil, err
	}

	kd := &KeyData{fname, label, keyObj.Address(), keyObj}
	ks.Keys = append(ks.Keys, kd)

	return kd, nil
}

func (ks *Keystore) Delete(addrress crypto.Address, auth string) error {
	/// check passsword. and them remove
	return nil
}

func (ks *Keystore) Update(addrress crypto.Address, oldAuth, newAuth, label string) error {
	/// update the key data
	return nil
}
