package executors

import (
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs/tx"
)

func getStateObjs(st *state.State, tx tx.Tx) (
	objs map[crypto.Address]state.StateObj, err error) {

	objs = make(map[crypto.Address]state.StateObj)
	ins := tx.Inputs()
	outs := tx.Outputs()
	inAmt := uint64(0)
	outAmt := uint64(0)
	for _, in := range ins {
		// Check Input basic
		if err := in.ValidateBasic(); err != nil {
			return nil, err
		}

		obj := st.GetObj(in.Address)
		if obj == nil {
			return nil, e.Errorf(e.ErrInvalidAddress, "Account %s doesn't exist", in.Address)
		}

		// Check sequences
		if obj.Sequence()+1 != uint64(in.Sequence) {
			return nil, e.Errorf(e.ErrInvalidSequence, "%s has set sequence to %s. It should be %s", in.Address, in.Sequence, obj.Sequence()+uint64(1))
		}

		// Check amount
		if obj.Balance() < uint64(in.Amount) {
			return nil, e.Error(e.ErrInsufficientFunds)
		}

		// Account shouldn't be duplicated
		if _, ok := objs[in.Address]; ok {
			return nil, e.Error(e.ErrTxDuplicateAddress)
		}

		objs[in.Address] = obj
		inAmt += in.Amount
	}

	for _, out := range outs {
		// Check Output basic
		if err := out.ValidateBasic(); err != nil {
			return nil, err
		}

		obj := st.GetObj(out.Address)

		// Account shouldn't be duplicated
		if _, ok := objs[out.Address]; ok {
			return nil, e.Error(e.ErrTxDuplicateAddress)
		}

		objs[out.Address] = obj
		outAmt += out.Amount
	}

	if inAmt < outAmt {
		return nil, e.Error(e.ErrInsufficientFunds)
	}

	return objs, nil
}

func adjustInputs(objs map[crypto.Address]state.StateObj, ins []tx.TxInput) error {
	for _, in := range ins {
		obj := objs[in.Address]
		if obj == nil {
			return e.Error(e.ErrTxInvalidAddress)
		}

		err := obj.SubtractFromBalance(in.Amount)
		if err != nil {
			return err
		}

		obj.IncSequence()
	}
	return nil
}

func adjustOutputs(bojs map[crypto.Address]state.StateObj, outs []tx.TxOutput) error {
	for _, out := range outs {
		obj := bojs[out.Address]
		if obj == nil {
			return e.Error(e.ErrTxInvalidAddress)
		}

		err := obj.AddToBalance(out.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}

// HasPermissions ensures that an account has required permissions
func HasPermissions(st *state.State, acc *account.Account, perm account.Permissions) bool {
	if !permission.EnsureValid(perm) {
		return false
	}

	gAcc := st.GlobalAccount()
	if gAcc.HasPermissions(perm) {
		return true
	}

	if acc.HasPermissions(perm) {
		return true
	}

	return false
}

func HasSendPermission(st *state.State, acc *account.Account) bool {
	return HasPermissions(st, acc, permission.Send)
}

func HasCallPermission(st *state.State, acc *account.Account) bool {
	return HasPermissions(st, acc, permission.Call)
}

func HasCreateContractPermission(st *state.State, acc *account.Account) bool {
	return HasPermissions(st, acc, permission.CreateContract)
}

func HasCreateAccountPermission(st *state.State, acc *account.Account) bool {
	return HasPermissions(st, acc, permission.CreateAccount)
}

func HasBondPermission(st *state.State, acc *account.Account) bool {
	return HasPermissions(st, acc, permission.Bond)
}

func HasModifyPermission(st *state.State, acc *account.Account) bool {
	return HasPermissions(st, acc, permission.ModifyPermission)
}
