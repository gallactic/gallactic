package executors

import (
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"

	"github.com/hyperledger/burrow/logging"
)

type UnbondContext struct {
	Committing bool
	BC         *blockchain.Blockchain
	Cache      *state.Cache
	Logger     *logging.Logger
}

func (ctx *UnbondContext) Execute(txEnv *txs.Envelope) error {
	tx, ok := txEnv.Tx.(*tx.UnbondTx)
	if !ok {
		return e.Error(e.ErrInvalidTxType)
	}

	from, err := getInputValidator(ctx.Cache, tx.From())
	if err != nil {
		return err
	}

	to, err := getOutputAccount(ctx.Cache, tx.To())
	if err != nil {
		return err
	}
	if to == nil {
		return e.Error(e.ErrInvalidAddress)
	}

	// Good! Adjust accounts
	err = adjustInputValidator(from, tx.From())
	if err != nil {
		return err
	}

	err = adjustOutputAccount(to, tx.To())
	if err != nil {
		return err
	}

	/// Update account
	ctx.Cache.UpdateValidator(from)
	ctx.Cache.UpdateAccount(to)

	return nil
}
