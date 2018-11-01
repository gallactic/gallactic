package main

import (
	"os"

	"github.com/gallactic/gallactic/cmd/gallactic/key"
	"github.com/jawher/mow.cli"
)

var title = `
   ____    _    _     _        _    ____ _____ ___ ____
  / ___|  / \  | |   | |      / \  / ___|_   _|_ _/ ___|
 | |  _  / _ \ | |   | |     / _ \| |     | |  | | |
 | |_| |/ ___ \| |___| |___ / ___ \ |___  | |  | | |___
  \____/_/   \_\_____|_____/_/   \_\____| |_| |___\____|

 `

func gallactic() *cli.Cli {
	app := cli.App("gallactic", "Gallactic blockchain node")

	app.Command("init", "initialize the gallactic blockchain", Init())
	app.Command("start", "start the gallactic blockchain", Start())
	app.Command("version", "print the gallactic version", Version())
	app.Command("key", "gallactic key manager", func(k *cli.Cmd) {
		k.Command("generate", "generate a new key", key.Generate())
		k.Command("inspect", "inspect a key file", key.Inspect())
		k.Command("signmessage", "inspect a key file", key.Sign())
		k.Command("verify", "verify a signature of a messsage", key.Verify())
		k.Command("changeauth", "change the passphrase of the keyfile", key.ChangePassphrase())
	})
	return app
}

func main() {
	gallactic().Run(os.Args)
}
