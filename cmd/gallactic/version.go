package main

import (
	"fmt"

	"github.com/gallactic/gallactic/version"
	"github.com/jawher/mow.cli"
)

//Version prints the version of the Gallactic node
func Version() func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			fmt.Println(version.Version)
		}
	}
}
