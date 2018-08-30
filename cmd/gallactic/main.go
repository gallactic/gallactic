package main

import (
	"os"

	"github.com/jawher/mow.cli"
)

func gallactic() *cli.Cli {
	/// help string. ::TBD
	/// gallactic blockchain node with Hyperledger Burrow's EVM and Tendermint consensus engine
	app := cli.App("gallactic", "Start gallactic node")
    app.Command("init", "initialize the gallactic blockchain", Init())
	app.Command("start", "Start a Burrow node", Start())
	app.Command("version", "Version of the gallactic node", Version())

	return app
}

func main() {
	gallactic().Run(os.Args)
}
