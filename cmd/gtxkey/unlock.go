package main

import (
	"github.com/abiosoft/ishell"
	"github.com/gallactic/gallactic/cmd"
	"github.com/gallactic/gallactic/keystore"
)

func Unlock() func(c *ishell.Context) {
	return func(c *ishell.Context) {
		ks := c.Get("keystore").(*keystore.Keystore)

		input := cmd.PromptInput("Key in account number/address to unlock the account (press ENTER to cancel): ")
		auth := cmd.PromptPassphrase("Key in passphrase to unlock account("+input+") :", true)

		err := ks.Unlock(auth, input)
		if err == nil {
			cmd.PrintSuccessMsg("Successfully Unlocked account")
		} else {
			cmd.PrintErrorMsg("Failed to Unlock: %v", err)
		}
	}
}
