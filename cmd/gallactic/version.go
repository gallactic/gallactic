package main

import (
	"fmt"

	"github.com/gallactic/gallactic/version"
	"github.com/jawher/mow.cli"
)

func Version() func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {

		versionOpt := cmd.String(cli.StringOpt{
			Name: "v version",
			Desc: "Gallactic version",
		})

		cmd.Spec = "[-v]"

		cmd.Action = func() {
			fmt.Println(version.Version, *versionOpt)
		}
	}
}
