package executors

import (
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs/tx"
)

func getInputAccount(ch *state.Cache, in tx.TxInput, req account.Permissions) (*account.Account, error) {
	acc, err := ch.GetAccount(in.Address)
	if err != nil {
		return nil, err
	}

	// Check sequences
	if acc.Sequence()+1 != uint64(in.Sequence) {
		return nil, e.Errorf(e.ErrInvalidSequence, "%v has set sequence to %v. It should be %v", in.Address, in.Sequence, acc.Sequence()+1)
	}

	// Check amount
	if acc.Balance() < uint64(in.Amount) {
		return nil, e.Error(e.ErrInsufficientFunds)
	}

	if !ch.HasPermissions(acc, req) {
		return nil, e.Errorf(e.ErrPermDenied, "%v has %v but needs %v permission", in.Address, acc.Permissions(), permission.Send)
	}

	return acc, nil
}

func getOutputAccount(ch *state.Cache, out tx.TxOutput) (*account.Account, error) {
	if !ch.HasAccount(out.Address) {
		return nil, nil
	}

	return ch.GetAccount(out.Address)
}

func getInputValidator(ch *state.Cache, in tx.TxInput) (*validator.Validator, error) {
	val, err := ch.GetValidator(in.Address)
	if err != nil {
		return nil, err
	}

	// Check sequences
	if val.Sequence()+1 != uint64(in.Sequence) {
		return nil, e.Errorf(e.ErrInvalidSequence, "%v has set sequence to %v. It should be %v", in.Address, in.Sequence, val.Sequence()+1)
	}

	// Check amount
	if val.Stake() < uint64(in.Amount) {
		return nil, e.Error(e.ErrInsufficientFunds)
	}

	return val, nil
}

func getOutputValidator(ch *state.Cache, out tx.TxOutput) (*validator.Validator, error) {
	if !ch.HasValidator(out.Address) {
		return nil, nil
	}
	return ch.GetValidator(out.Address)
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
			return e.Error(e.ErrInvalidAddress)
		}

		if err := adjustInputAccount(acc, in); err != nil {
			return err
		}
	}
	return nil
}

func adjustOutputAccounts(accs map[crypto.Address]*account.Account, outs []tx.TxOutput) error {
	for _, out := range outs {
		acc := accs[out.Address]
		if acc == nil {
			return e.Error(e.ErrInvalidAddress)
		}

		if err := adjustOutputAccount(acc, out); err != nil {
			return err
		}
	}
	return nil
}

func adjustInputAccount(acc *account.Account, in tx.TxInput) error {
	err := acc.SubtractFromBalance(in.Amount)
	if err != nil {
		return err
	}

	acc.IncSequence()
	return nil
}

func adjustOutputAccount(acc *account.Account, out tx.TxOutput) error {
	err := acc.AddToBalance(out.Amount)
	if err != nil {
		return err
	}

	return nil
}

func adjustInputValidator(val *validator.Validator, in tx.TxInput) error {
	err := val.SubtractFromStake(in.Amount)
	if err != nil {
		return err
	}

	val.IncSequence()
	return nil
}

func adjustOutputValidator(val *validator.Validator, out tx.TxOutput) error {
	err := val.AddToStake(out.Amount)
	if err != nil {
		return err
	}

	return nil
}
