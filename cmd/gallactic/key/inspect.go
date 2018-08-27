package key

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gallactic/gallactic/keystore/key"
	"github.com/jawher/mow.cli"
)

//Inspect displays various information of the keyfile
func Inspect() func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		keyfile := cmd.String(cli.StringArg{
			Name: "KEYFILE",
			Desc: "KEYFILE of the account",
		})
		showPrivate := cmd.Bool(cli.BoolOpt{
			Name: "p private",
			Desc: "include the private key in the output",
		})
		cmd.Spec = "KEYFILE [--private]"

		cmd.Action = func() {
			keyfilepath := defaultKeyfilePath + *keyfile //TODO include custom path as well
			// Read key from file.
			keyjson, err := ioutil.ReadFile(keyfilepath)
			if err != nil {
				log.Fatalf("Failed to read the keyfile at '%s': %v", keyfilepath, err)
			}
			// Decrypt key with passphrase.
			passphrase := PromptPassphrase(true)
			keyObj, err := key.DecryptKey(keyjson, passphrase)
			if err != nil {
				log.Fatalf("Error decrypting key: %v", err)
			}
			fmt.Println("Address: ", keyObj.Address().String())
			fmt.Println("Public key: ", keyObj.PublicKey().String())
			if *showPrivate {
				fmt.Println("Private key: ", keyObj.PublicKey().String())
			}
		}
	}
}
