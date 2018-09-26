package sputnik

import (
	"fmt"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/txs/tx"

	"math/big"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/sputnikvm-ffi/go/sputnikvm"
)

func Call(bc *blockchain.Blockchain, cache *state.Cache, caller, callee *account.Account, tx *tx.CallTx, gas *uint64) (output []byte, err error) {

	//var ret []byte
	var retError error

	var addrCaller common.Address

	var addrCallee common.Address

	callerBytes := caller.Address().RawBytes()
	addrCaller.SetBytes(callerBytes)

	// if tx.CreateContract() {
	// 	addr := DeriveNewContractAddress(caller)
	// 	calleeBytes := addr.RawBytes()
	// 	addrCallee.SetBytes(calleeBytes)
	// } else {
	calleeBytes := callee.Address().RawBytes()
	addrCallee.SetBytes(calleeBytes)
	//}

	transaction := sputnikvm.Transaction{
		Caller:   addrCaller,
		GasPrice: new(big.Int).SetUint64(0),
		GasLimit: new(big.Int).SetUint64(tx.GasLimit()),
		Address:  &addrCallee,
		Value:    new(big.Int).SetUint64(tx.Amount()),
		Input:    tx.Data(),
		Nonce:    new(big.Int).SetUint64(tx.Caller().Sequence),
	}

	header := sputnikvm.HeaderParams{
		Beneficiary: *new(common.Address),
		Timestamp:   uint64(bc.LastBlockTime().Unix()),
		Number:      new(big.Int).SetUint64(bc.LastBlockHeight()),
		Difficulty:  new(big.Int).SetUint64(0),
		GasLimit:    new(big.Int).SetUint64(tx.GasLimit()),
	}

	vm := sputnikvm.NewFrontier(&transaction, &header)

	fmt.Println("\nSputnikvm is starting...")

Loop:
	for {
		require := vm.Fire()

		switch require.Typ() {

		case sputnikvm.RequireNone:
			break Loop

		case sputnikvm.RequireAccount:

			if require.Address().IsEmpty() {
				vm.CommitNonexist(require.Address())
			} else {
				addr, err := crypto.AddressFromRawBytes(require.Address().Bytes())
				if err != nil {
					return nil, err
				}
				acc, err := cache.GetAccount(addr)
				if err != nil {
					return nil, err
				}
				vm.CommitAccount(require.Address(), new(big.Int).SetUint64(acc.Sequence()), new(big.Int).SetUint64(acc.Balance()), acc.Code())
			}

		case sputnikvm.RequireAccountCode:
			addr, err := crypto.AddressFromRawBytes(require.Address().Bytes())
			if err != nil {
				return nil, err
			}
			acc, err := cache.GetAccount(addr)
			if err != nil {
				return nil, err
			}
			vm.CommitAccountCode(require.Address(), acc.Code())

		case sputnikvm.RequireAccountStorage:
			addr, err := crypto.AddressFromRawBytes(require.Address().Bytes())
			if err != nil {
				return nil, err
			}
			var key binary.Word256
			copy(require.StorageKey().Bytes(), key.Bytes())
			storage, err := cache.GetStorage(addr, key)
			if err != nil {
				vm.CommitAccountStorage(require.Address(), require.StorageKey(), new(big.Int).SetUint64(0))
			}
			var value big.Int
			copy(storage.Bytes(), value.Bytes())
			vm.CommitAccountStorage(require.Address(), require.StorageKey(), &value)

		case sputnikvm.RequireBlockhash:
			blockNumber := new(big.Int).SetUint64(bc.LastBlockHeight())
			var hash common.Hash
			copy(bc.LastBlockHash(), hash.Bytes())
			vm.CommitBlockhash(blockNumber, hash)

		default:
			//panic("Panic : unreachable!")
			return nil, retError
		}
	}

	*gas = vm.UsedGas().Uint64()

	var out []byte
	copy(vm.Output(), out)

	//cache.Flush()
	return out, retError
}

// Create a new account from a parent 'creator' account. The creator account will have its
// sequence number incremented
func DeriveNewContractAddress(creator *account.Account) crypto.Address {
	// Generate an address
	seq := creator.Sequence()
	creator.IncSequence()

	addr := crypto.DeriveContractAddress(creator.Address(), seq)

	return addr
}

/*
func _Call(bc *blockchain.Blockchain, caller, callee *account.Account, tx *tx.CallTx, gas *uint64) (output []byte, err error) {

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

	if err != nil {
		err = e.Error(e.ErrInternalEvm)
	}

	return ret, err
}
*/
