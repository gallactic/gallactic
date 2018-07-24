package burrow

import (
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/txs/tx"
	burrowState "github.com/hyperledger/burrow/account/state"
	burrowBinary "github.com/hyperledger/burrow/binary"
	burrowEVM "github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/logging"
	burrowTx "github.com/hyperledger/burrow/txs"
	burrowPayload "github.com/hyperledger/burrow/txs/payload"
)

func CallCode(bc *blockchain.Blockchain, caller, callee *account.Account, data []byte, value, fee, gasLimit uint64, gas *uint64) (output []byte, err error) {

	params := burrowEVM.Params{
		BlockHeight: bc.LastBlockHeight(),
		BlockHash:   burrowBinary.LeftPadWord256(bc.LastBlockHash()),
		BlockTime:   bc.LastBlockTime().Unix(),
		GasLimit:    uint64(1000000),
	}

	bCaller := toBurrowAccount(caller)
	bCallee := toBurrowAccount(callee)
	bCalleeAddr := bCallee.Address()
	code := bCallee.Code()

	bPayload := burrowPayload.NewCallTxWithSequence(bCaller.PublicKey(), &bCalleeAddr,
		data, value, gasLimit, fee, bCaller.Sequence()+1)
	bTx := burrowTx.NewTx(bPayload)

	st := bState{st: bc.State()}
	txCache := burrowState.NewCache(st, burrowState.Name("TxCache"))
	publisher := eventPublisher{}
	vm := burrowEVM.NewVM(params, bCaller.Address(), bTx, logging.NewNoopLogger())
	vm.SetPublisher(publisher)
	ret, exception := vm.Call(txCache, bCaller, bCallee, code, data, value, gas)
	txCache.Flush(st, st)

	return ret, exception

}

func Call(bc *blockchain.Blockchain, caller, callee *account.Account, tx *tx.CallTx, gas *uint64) (output []byte, err error) {
	data := tx.Data()
	value := tx.Amount()
	gasLimit := tx.GasLimit()

	return CallCode(bc, caller, callee, data, value, tx.Fee(), gasLimit, gas)
}
