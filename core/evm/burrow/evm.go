package burrow

import (
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/txs/tx"
	burrowState "github.com/hyperledger/burrow/acm/state"
	burrowBinary "github.com/hyperledger/burrow/binary"
	burrowEVM "github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/logging"
	burrowTx "github.com/hyperledger/burrow/txs"
	burrowPayload "github.com/hyperledger/burrow/txs/payload"
	"github.com/gallactic/gallactic/errors"
)

func Call(bc *blockchain.Blockchain, caller, callee *account.Account, tx *tx.CallTx, gas *uint64) (output []byte, err error) {

	params := burrowEVM.Params{
		BlockHeight: bc.LastBlockHeight(),
		BlockHash:   burrowBinary.LeftPadWord256(bc.LastBlockHash()),
		BlockTime:   bc.LastBlockTime().Unix(),
		GasLimit:    tx.GasLimit(),
	}

	bCaller := toBurrowAccount(caller)
	bCallee := toBurrowAccount(callee)
	bCalleeAddr := bCallee.Address()
	code := bCallee.Code()

	bPayload := burrowPayload.NewCallTxWithSequence(bCaller.PublicKey(), &bCalleeAddr,
		tx.Data(), tx.Amount(), tx.GasLimit(), tx.Fee(), bCaller.Sequence()+1)
	bTx := burrowTx.NewTx(bPayload)

	st := bState{st: bc.State()}
	txCache := burrowState.NewCache(st, burrowState.Name("TxCache"))
	vm := burrowEVM.NewVM(params, bCaller.Address(), bTx, logging.NewNoopLogger())
	eventSink := &noopEventSink{}
	vm.SetEventSink(eventSink)
	*gas = tx.GasLimit()
	ret, err := vm.Call(txCache, bCaller, bCallee, code, tx.Data(), tx.Amount(), gas)
	txCache.Flush(st, st)


	// TODO We need to fix this code in future, it's not a good Idea to only return a generic error for all evm issues!

	if err != nil{
		err = e.Error(e.ErrInternalEvm)
	}

	return ret, err
}
