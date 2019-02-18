package executors

import (
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	e "github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"
)

type BondContext struct {
	Committing bool
	BC         *blockchain.Blockchain
	Cache      *state.Cache
}

func (ctx *BondContext) Execute(txEnv *txs.Envelope, txRec *txs.Receipt) error {
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
		to, err = validator.NewValidator(
			tx.PublicKey(),
			ctx.BC.LastBlockHeight())

		if err != nil {
			return err
		}
	}

	// Good! Adjust account and validator
	err = adjustInputAccount(from, tx.From())
	if err != nil {
		return err
	}

	err = adjustOutputValidator(to, tx.To())
	if err != nil {
		return err
	}

	/// Update state cache
	ctx.Cache.UpdateAccount(from)
	ctx.Cache.UpdateValidator(to)

	return nil
}
