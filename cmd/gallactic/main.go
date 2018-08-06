package main

import (
	"os"

	"github.com/jawher/mow.cli"
)

func gallactic() *cli.Cli {
	/// help string. ::TBD
	/// gallactic blockchain node with Hyperledger Burrow's EVM and Tendermint consensus engine
	app := cli.App("gallactic", "Gallactic blockchain node")

	app.Command("start", "start the gallactic blockchain", Start())
	app.Command("version", "print the gallactic version", Version())

	return app
}

func main() {
	gallactic().Run(os.Args)
}
