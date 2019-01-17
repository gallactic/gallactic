package executors

import (
	"errors"

	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/state"
	e "github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"

	"github.com/hyperledger/burrow/logging"
	tmRPC "github.com/tendermint/tendermint/rpc/core"
)

type SortitionContext struct {
	Committing bool
	BC         *blockchain.Blockchain
	Cache      *state.Cache
	Logger     *logging.Logger
}

func (ctx *SortitionContext) Execute(txEnv *txs.Envelope, txRec *txs.Receipt) error {
	tx, ok := txEnv.Tx.(*tx.SortitionTx)
	if !ok {
		return e.Error(e.ErrInvalidTxType)
	}

	sortitionThreshold := uint64(ctx.BC.MaximumPower())

	/// Check if sortition tx belongs to next height, otherwise discard it
	curBlockHeight := ctx.BC.LastBlockHeight() + 1
	if tx.Height() < curBlockHeight-sortitionThreshold {
		return errors.New("Invalid block height")
	}

	val, err := getInputValidator(ctx.Cache, tx.Validator())
	if err != nil {
		return err
	}

	/// validators should not submit sortition before the threshold
	if tx.Height() < val.BondingHeight()+sortitionThreshold {
		return errors.New("Invalid block height")
	}

	isInSet := ctx.BC.ValidatorSet().Contains(tx.Validator().Address)
	if isInSet {
		return errors.New("This validator is already in set")
	}

	/// Verify the sortition
	var blockHeight = int64(tx.Height())
	result, err := tmRPC.Block(&blockHeight)
	if err != nil {
		return err
	}

	blockHash := result.Block.Hash()
	isValid := ctx.BC.VerifySortition(blockHash, txEnv.Signatories[0].PublicKey, tx.Index(), tx.Proof())
	if !isValid {
		return errors.New("Sortition transaction is invalid")
	}

	// Good! Adjust validator
	err = adjustInputValidator(val, tx.Validator())
	if err != nil {
		return err
	}

	/// Update state cache
	ctx.Cache.AddToSet(val)

	return nil
}
