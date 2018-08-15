package main

import (
	"os"

	"github.com/jawher/mow.cli"
)

func gallactic() *cli.Cli {
	/// help string. ::TBD
	/// gallactic blockchain node with Hyperledger Burrow's EVM and Tendermint consensus engine
	app := cli.App("gallactic", "Gallactic blockchain node")
    app.Command("init", "initialize the gallactic blockchain", Init())
	app.Command("start", "start the gallactic blockchain", Start())
	app.Command("version", "print the gallactic version", Version())
	app.Command("gtxkey", "gallactic key manager", func(key *cli.Cmd) {
		key.Command("generate", "generate a new key", Generate())
		key.Command("inspect", "inspect a key file", Inspect())
		key.Command("signmessage", "inspect a key file", Sign())
		key.Command("verify", "verify a signature of a messsage", Verify())
	})
	return app
}

func main() {
	gallactic().Run(os.Args)
}
