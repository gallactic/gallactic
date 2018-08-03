package main

import (
	"fmt"

	"github.com/gallactic/gallactic/version"
	"github.com/jawher/mow.cli"
)

func Version() func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			fmt.Println(version.Version)
		}
	}
}
