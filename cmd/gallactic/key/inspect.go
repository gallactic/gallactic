package key

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gallactic/gallactic/cmd"
	"github.com/gallactic/gallactic/keystore/key"
	"github.com/jawher/mow.cli"
)

//Inspect displays various information of the keyfile
func Inspect() func(c *cli.Cmd) {
	return func(c *cli.Cmd) {
		keyFile := c.String(cli.StringOpt{
			Name: "k keyfile",
			Desc: "Path to the encrypted key file",
		})
		showPrivate := c.Bool(cli.BoolOpt{
			Name: "e expose-private-key",
			Desc: "expose the private key in the output",
		})
		c.Spec = "[-k=<path to the key file>] [-e]"
		c.Before = func() { fmt.Println(title) }
		c.Action = func() {
			if *keyFile == "" {
				fmt.Println("Key file is not specified.")
				return
			}
			// Read key from file.
			keyjson, err := ioutil.ReadFile(*keyFile)
			if err != nil {
				log.Fatalf("Failed to read the keyfile at '%s': %v", *keyFile, err)
			}
			// Decrypt key with passphrase.
			passphrase := cmd.PromptPassphrase("Passphrase: ", false)
			keyObj, err := key.DecryptKey(keyjson, passphrase)
			if err != nil {
				log.Fatalf("Error decrypting key: %v", err)
			}
			fmt.Println("Address: ", keyObj.Address())
			fmt.Println("Public key: ", keyObj.PublicKey())
			if *showPrivate {
				fmt.Println("Private key: ", keyObj.PrivateKey())
			}
		}
	}
}
