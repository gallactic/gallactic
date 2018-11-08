package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gallactic/gallactic/common"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/config"
	proposal "github.com/gallactic/gallactic/core/proposal"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/keystore/key"
	"github.com/jawher/mow.cli"
)

//initialize the gallactic
func Init() func(c *cli.Cmd) {
	return func(c *cli.Cmd) {
		workingDir := c.String(cli.StringOpt{
			Name: "w working-dir",
			Desc: "Working directory to save configuration and genesis files",
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

			// Check for working path
			if *workingDir == "" {
				*workingDir = "/tmp/" + *chainName
			}

			gen := makeGenesis(*workingDir, *chainName)
			conf := makeConfigfile()

			// save genesis file to file system
			if err := gen.SaveToFile(*workingDir + "/genesis.json"); err != nil {
				log.Fatalf("%v", err)
			}

			// save config file to file system
			if err := conf.SaveToFile(*workingDir + "/config.toml"); err != nil {
				log.Fatalf("%v", err)
			}

			log.Printf("A gallactic node is successfully initialized at %v.", *workingDir)
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

//make configuratin file
func makeConfigfile() *config.Config {
	conf := config.DefaultConfig()
	conf.Tendermint.ListenAddress = "127.0.0.1:46656"
	conf.Tendermint.Moniker = "moniker"
	conf.Tendermint.TendermintRoot = "tendermint"
	return conf

}
