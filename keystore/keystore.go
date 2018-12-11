package keystore

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

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

func (ks *Keystore) Unlock(auth, indexOrAddress string) error {
	// get the keydata struct
	keyObject, err := ks.getKeyObject(indexOrAddress)
	if err != nil {
		return err
	}

	// do decrypt and return the data
	key, err := key.DecryptKeyFile(keyObject.Filename, auth)
	if err != nil {
		return err
	}

	keyObject.Key = key

	return nil
}

func (ks *Keystore) Lock(indexOrAddress string) error {
	keyObject, err := ks.getKeyObject(indexOrAddress)
	if err != nil {
		return err
	}
	keyObject.Key = nil
	return nil
}

func (ks *Keystore) Delete(address crypto.Address, auth string) error {
	/// check passsword. and them remove
	return nil
}

func (ks *Keystore) Update(address crypto.Address, oldAuth, newAuth, label string) error {
	/// update the key data
	return nil
}

func (ks *Keystore) getKeyObject(indexOrAddress string) (*KeyData, error) {
	index, err := strconv.Atoi(indexOrAddress)
	if err != nil {
		// if error, try check if a valid address
		addr, err := crypto.AddressFromString(indexOrAddress)
		if err != nil {
			return nil, fmt.Errorf("Error, string is not a valid index or address")
		}

		// try to find index of the address in keystore
		for i, e := range ks.Keys {
			if e.Address == addr {
				index = i
			}
		}
	} else {
		// reduce index by 1 as keystore listing always start from 1
		index = index - 1
	}

	// if index is out of range
	if index >= len(ks.Keys) || index <= 0 {
		return nil, fmt.Errorf("Index is out of range")
	}

	return ks.Keys[index], nil
}
