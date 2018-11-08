package key

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gallactic/gallactic/cmd"
	"github.com/gallactic/gallactic/common"
	"github.com/gallactic/gallactic/keystore/key"
	"github.com/jawher/mow.cli"
)

// ChangePassphrase changes the passphrase of the key file
func ChangePassphrase() func(c *cli.Cmd) {
	return func(c *cli.Cmd) {
		keyFile := c.String(cli.StringOpt{
			Name: "k keyfile",
			Desc: "Path to the encrypted key file",
		})

		c.Spec = "[-k=<path to the key file>]"
		c.Before = func() { fmt.Println(title) }
		c.Action = func() {
			if keyFile == nil {
				fmt.Println("Key file is not specified.")
				return
			}
			//Read the key from the keyfile
			keyjson, err := ioutil.ReadFile(*keyFile)
			if err != nil {
				log.Fatalf("Failed to read the keyfile at '%s': %v", *keyFile, err)
			}
			// Decrypt key with passphrase.
			passphrase := cmd.PromptPassphrase("Old passphrase: ", false)
			keyObj, err := key.DecryptKey(keyjson, passphrase)
			if err != nil {
				log.Fatalf("Password does not match: %v", err)
			}
			//Prompt for the new passphrase
			passphrase = cmd.PromptPassphrase("New passphrase: ", true)
			//Prompt for the label
			label := cmd.PromptInput("Label: ")
			// Encrypt key with passphrase.
			keyjson, err = key.EncryptKey(keyObj, passphrase, label)
			if err != nil {
				log.Fatalf("Failed to Encrypt: %v", err)
			}
			// Store the file to disk.
			if err := common.WriteFile(*keyFile, keyjson); err != nil {
				log.Fatalf("%v", err)
			}
			fmt.Println("Password changed successfully")
		}
	}
}
