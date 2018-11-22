package sputnik

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/gallactic/sputnikvm-ffi/go/sputnikvm"
)

func Execute(adapter Adapter) (Output, error) {
	var out Output
	var retError error

	if adapter.calleeAddress() == nil && len(adapter.GetData()) == 0 {
		return out, errors.New("Zero Bytes of Codes")
	}

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

	vm := sputnikvm.NewFrontier(&transaction, &header)

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
					vm.CommitAccount(require.Address(), new(big.Int).SetUint64(acc.Sequence()), new(big.Int).SetUint64(acc.Balance()), acc.Code())
				} else {
					adapter.createContractAccount(require.Address())
					vm.CommitAccount(require.Address(), new(big.Int).SetUint64(0), new(big.Int).SetUint64(0), adapter.GetData())
				}
			}

		case sputnikvm.RequireAccountCode:
			acc := adapter.getAccount(require.Address())
			if acc != nil {
				vm.CommitAccountCode(require.Address(), acc.Code())
			} else {
				return out, errors.New("No Exist Account for Acquire Code")
			}

		case sputnikvm.RequireAccountStorage:
			storage, err := adapter.getStorage(require.Address(), require.StorageKey())
			if err != nil {
				vm.CommitAccountStorage(require.Address(), require.StorageKey(), new(big.Int).SetUint64(0))
				break
			}
			vm.CommitAccountStorage(require.Address(), require.StorageKey(), storage)

		case sputnikvm.RequireBlockhash:
			//Require Blockhash
			blockNumber := new(big.Int).SetUint64(adapter.LastBlockNumber())
			var blockHash common.Hash
			blockHash.SetBytes(adapter.LastBlockHash())
			vm.CommitBlockhash(blockNumber, blockHash)

		default:
			return out, errors.New("Not Supperted Requirement by Sputnik")
		}
	}

	if vm.Failed() {
		out.Failed = true
		return out, fmt.Errorf("VM Failed")
	}

	changedAccs := vm.AccountChanges()
	accLen := len(changedAccs)
	contractAddressIsSet := false

	for i := 0; i < accLen; i++ {
		changedAcc := changedAccs[i]

		if changedAcc.Address().IsEmpty() {
			continue
		}

		if !contractAddressIsSet && adapter.calleeAddress() == nil {
			out.ContractAddress.SetBytes(changedAcc.Address().Bytes())
			adapter.setCalleeAddress(changedAcc.Address())
			contractAddressIsSet = true
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
			//TODO: removeAccount(changedAcc.Address())

		case sputnikvm.AccountChangeFull, sputnikvm.AccountChangeCreate:
			// Change or Create Account
			if changedAcc.Typ() == sputnikvm.AccountChangeFull {
				changeStorage := changedAcc.ChangedStorage()
				if len(changeStorage) > 0 {
					for i := 0; i < len(changeStorage); i++ {
						key := changeStorage[i].Key
						value := changeStorage[i].Value
						adapter.updateStorage(changedAcc.Address(), key, value)
					}
				}
			} else {
				changeStorage := changedAcc.Storage()
				if len(changeStorage) > 0 {
					for i := 0; i < len(changeStorage); i++ {
						key := changeStorage[i].Key
						value := changeStorage[i].Value
						adapter.updateStorage(changedAcc.Address(), key, value)
					}
				}
				adapter.setCode(changedAcc.Address(), changedAcc.Code())
			}

		default:
			//Return error :unreachable!
			return out, errors.New("unreachable")
		}

	}

	//Extract logs and events
	for _, log := range vm.Logs() {
		adapter.log(log.Address, log.Topics, log.Data)
	}

	out.UsedGas = vm.UsedGas().Uint64()
	out.Output = make([]uint8, vm.OutLen())
	copy(out.Output, vm.Output())

	vm.Free()
	return out, retError
}
