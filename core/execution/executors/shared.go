package executors

import (
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs/tx"
)

func getInputAccount(ch *state.Cache, in tx.TxInput, req account.Permissions) (*account.Account, error) {
	// Check Input basic
	if err := in.ValidateBasic(); err != nil {
		return nil, err
	}

	acc := ch.GetAccount(in.Address)
	if acc == nil {
		return nil, e.Errorf(e.ErrInvalidAddress, "Account %s doesn't exist", in.Address)
	}

	// Check sequences
	if acc.Sequence()+1 != uint64(in.Sequence) {
		return nil, e.Errorf(e.ErrInvalidSequence, "%s has set sequence to %s. It should be %s", in.Address, in.Sequence, acc.Sequence()+1)
	}

	// Check amount
	if acc.Balance() < uint64(in.Amount) {
		return nil, e.Error(e.ErrInsufficientFunds)
	}

	if !ch.HasPermissions(acc, req) {
		return nil, e.Errorf(e.ErrPermDenied, "%s has %s but needs %s", in.Address, acc.Permissions(), permission.Send)
	}

	return acc, nil
}

func getOutputAccount(ch *state.Cache, out tx.TxOutput) (*account.Account, error) {
	// Check Input basic
	if err := out.ValidateBasic(); err != nil {
		return nil, err
	}

	acc := ch.GetAccount(out.Address)

	return acc, nil
}

func adjustInputAccount(acc *account.Account, in tx.TxInput) error {
	return nil
}

func adjustOutpuAccount(acc *account.Account, in tx.TxOutput) error {
	return nil
}

func getInputAccounts(ch *state.Cache, ins []tx.TxInput, req account.Permissions, accs map[crypto.Address]*account.Account) error {

	for _, in := range ins {
		acc, err := getInputAccount(ch, in, req)
		if err != nil {
			return err
		}

		accs[in.Address] = acc
	}

	return nil
}

func getOutputAccounts(ch *state.Cache, outs []tx.TxOutput, accs map[crypto.Address]*account.Account) error {

	for _, out := range outs {
		acc, err := getOutputAccount(ch, out)
		if err != nil {
			return err
		}

		accs[out.Address] = acc
	}
	return nil
}

func adjustInputAccounts(accs map[crypto.Address]*account.Account, ins []tx.TxInput) error {
	for _, in := range ins {
		acc := accs[in.Address]
		if acc == nil {
			return e.Error(e.ErrTxInvalidAddress)
		}

		err := acc.SubtractFromBalance(in.Amount)
		if err != nil {
			return err
		}

		acc.IncSequence()
	}
	return nil
}

func adjustOutputAccounts(accs map[crypto.Address]*account.Account, outs []tx.TxOutput) error {
	for _, out := range outs {
		acc := accs[out.Address]
		if acc == nil {
			return e.Error(e.ErrTxInvalidAddress)
		}

		err := acc.AddToBalance(out.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
func updateCache(c *state.Cache, accs map[crypto.Address]state.Stateacc, outs []tx.TxOutput) error {
	for _, out := range outs {
		acc := accs[out.Address]
		c.Updateacc(acc)
	}
	return nil
}
*/
