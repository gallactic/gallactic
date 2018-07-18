package test

import (
	"crypto"
	"os"
	"testing"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/genesis"
	"github.com/gallactic/gallactic/core/state"
	"github.com/hyperledger/burrow/logging"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var chainID string
var accountPool map[string]*account.Account
var signerPool map[string]crypto.Signer /// private keys
var genesisDoc *genesis.Genesis
var bc1 *blockchain.Blockchain
var bc1Db dbm.DB
var bc1State *state.State
var nopLogger *logging.Logger

/*
var checker state.BatchExecutor
var committer state.BatchCommitter
*/
func TestMain(m *testing.M) {
	nopLogger = logging.NewNoopLogger()

	setupAccountPool(m)
	setupGenesis(m)
	//tcore.setupBlockchain(m)
	//tcore.setupBatchChecker(m)

	exitCode := m.Run()

	os.Exit(exitCode)
}
