package key

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gallactic/gallactic/common"
	"github.com/gallactic/gallactic/keystore/key"
	"github.com/jawher/mow.cli"
)

//ChangePassphrase changes the passphrase of the key file
func ChangePassphrase() func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		keyfile := cmd.String(cli.StringArg{
			Name: "KEYFILE",
			Desc: "KEYFILE of the account",
		})
		cmd.Spec = "KEYFILE"

		cmd.Action = func() {
			keyfilepath := defaultKeyfilePath + *keyfile
			//Read the key from the keyfile
			keyjson, err := ioutil.ReadFile(keyfilepath)
			if err != nil {
				log.Fatalf("Failed to read the keyfile at '%s': %v", keyfilepath, err)
			}
			// Decrypt key with passphrase.
			passphrase := oldPassphrase()
			keyObj, err := key.DecryptKey(keyjson, passphrase)
			if err != nil {
				log.Fatalf("Password does not match: %v", err)
			}
			//Prompt for the new passphrase
			passphrase = PromptPassphrase(true)
			// Encrypt key with passphrase.
			keyjson, err = key.EncryptKey(keyObj, passphrase)
			if err != nil {
				log.Fatalf("Failed to Encrypt: %v", err)
			}
			keyfilepath = defaultKeyfilePath + keyObj.Address().String() + ".json"
			// Store the file to disk.
			if err := common.WriteFile(keyfilepath, keyjson); err != nil {
				log.Fatalf("%v", err)
			}
			fmt.Println("Password changed successfully")
		}
	}
}
