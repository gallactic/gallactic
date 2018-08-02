package main

import (
	"context"
	"log"
	"os"

	"github.com/gallactic/gallactic/core"
	"github.com/gallactic/gallactic/core/config"
	"github.com/gallactic/gallactic/core/proposal"
	"github.com/jawher/mow.cli"
)

func Start() func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {

		workingDirOpt := cmd.String(cli.StringOpt{
			Name: "w working-dir",
			Desc: "Working directory of the configuration files",
		})

		/*
			passphraseOpt := cmd.String(cli.StringOpt{
				Name: "p passphrase",
				Desc: "The passphrase of the validator keystore",
			})
		*/

		cmd.Spec = "[--working-dir=<Working directory of the configuration files>] " +
			"[--passphrase=<passphrase of the validator keystore>]"

		cmd.Action = func() {
			workingDir := *workingDirOpt

			// change working directory
			if err := os.Chdir(workingDir); err != nil {
				log.Fatalf("Unable to changes working directory: %v", err)
			}

			configFile := "./config.toml"
			genesisFile := "./genesis.json"

			gen, err := proposal.LoadFromFile(genesisFile)
			if err != nil {
				log.Fatalf("could not obtain genesis from file: %v", err)
			}

			conf, err := config.LoadFromFile(configFile)
			if err != nil {
				log.Fatalf("could not obtain config from file: %v", err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			kernel, err := core.NewKernel(ctx, gen, conf, nil)
			if err != nil {
				log.Fatalf("could not create kernel: %v", err)
			}

			err = kernel.Boot()
			if err != nil {
				log.Fatalf("could not boot kernel: %v", err)
			}

			kernel.WaitForShutdown()
		}
	}
}
