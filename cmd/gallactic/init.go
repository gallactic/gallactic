package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/config"
	proposal "github.com/gallactic/gallactic/core/proposal"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/jawher/mow.cli"
)

type validatorKey struct {
	ValidatorPrivateKey crypto.PrivateKey `json:"ValidatorPrivateKey"`
}

//initialize the gallactic
func Init() func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		workingDirOpts := cmd.String(cli.StringOpt{
			Name: "w working-dir",
			Desc: "working directory to save configuration and genesis files",
		})

		ChainNameOpts := cmd.String(cli.StringOpt{
			Name: "n chain-name",
			Desc: "chainname for genesis block",
		})

		cmd.Spec = "[--working-dir=<Working directory to save the configuration files>] " + "[--chainname =<chainname for genesis block>]"
		cmd.Action = func() {
			workingDir := *workingDirOpts
			chainName := *ChainNameOpts
			if workingDir != "" {

				//create genesis file
				genesis, msg := makeGenesis(workingDir, chainName)
				//save genesis file to file system
				gen := genesis.Save(workingDir)
				//save config file to file system
				conf := config.SaveConfigFile(workingDir)

				fmt.Println("config.toml", conf)
				fmt.Println("genesis.json", gen)
				fmt.Println("pv_validator.json", msg)

			} else {
				fmt.Println("see 'gallactic init --help' please enter the working directory")
			}
		}

	}
}

//make genisis file while on initialize
func makeGenesis(workingDir string, chainName string) (*proposal.Genesis, string) {

	/* Generate accounts private public key */
	accPubkey, _ := crypto.GenerateKey(nil)
	acc1Pubkey, _ := crypto.GenerateKey(nil)

	/* Generate validator private public key */
	valPubKey, valPrivKey := crypto.GenerateKey(nil)
	valKey := validatorKey{ValidatorPrivateKey: valPrivKey}

	/*marshal  validator key*/
	valPrivateKey, valerr := json.Marshal(valKey)
	if valerr != nil {
		log.Fatalf("validator in genesis %s", (valerr))
	}

	/*check for working path */
	if workingDir == "" {
		workingDir = "/tmp/chain/"
	}

	fileDir := workingDir + "pv_validator.json"
	/* create the directory */
	if err := os.MkdirAll(filepath.Dir(fileDir), 0777); err != nil {
		log.Fatalf("could not create directory %s", filepath.Dir(fileDir))
	}
	/* write  validiator private key to file */
	if err := ioutil.WriteFile(fileDir, valPrivateKey, 0600); err != nil {
		log.Fatalf("failed to write genesisfile to %s: %v", fileDir, err)
	}
	msg := "created at" + " " + fileDir

	/* create the address from public_key */
	address1 := accPubkey.AccountAddress()
	address2 := acc1Pubkey.AccountAddress()

	/* create  accounts for genesis */
	acc1, err1 := account.NewAccount(address1)
	if err1 != nil {
		log.Fatalf("Account1 in genesis %s", (err1))
	}
	acc2, err2 := account.NewAccount(address2)
	if err2 != nil {
		log.Fatalf("Account2 in genesis %s", (err2))
	}

	/*create validator account for genesis  */
	validtorAcc, valerror := validator.NewValidator(valPubKey, 100)
	if valerror != nil {
		log.Fatalf("validatorAcc %s", (valerror))
	}

	/* create global account*/
	gAcc, _ := account.NewAccount(crypto.GlobalAddress)
	/*create account list to generate genesis file */
	accounts := []*account.Account{acc1, acc2}
	/* create validator account to generate genesis file*/
	validators := []*validator.Validator{validtorAcc}

	//check chain-name for genesis
	gChainName := " "
	if gChainName != " " {
		gChainName = chainName
	} else {
		gChainName = "Genesis-001"
	}

	/* create genesis */
	gene := proposal.MakeGenesis(gChainName, time.Now(), gAcc, accounts, nil, validators)
	return gene, msg

}
