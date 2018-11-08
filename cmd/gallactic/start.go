package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gallactic/gallactic/cmd"
	"github.com/gallactic/gallactic/common"
	"github.com/gallactic/gallactic/core"
	"github.com/gallactic/gallactic/core/config"
	"github.com/gallactic/gallactic/core/proposal"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/keystore/key"
	"github.com/gallactic/gallactic/version"
	"github.com/jawher/mow.cli"
)

//Start starts the gallactic node
func Start() func(c *cli.Cmd) {
	return func(c *cli.Cmd) {

		workingDir := c.String(cli.StringOpt{
			Name: "w working-dir",
			Desc: "Working directory of the configuration and genesis files",
		})
		privateKey := c.String(cli.StringOpt{
			Name: "p privatekey",
			Desc: "Private key of the node's validator",
		})
		keyFile := c.String(cli.StringOpt{
			Name: "k keyfile",
			Desc: "Path to the encrypted key file contains validator's private key",
		})
		keyFileAuth := c.String(cli.StringOpt{
			Name: "a auth",
			Desc: "Key file's passphrase",
		})

		c.Spec = "[-w=<working directory>] [-p=<validator's private key>]"
		c.LongDesc = "Starting the node"
		c.Before = func() { fmt.Println(title) }
		c.Action = func() {

			if *workingDir == "" {
				fmt.Println("working directory is not specified.")
				fmt.Println("see 'gallactic start --help' for list of available commands to start gallactic node")
				return
			}

			fmt.Println("You are running a gallactic block chain node version: ", version.Version, ". Welcome!")
			keyObj := new(key.Key)

			switch {
			case *keyFile == "" && *privateKey == "":
				f := *workingDir + "/validator_key.json"
				if common.FileExists(f) {
					kj, err := key.DecryptKeyFile(f, "")
					if err != nil {
						log.Fatalf("Aborted: %v", err)
					}
					keyObj = kj
				} else {
					// Creating KeyObject from Private Key
					kj, err := cmd.PromptPrivateKey("Please enter the privateKey for the validator: ", false)
					if err != nil {
						log.Fatalf("Aborted: %v", err)
					}
					keyObj = kj
				}
			case *keyFile != "" && *keyFileAuth != "":
				//Creating KeyObject from keystore
				passphrase := *keyFileAuth
				kj, err := key.DecryptKeyFile(*keyFile, passphrase)
				if err != nil {
					log.Fatalf("Could not decrypt file: %v", err)
				}
				keyObj = kj
			case *keyFile != "" && *keyFileAuth == "":
				//Creating KeyObject from keystore
				passphrase := cmd.PromptPassphrase("Passphrase: ", false)
				kj, err := key.DecryptKeyFile(*keyFile, passphrase)
				if err != nil {
					log.Fatalf("Could not decrypt file: %v", err)
				}
				keyObj = kj
			case *privateKey != "":
				// Creating KeyObject from Private Key
				pv, err := crypto.PrivateKeyFromString(*privateKey)
				if err != nil {
					log.Fatalf("Could not decrypt file: %v", err)
				}
				keyObj, _ = key.NewKey(pv.PublicKey().ValidatorAddress(), pv)
			}

			fmt.Println("Validator address: ", keyObj.Address().String())

			// change working directory
			if err := os.Chdir(*workingDir); err != nil {
				log.Fatalf("Unable to changes working directory: %v", err)
			}
			configFile := "./config.toml"
			genesisFile := "./genesis.json"

			gen, err := proposal.LoadFromFile(genesisFile)
			if err != nil {
				log.Fatalf("Could not obtain genesis from file: %v", err)
			}

			conf, err := config.LoadFromFile(configFile)
			if err != nil {
				log.Fatalf("Could not obtain config from file: %v", err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			signer := crypto.NewValidatorSigner(keyObj.PrivateKey())
			kernel, err := core.NewKernel(ctx, gen, conf, signer)
			if err != nil {
				log.Fatalf("Could not create kernel: %v", err)
			}

			err = kernel.Boot()
			if err != nil {
				log.Fatalf("Could not boot kernel: %v", err)
			}

			kernel.WaitForShutdown()
		}
	}
}
