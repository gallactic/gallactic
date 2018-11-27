package main

import (
	"fmt"

	"github.com/gallactic/gallactic/cmd"
	"github.com/gallactic/gallactic/version"
	"github.com/jawher/mow.cli"
)

//Version prints the version of the Gallactic node
func Version() func(c *cli.Cmd) {
	return func(c *cli.Cmd) {
		c.Before = func() { fmt.Println(title) }
		c.Action = func() {
			fmt.Println()
			cmd.PrintInfoMsg("Gallactic version: %v", version.Version)
		}
	}
}
