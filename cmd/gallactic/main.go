package main

import (
	"os"

	"github.com/gallactic/gallactic/cmd/gallactic/key"
	"github.com/jawher/mow.cli"
)

var title = `
               .__  .__                 __  .__
    _________  |  | |  | _____    _____/  |_|__| ____
   / ___\__  \ |  | |  | \__  \ _/ ___\   __\  |/ ___\
  / /_/  > __ \|  |_|  |__/ __ \\  \___|  | |  \  \___
  \___  (____  /____/____(____  /\___  >__| |__|\___  >
 /_____/     \/               \/     \/             \/
 `

func gallactic() *cli.Cli {
	app := cli.App("gallactic", "Gallactic blockchain node")

	app.Command("init", "Initialize the gallactic blockchain", Init())
	app.Command("start", "Start the gallactic blockchain", Start())
	app.Command("key", "Create gallactic key file for signing messages", func(k *cli.Cmd) {
		k.Command("generate", "Generate a new key", key.Generate())
		k.Command("inspect", "Inspect a key file", key.Inspect())
		k.Command("sign", "Sign a transaction or message with a key file", key.Sign())
		k.Command("verify", "Verify a signature", key.Verify())
		k.Command("change-auth", "Change the passphrase of a keyfile", key.ChangePassphrase())
	})
	app.Command("version", "Print the gallactic version", Version())
	return app
}

func main() {
	gallactic().Run(os.Args)
}
