package core

import "testing"

func TestBootThenShutdown(t *testing.T) {
}

/*
import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gallactic/gallactic/core/genesis"
	"github.com/gallactic/gallactic/core/types/permission"
	"github.com/gallactic/gallactic/core/consensus/tendermint"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/hyperledger/burrow/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmConfig "github.com/tendermint/tendermint/config"
	tmTypes "github.com/tendermint/tendermint/types"
)

const testDir = "./test_scratch/kernel_test"

func TestBootThenShutdown(t *testing.T) {
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.Chdir(testDir)
	tmConf := tmConfig.DefaultConfig()
	logger, _ := lifecycle.NewStdErrLogger()
	//logger := logging.NewNoopLogger()
	gen, privAccount := deterministicGenesisDoc()
	privValidator := tmValidator.NewPrivValidatorMemory(privAccount.PublicKey(), privAccount)
	assert.NoError(t, bootWaitBlocksShutdown(privValidator, gen, tmConf, logger, nil))
}

func TestBootShutdownResume(t *testing.T) {
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.Chdir(testDir)
	tmConf := tmConfig.DefaultConfig()
	logger, _ := lifecycle.NewStdErrLogger()
	//logger := logging.NewNoopLogger()

	genDoc1, privAccount := deterministicGenesisDoc()
	privValidator := tmValidator.NewPrivValidatorMemory(privAccount.PublicKey(), privAccount)

	i := int64(0)
	// asserts we get a consecutive run of blocks
	blockChecker := func(block *tmTypes.EventDataNewBlock) bool {
		assert.Equal(t, i+1, block.Block.Height)
		i++
		// stop every third block
		return i%3 != 0
	}
	// First run
	require.NoError(t, bootWaitBlocksShutdown(privValidator, genDoc1, tmConf, logger, blockChecker))
	// Resume and check we pick up where we left off
	require.NoError(t, bootWaitBlocksShutdown(privValidator, genDoc1, tmConf, logger, blockChecker))
	// Resuming with mismatched genesis should fail
	genDoc2 := genesis.MakeGenesisDoc("tm-chain2", genDoc1.GenesisTime(), genDoc1.GlobalPermissions(), genDoc1.Accounts(), genDoc1.Validators())
	assert.Error(t, bootWaitBlocksShutdown(privValidator, genDoc2, tmConf, logger, blockChecker))
}

func deterministicGenesisDoc() (*genesis.Genesis, acm.PrivateAccount) {
	privAccount := acm.GeneratePrivateAccountFromSecret("test-account")
	account := acm.NewAccount(privAccount.Address())
	account.AddToBalance(1000)
	accounts := []*acm.Account{account}
	validators := []*validator.Validator{validator.NewValidator(privAccount.PublicKey(), 1000, 0)}
	gen := genesis.MakeGenesisDoc("tm-chain", time.Now(), permission.ZeroPermissions, accounts, validators)

	return gen, privAccount
}

func bootWaitBlocksShutdown(privValidator tmTypes.PrivValidator, gen *genesis.Genesis,
	tmConf *tmConfig.Config, logger *logging.Logger,
	blockChecker func(block *tmTypes.EventDataNewBlock) (cont bool)) error {

	keyStore := keys.NewKeyStore(keys.DefaultKeysDir, false, logger)
	keyClient := keys.NewLocalKeyClient(keyStore, logging.NewNoopLogger())
	kern, err := NewKernel(context.Background(), keyClient, privValidator, gen, tmConf,
		rpc.DefaultRPCConfig(), keys.DefaultKeysConfig(), keyStore, nil, logger)
	if err != nil {
		return err
	}

	err = kern.Boot()
	if err != nil {
		return err
	}

	ch, err := tendermint.SubscribeNewBlock(context.Background(), kern.Emitter)
	if err != nil {
		return err
	}
	cont := true
	for cont {
		select {
		case <-time.After(2 * time.Second):
			if err != nil {
				return fmt.Errorf("timed out waiting for block")
			}
		case ednb := <-ch:
			if blockChecker == nil {
				cont = false
			} else {
				cont = blockChecker(ednb)
			}
		}
	}
	return kern.Shutdown(context.Background())
}
*/
