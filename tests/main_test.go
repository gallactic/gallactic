package tests

import (
	"os"
	"testing"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/execution"
	"github.com/gallactic/gallactic/core/proposal"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/hyperledger/burrow/logging"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var tChainID string
var tAccounts map[string]*account.Account
var tValidators map[string]*validator.Validator
var tSigners map[string]crypto.Signer /// private keys
var tGenesis *proposal.Genesis
var tBC *blockchain.Blockchain
var tDB dbm.DB
var tState *state.State
var tLogger *logging.Logger
var tChecker execution.BatchExecutor
var tCommitter execution.BatchCommitter

func TestMain(m *testing.M) {
	tLogger = logging.NewNoopLogger()

	setupAccountPool(m)
	setupGenesis(m)
	setupBlockchain(m)

	exitCode := m.Run()

	os.Exit(exitCode)
}
