package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gallactic/gallactic/cmd"
	"github.com/gallactic/gallactic/common"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/config"
	proposal "github.com/gallactic/gallactic/core/proposal"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/keystore/key"
	"github.com/jawher/mow.cli"
)

// Init initializes the gallactic blockchain
func Init() func(c *cli.Cmd) {
	return func(c *cli.Cmd) {
		workingDir := c.String(cli.StringOpt{
			Name:  "w working-dir",
			Desc:  "Working directory to save configuration and genesis files",
			Value: ".",
		})
		chainName := c.String(cli.StringOpt{
			Name: "n chain-name",
			Desc: "A name for the blockchain",
		})

		c.Spec = "[-w=<Working directory>] [-n=<a name for the blockchain>]"
		c.LongDesc = "Initializing the working directory"
		c.Before = func() { fmt.Println(title) }
		c.Action = func() {

			// Check chain-name for genesis
			if *chainName == "" {
				*chainName = fmt.Sprintf("test-chain-%v", common.RandomHex(2))
			}

			path, _ := filepath.Abs(*workingDir)
			gen := makeGenesis(*workingDir, *chainName)
			conf := makeConfigfile()
			fmt.Println(os.Getwd())

			// save genesis file to file system
			genFile := path + "/genesis.json"
			if err := gen.SaveToFile(genFile); err != nil {
				cmd.PrintErrorMsg("Failed to write genesis file: %v", err)
				return
			}

			// save config file to file system
			confFile := path + "/config.toml"
			if err := conf.SaveToFile(confFile); err != nil {
				cmd.PrintErrorMsg("Failed to write config file: %v", err)
				return
			}

			fmt.Println()
			cmd.PrintSuccessMsg("A gallactic node is successfully initialized at %v", path)
		}
	}
}

//make genisis file while on initialize
func makeGenesis(workingDir string, chainName string) *proposal.Genesis {

	// create  accounts for genesis
	accs := make([]*account.Account, 4)
	for i := 0; i < len(accs); i++ {
		k := key.GenAccountKey()
		acc, _ := account.NewAccount(k.Address())
		acc.AddToBalance(10000000000000000000)
		acc.SetPermissions(permission.AllPermissions)
		accs[i] = acc
	}

	// create validator account for genesis
	k := key.GenValidatorKey()
	key.EncryptKeyFile(k, workingDir+"/validator_key.json", "", "")
	val, _ := validator.NewValidator(k.PublicKey(), 0)
	vals := []*validator.Validator{val}

	// create global account
	gAcc, _ := account.NewAccount(crypto.GlobalAddress)

	/* create genesis */
	gen := proposal.MakeGenesis(chainName, time.Now(), gAcc, accs, nil, vals)
	return gen

}

//make configuration file
func makeConfigfile() *config.Config {
	conf := config.DefaultConfig()
	return conf

}
