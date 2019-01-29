package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
			Name:  "w working-dir",
			Desc:  "Working directory of the configuration and genesis files",
			Value: ".",
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

		c.Spec = "[-w=<working directory>] [-p=<validator's private key>] | [-k=<path to the key file>] [-a=<key file's password>]"
		c.LongDesc = "Starting the node"
		c.Before = func() { fmt.Println(title) }
		c.Action = func() {

			path, _ := filepath.Abs(*workingDir)
			var keyObj *key.Key
			switch {
			case *keyFile == "" && *privateKey == "":
				f := path + "/validator_key.json"
				if common.FileExists(f) {
					kj, err := key.DecryptKeyFile(f, "")
					if err != nil {
						cmd.PrintErrorMsg("Aborted! %v", err)
						return
					}
					keyObj = kj
				} else {
					// Creating KeyObject from Private Key
					kj, err := cmd.PromptPrivateKey("Please enter the privateKey for the validator: ", false)
					if err != nil {
						cmd.PrintErrorMsg("Aborted! %v", err)
						return
					}
					keyObj = kj
				}
			case *keyFile != "" && *keyFileAuth != "":
				//Creating KeyObject from keystore
				passphrase := *keyFileAuth
				kj, err := key.DecryptKeyFile(*keyFile, passphrase)
				if err != nil {
					cmd.PrintErrorMsg("Aborted! %v", err)
					return
				}
				keyObj = kj
			case *keyFile != "" && *keyFileAuth == "":
				//Creating KeyObject from keystore
				passphrase := cmd.PromptPassphrase("Passphrase: ", false)
				kj, err := key.DecryptKeyFile(*keyFile, passphrase)
				if err != nil {
					cmd.PrintErrorMsg("Aborted! %v", err)
					return
				}
				keyObj = kj
			case *privateKey != "":
				// Creating KeyObject from Private Key
				pv, err := crypto.PrivateKeyFromString(*privateKey)
				if err != nil {
					cmd.PrintErrorMsg("Aborted! %v", err)
					return
				}
				keyObj, _ = key.NewKey(pv.PublicKey().ValidatorAddress(), pv)
			}

			cmd.PrintInfoMsg("Validator address: %v", keyObj.Address())

			// change working directory
			if err := os.Chdir(path); err != nil {
				cmd.PrintErrorMsg("Unable to changes working directory. %v", err)
				return
			}
			configFile := "./config.toml"
			genesisFile := "./genesis.json"

			gen, err := proposal.LoadFromFile(genesisFile)
			if err != nil {
				cmd.PrintErrorMsg("Could not obtain genesis. %v", err)
				return
			}

			conf, err := config.LoadFromFile(configFile)
			if err != nil {
				cmd.PrintErrorMsg("Could not obtain config. %v", err)
				return
			}

			err = conf.Check()
			if err != nil {
				cmd.PrintErrorMsg("Config is invalid - %v", err)
				return
			}

			err = os.Setenv("MAINNET_URL", conf.SputnikVM.Web3Address)
			if err != nil {
				cmd.PrintErrorMsg("Failed to set environment variable: %v", err)
				return
			}
			cmd.PrintSuccessMsg("Gallactic successfully connected to Ethereum Network")

			cmd.PrintInfoMsg("You are running a gallactic block chain node version: %v. Welcome! ", version.Version)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			signer := crypto.NewValidatorSigner(keyObj.PrivateKey())
			kernel, err := core.NewKernel(ctx, gen, conf, signer)
			if err != nil {
				cmd.PrintErrorMsg("Could not create kernel. %v", err)
				return
			}

			err = kernel.Boot()
			if err != nil {
				cmd.PrintErrorMsg("Could not boot kernel. %v", err)
				return
			}

			kernel.WaitForShutdown()
		}
	}
}
