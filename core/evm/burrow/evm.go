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

func Call(bc *blockchain.Blockchain, caller, callee *account.Account, tx *tx.CallTx) {

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
	data := tx.Data()
	value := tx.Amount()
	gas := tx.GasLimit()

	bPayload := burrowPayload.NewCallTxWithSequence(bCaller.PublicKey(), &bCalleeAddr,
		tx.Data(), tx.Amount(), tx.GasLimit(), tx.Fee(), tx.Sequence())
	bTx := burrowTx.NewTx(bPayload)

	st := bState{st: bc.State()}
	txCache := burrowState.NewCache(st, burrowState.Name("TxCache"))
	publisher := eventPublisher{}
	vm := burrowEVM.NewVM(params, bCaller.Address(), bTx, logging.NewNoopLogger())
	vm.SetPublisher(publisher)
	/*ret, exception := */ vm.Call(txCache, bCaller, bCallee, code, data, value, &gas)

}
