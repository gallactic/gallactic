package main

import (
	"fmt"
	"log"

	"github.com/abiosoft/ishell"
	"github.com/gallactic/gallactic/cmd"
	"github.com/gallactic/gallactic/keystore"
)

//
func New() func(c *ishell.Context) {
	return func(c *ishell.Context) {

		ks := c.Get("keystore").(*keystore.Keystore)
		valAddr := false

		if len(c.Args) > 0 {
			if c.Args[0] == "va" {
				valAddr = true
			}
		}

		// Encrypt key with passphrase.
		auth := cmd.PromptPassphrase("Passphrase: ", true)
		label := cmd.PromptInput("label: ")

		kd, err := ks.New(auth, label, valAddr)
		if err != nil {
			log.Fatalf("%v", err)
		}

		fmt.Println()
		cmd.PrintInfoMsg("Successfull created a key with this address: %v", kd.Address)
	}
}
