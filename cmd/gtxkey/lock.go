package main

import (
	"fmt"
	"log"

	"github.com/abiosoft/ishell"
	"github.com/gallactic/gallactic/cmd"
	"github.com/gallactic/gallactic/keystore"
)

func Lock() func(c *ishell.Context) {
	return func(c *ishell.Context) {
		// open keystore
		ks := c.Get("keystore").(*keystore.Keystore)

		input := cmd.PromptInput("Key in account number/address to lock the account (press ENTER to cancel): ")

		err := ks.Lock(input)
		if err == nil {
			fmt.Println("Successfully Locked account")
		} else {
			log.Fatal(err)
		}
	}
}
