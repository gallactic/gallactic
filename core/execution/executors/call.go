package executors

import (
	"fmt"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/evm"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"
	"github.com/hyperledger/burrow/logging"
)

type CallContext struct {
	Committer bool
	Cache     *state.Cache
	Logger    *logging.Logger
}

func (ctx *CallContext) Execute(txEnv *txs.Envelope) error {
	tx, ok := txEnv.Tx.(*tx.CallTx)
	if !ok {
		return e.Error(e.ErrTxWrongPayload)
	}

	caller, err := getInputAccount(ctx.Cache, tx.Caller(), permission.Call)
	if err != nil {
		return err
	}

	var callee *account.Account
	if tx.CreateContract() {
		if !ctx.Cache.HasPermissions(caller, permission.CreateContract) {
			return e.Errorf(e.ErrPermDenied, "%s has %s but needs %s", caller.Address(), caller.Permissions(), permission.CreateContract)
		}
	} else {
		// check if its a native contract
		if evm.IsRegisteredNativeContract(tx.Callee().Address.Word256()) {
			return fmt.Errorf("attempt to call a native contract at %s, "+
				"but native contracts cannot be called using CallTx. Use a "+
				"contract that calls the native contract or the appropriate tx "+
				"type (eg. PermissionsTx, NameTx)", tx.Callee().Address)
		}

		/// TODO : write test for this case: create and call in same block
		callee = ctx.Cache.GetAccount(tx.Callee().Address)
	}

	/// Update account
	ctx.Cache.UpdateAccount(caller)
	ctx.Cache.UpdateAccount(callee)

	return nil
}
