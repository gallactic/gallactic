package main

import (
	"fmt"
	"log"

	"github.com/gallactic/gallactic/version"
	"github.com/jawher/mow.cli"
)

//Version prints the version of the Gallactic node
func Version() func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Before = func() { fmt.Println(ascii) }
		cmd.Action = func() {
			log.Printf("Version: " + version.Version)
		}
	}
}
