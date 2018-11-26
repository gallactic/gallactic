package key

import (
	"fmt"

	"github.com/gallactic/gallactic/cmd"
	"github.com/gallactic/gallactic/common"
	"github.com/gallactic/gallactic/keystore/key"
	"github.com/jawher/mow.cli"
)

var title = `
              .__  .__                 __  .__          __
   _________  |  | |  | _____    _____/  |_|__| ____   |  | __ ____ ___.__.
  / ___\__  \ |  | |  | \__  \ _/ ___\   __\  |/ ___\  |  |/ // __ <   |  |
 / /_/  > __ \|  |_|  |__/ __ \\  \___|  | |  \  \___  |    <\  ___/\___  |
 \___  (____  /____/____(____  /\___  >__| |__|\___  > |__|_ \\___  > ____|
/_____/     \/               \/     \/             \/       \/    \/\/
`

// Generate creates a new account and stores the keyfile in the disk
func Generate() func(c *cli.Cmd) {
	return func(c *cli.Cmd) {
		addressType := c.String(cli.StringOpt{
			Name:  "t type",
			Desc:  "Use ac for the 'account address' and va for the 'validator address'",
			Value: "ac",
		})

		c.Spec = "[-t=<account type>]"
		c.Before = func() { fmt.Println(title) }
		c.Action = func() {
			keyObj := new(key.Key)
			if *addressType == "va" {
				keyObj = key.GenValidatorKey()
			} else {
				keyObj = key.GenAccountKey()
			}
			passphrase := cmd.PromptPassphrase("Passphrase: ", true)
			label := cmd.PromptInput("Label: ")
			keyjson, err := key.EncryptKey(keyObj, passphrase, label)
			if err != nil {
				cmd.PrintErrorMsg("Failed to encrypt: %v", err)
				return
			}
			keyfilepath := common.GallacticKeystoreDir() + keyObj.Address().String() + ".json"

			// Store the file to disk.
			if err := common.WriteFile(keyfilepath, keyjson); err != nil {
				cmd.PrintErrorMsg("Failed to write the key file: %v", err)
				return
			}

			fmt.Println()
			cmd.PrintInfoMsg("Key path: %v", keyfilepath)
			cmd.PrintInfoMsg("Address: %v", keyObj.Address())
			cmd.PrintInfoMsg("Public key: %v", keyObj.PublicKey())
		}
	}
}
