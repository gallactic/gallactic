package executors

import (
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"

	"github.com/hyperledger/burrow/logging"
)

type BondContext struct {
	Committing bool
	BC         *blockchain.Blockchain
	Cache      *state.Cache
	Logger     *logging.Logger
}

func (ctx *BondContext) Execute(txEnv *txs.Envelope) error {
	tx, ok := txEnv.Tx.(*tx.BondTx)
	if !ok {
		return e.Error(e.ErrInvalidTxType)
	}

	from, err := getInputAccount(ctx.Cache, tx.From(), permission.Bond)
	if err != nil {
		return err
	}

	to, err := getOutputValidator(ctx.Cache, tx.To())
	if err != nil {
		return err
	}

	if to == nil {
		to = validator.NewValidator(
			tx.PublicKey(),
			tx.To().Amount,
			ctx.BC.LastBlockHeight())
	} else {
		to.AddToStake(tx.To().Amount)
	}

	// Good! Adjust accounts
	err = adjustInputAccount(from, tx.From())
	if err != nil {
		return err
	}

	err = adjustOutputValidator(to, tx.To())
	if err != nil {
		return err
	}

	/// Update account
	ctx.Cache.UpdateAccount(from)
	ctx.Cache.UpdateValidator(to)

	return nil
}
