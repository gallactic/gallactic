package main

import (
	"context"
	"fmt"
	"log"
	"os"

	gtxkey "github.com/gallactic/gallactic/cmd/gallactic/key"
	"github.com/gallactic/gallactic/core"
	"github.com/gallactic/gallactic/core/config"
	"github.com/gallactic/gallactic/core/proposal"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/keystore/key"
	"github.com/jawher/mow.cli"
)

//Start starts the gallactic node
func Start() func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {

		workingDirOpt := cmd.String(cli.StringOpt{
			Name: "w working-dir",
			Desc: "working directory of the configuration files",
		})

		keystoreOpt := cmd.String(cli.StringOpt{
			Name: "k keyPath",
			Desc: "path to the key file",
		})

		privatekeyOpt := cmd.Bool(cli.BoolOpt{
			Name: "p privatekey",
			Desc: "private key of the account",
		})

		cmd.Spec = "[--working-dir=<Working directory of the configuration files>] " +
			"[--keyPath=<path to the key file>]" +
			"[--privatekey]"

		cmd.LongDesc = "Starting the node"
		cmd.Before = func() { fmt.Println(ascii) }
		cmd.Action = func() {
			fmt.Println("\n\n\nYou are running a gallactic blockchian node version: ", version.Version, ". Welcome!")
			workingDir := *workingDirOpt
			if workingDir != "" {
				if *privatekeyOpt || *keystoreOpt != "" {
					keyObj := new(key.Key)

					if *privatekeyOpt {
						fmt.Println("you are going to run gallactic blockchain ! Cheers!!")
						// Creating KeyObject from Private Key
						kj, err := PromptPrivateKey(*privatekeyOpt)
						if err != nil {
							log.Fatalf("Aborted: %v", err)
						}
						keyObj = kj
					}

					if *keystoreOpt != "" {
						//Creating KeyObject from keystore
						passphrase := promptPassphrase(true)
						kj, err := key.DecryptKeyFile(*keystoreOpt, passphrase)
						if err != nil {
							log.Fatalf("could not decrypt file: %v", err)
						}
						keyObj = kj
					}

					fmt.Println("", keyObj.Address().String())
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

				} else {
					fmt.Println("see 'gallactic start --help ' list available commands to start gallactic node")
				}
			} else {
				fmt.Println("see 'gallactic start --help ' list available commands to start gallactic node")
			}
		}
	}
}

func CreateKey(pv crypto.PrivateKey) *key.Key {
	addr := pv.PublicKey().ValidatorAddress()
	return key.NewKey(addr, pv)
}
