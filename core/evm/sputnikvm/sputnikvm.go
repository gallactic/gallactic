package sputnikvm

import (
	"math/big"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/gallactic/sputnikvm-ffi/go/sputnikvm"

	tmRPC "github.com/tendermint/tendermint/rpc/core"
)

func Execute(adapter Adapter) Output {
	var out Output

	transaction := sputnikvm.Transaction{
		Caller:   adapter.callerAddress(),
		GasPrice: new(big.Int).SetUint64(0),
		GasLimit: new(big.Int).SetUint64(adapter.GetGasLimit()),
		Address:  adapter.calleeAddress(),
		Value:    new(big.Int).SetUint64(adapter.GetAmount()),
		Input:    adapter.GetData(),
		Nonce:    new(big.Int).SetUint64(adapter.GetNonce()),
	}

	header := sputnikvm.HeaderParams{
		Beneficiary: *new(common.Address),
		Timestamp:   adapter.TimeStamp(),
		Number:      new(big.Int).SetUint64(adapter.LastBlockNumber()),
		Difficulty:  new(big.Int).SetUint64(0),
		GasLimit:    new(big.Int).SetUint64(adapter.GetGasLimit()),
	}

	vm := sputnikvm.NewGallactic(&transaction, &header)

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
				acc := adapter.getAccount(require.Address())
				if acc != nil {
					vm.CommitAccount(require.Address(), new(big.Int).SetUint64(acc.Sequence()),
						new(big.Int).SetUint64(acc.Balance()), acc.Code())
				} else {
					vm.CommitNonexist(require.Address())
				}
			}

		case sputnikvm.RequireAccountCode:
			acc := adapter.getAccount(require.Address())
			if acc != nil {
				vm.CommitAccountCode(require.Address(), acc.Code())
			} else {
				vm.CommitNonexist(require.Address())
			}

		case sputnikvm.RequireAccountStorage:
			storage, err := adapter.getStorage(require.Address(), require.StorageKey())
			if err != nil {
				vm.CommitAccountStorage(require.Address(), require.StorageKey(), new(big.Int).SetUint64(0))
				break
			}
			vm.CommitAccountStorage(require.Address(), require.StorageKey(), storage)

		case sputnikvm.RequireBlockhash:
			var blockHash common.Hash

			reqblockNumber := require.BlockNumber().Int64()
			block, err := tmRPC.Block(&reqblockNumber)
			if err == nil {
				hash := block.Block.Hash().Bytes()
				blockHash.SetBytes(hash)
			}
			vm.CommitBlockhash(require.BlockNumber(), blockHash)

		default:
			/// The cache is irreversible, during delivering call transaction
			/// Better panic in case of unexpected error happens
			/// rather than changing the state which the tx is not delivered.
			panic("unreachable")
		}
	}

	changedAccs := vm.AccountChanges()
	accLen := len(changedAccs)

	for i := 0; i < accLen; i++ {
		changedAcc := changedAccs[i]

		if changedAcc.Address().IsEmpty() {
			continue
		}

		switch changedAcc.Typ() {

		case sputnikvm.AccountChangeIncreaseBalance:
			//Increase Balance
			amount := changedAcc.ChangedAmount()
			adapter.addBalance(changedAcc.Address(), amount.Uint64())

		case sputnikvm.AccountChangeDecreaseBalance:
			//Decrease Balance
			amount := changedAcc.ChangedAmount()
			adapter.subBalance(changedAcc.Address(), amount.Uint64())

		case sputnikvm.AccountChangeRemoved:
			//Removing Account
			adapter.removeAccount(changedAcc.Address())

		case sputnikvm.AccountChangeFull:
			changeStorage := changedAcc.ChangedStorage()
			if len(changeStorage) > 0 {
				for i := 0; i < len(changeStorage); i++ {
					key := changeStorage[i].Key
					value := changeStorage[i].Value
					adapter.updateStorage(changedAcc.Address(), key, value)
				}
			}
			adapter.setAccount(changedAcc.Address(), changedAcc.Balance().Uint64(), changedAcc.Code(), changedAcc.Nonce().Uint64())

		case sputnikvm.AccountChangeCreate:
			acc := adapter.createContractAccount(changedAcc.Address())

			changeStorage := changedAcc.Storage()
			if len(changeStorage) > 0 {
				for i := 0; i < len(changeStorage); i++ {
					key := changeStorage[i].Key
					value := changeStorage[i].Value
					adapter.updateStorage(changedAcc.Address(), key, value)
				}
			}
			acc.SetBalance(changedAcc.Balance().Uint64())
			acc.SetSequence(changedAcc.Nonce().Uint64())
			acc.SetCode(changedAcc.Code())

			adapter.updateAccount(acc)

			if out.ContractAddress == nil {
				addr := acc.Address()
				out.ContractAddress = &addr
			}

		default:
			panic("unreachable")
		}
	}

	if vm.Failed() {
		out.Failed = true
	} else {
		out.Failed = false
	}

	out.UsedGas = vm.UsedGas().Uint64()

	//Extract logs and events
	for _, log := range vm.Logs() {
		adapter.log(log.Address, log.Topics, log.Data)
	}

	out.UsedGas = vm.UsedGas().Uint64()
	out.Output = make([]uint8, vm.OutLen())
	copy(out.Output, vm.Output())

	vm.Free()

	return out
}
