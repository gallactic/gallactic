package main

import (
	"fmt"
	config "github.com/gallactic/gallactic/core/config"
	proposal "github.com/gallactic/gallactic/core/proposal"
	"github.com/jawher/mow.cli"
)

func Init() func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		workingDirOpts := cmd.String(cli.StringOpt{
			Name: "w working-dir",
			Desc: "working directory to save configuration and genesis files",
		})

		ChainNameOpts := cmd.String(cli.StringOpt{
			Name: "c chainname",
			Desc: "chainname for genesis block",
		})

		cmd.Spec = "[--working-dir=<Working directory to save the configuration files>] " + "[--chainname =<chainname for genesis block>]"

		cmd.Action = func() {
			workingDir := *workingDirOpts
			chainName := *ChainNameOpts
			if workingDir != "" {
				//save config file to file system
				conf := config.SaveConfigFile(workingDir)
				//save genesis file to file system
				gen := proposal.SaveGenesisFile(workingDir, chainName)

				fmt.Println("config.toml", conf)
				fmt.Println("genesis.json", gen)

			} else {
				fmt.Println("see 'gallactic init --help' please enter the working directory")
			}
		}

	}
}
