package key

import (
	"fmt"
	"log"

	"github.com/gallactic/gallactic/common"
	"github.com/gallactic/gallactic/keystore/key"
	"github.com/jawher/mow.cli"
)

const defaultKeyfilePath = "/tmp/gtxkeys/"

// Generate creates a new account and stores the keyfile in the disk
func Generate() func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[ -t=<addressType>]"
		addresstype := cmd.StringOpt("t addressType", "ac", "address type is ac for accountAddress and va for validatorAddress")

		cmd.Action = func() {
			keyObj := new(key.Key)
			if *addresstype == "va" {
				keyObj = key.GenValidatorKey()
			} else {
				keyObj = key.GenAccountKey()
			}
			// Encrypt key with passphrase.
			passphrase := PromptPassphrase(true)
			keyjson, err := key.EncryptKey(keyObj, passphrase)
			if err != nil {
				log.Fatalf("Failed to Encrypt: %v", err)
			}
			keyfilepath := defaultKeyfilePath + keyObj.Address().String() + ".json"

			// Store the file to disk.
			if err := common.WriteFile(keyfilepath, keyjson); err != nil {
				log.Fatalf("%v", err)
			}
			Address := keyObj.Address().String()
			fmt.Println("Address:", Address)
		}
	}
}
