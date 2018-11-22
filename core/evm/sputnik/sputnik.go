package sputnik

import (
	"bytes"
	"errors"

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

//Execute for executing virtual machine
func Execute(bc *blockchain.Blockchain, cache *state.Cache, caller, callee *account.Account, tx *tx.CallTx, gas *uint64, isDeploying bool) ([]uint8, error) {

	var retError error

	var addrCaller common.Address

	var addrCallee common.Address

	callerBytes := caller.Address().RawBytes()[2:22]
	addrCaller.SetBytes(callerBytes)

	calleeBytes := callee.Address().RawBytes()[2:22]
	addrCallee.SetBytes(calleeBytes)

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

Loop:
	for {
		require := vm.Fire()

		switch require.Typ() {

		case sputnikvm.RequireNone:
			break Loop

		case sputnikvm.RequireAccount:
			//Require Account
			if require.Address().IsEmpty() {
				vm.CommitNonexist(require.Address())
			} else {
				acc := GetAccount(cache, require.Address())
				if acc != nil {
					vm.CommitAccount(require.Address(), new(big.Int).SetUint64(acc.Sequence()), new(big.Int).SetUint64(acc.Balance()), acc.Code())
				} else {
					return []byte{}, errors.New("No Exist Account")
				}

			}

		case sputnikvm.RequireAccountCode:
			//Require Account Code
			acc := GetAccount(cache, require.Address())
			if acc != nil {
				vm.CommitAccountCode(require.Address(), acc.Code())
			} else {
				return []byte{}, errors.New("No Exist Account for Acquire Code")
			}

		case sputnikvm.RequireAccountStorage:
			//Require Account Storage
			converted, addr := fromEthAddress(require.Address(), true)
			if !converted {
				vm.CommitNonexist(require.Address())
				break
			}
			key := binary.Uint64ToWord256(require.StorageKey().Uint64())
			storage, err := cache.GetStorage(addr, key)
			if err != nil {
				vm.CommitAccountStorage(require.Address(), require.StorageKey(), new(big.Int).SetUint64(0))
				break
			}
			var value big.Int
			value.SetUint64(binary.Uint64FromWord256(storage))
			vm.CommitAccountStorage(require.Address(), require.StorageKey(), &value)

		case sputnikvm.RequireBlockhash:
			//Require Blockhash
			blockNumber := new(big.Int).SetUint64(bc.LastBlockHeight())
			var hash common.Hash
			hash.SetBytes(bc.LastBlockHash())
			vm.CommitBlockhash(blockNumber, hash)

		default:
			return nil, errors.New("Not Supperted Require by Sputnik")
		}
	}

	if vm.Failed() {
		return []byte{}, errors.New("VM Failed")
	}

	changedaccs := vm.AccountChanges()
	lacc := len(changedaccs)

	for i := 0; i < lacc; i++ {
		acc1 := changedaccs[i]

		if acc1.Address().IsEmpty() {
			continue
		}

		switch acc1.Typ() {

		case sputnikvm.AccountChangeIncreaseBalance:
			//Increase Balance
			amount := acc1.ChangedAmount()
			targetAcc := GetAccount(cache, acc1.Address())
			targetAcc.AddToBalance(amount.Uint64())
			cache.UpdateAccount(targetAcc)

		case sputnikvm.AccountChangeDecreaseBalance:
			//Decrease Balance
			amount := acc1.ChangedAmount()
			targetAcc := GetAccount(cache, acc1.Address())
			targetAcc.SubtractFromBalance(amount.Uint64())
			cache.UpdateAccount(targetAcc)

		case sputnikvm.AccountChangeRemoved:
			//Removing Account
			//TODO: removeAccount(acc1.Address())

		case sputnikvm.AccountChangeFull, sputnikvm.AccountChangeCreate:
			// Change or Create Account
			targetAcc := GetAccount(cache, acc1.Address())
			targetAcc.SetCode(acc1.Code())
			cache.UpdateAccount(targetAcc)

			if acc1.Typ() == sputnikvm.AccountChangeFull {
				changeStorage := acc1.ChangedStorage()
				if len(changeStorage) > 0 {
					for i := 0; i < len(changeStorage); i++ {
						key := binary.Uint64ToWord256(changeStorage[i].Key.Uint64())
						value := binary.Uint64ToWord256(changeStorage[i].Value.Uint64())
						cache.SetStorage(targetAcc.Address(), key, value)
						cache.UpdateAccount(targetAcc)
					}
					cache.UpdateAccount(targetAcc)
					//Ok, Storage is set successfully!
				}
			} else {
				//TODO: createAccount(acc1.Address())
				changeStorage := acc1.Storage()
				if len(changeStorage) > 0 {
					for i := 0; i < len(changeStorage); i++ {
						//TODO: addToAccountStorage(acc1.Address(), *changeStorage[i].Key, *changeStorage[i].Value)
					}
				}
			}

		default:
			//Return error :unreachable!
			return []byte{}, errors.New("unreachable")
		}

	}

	if !vm.Failed() && isDeploying {
		callee.SetCode(vm.Output())
	}

	*gas = vm.UsedGas().Uint64()

	out := make([]uint8, vm.OutLen())
	copy(out, vm.Output())

	vm.Free()
	return out, retError
}

//GetAccount for getting account using Sputnik Address
func GetAccount(cache *state.Cache, ethAddr common.Address) *account.Account {
	converted, addr := fromEthAddress(ethAddr, false)

	if converted {
		acc, _ := cache.GetAccount(addr)
		if acc != nil {
			return acc
		}
	}

	converted, addr = fromEthAddress(ethAddr, true)

	if converted {
		acc, _ := cache.GetAccount(addr)
		return acc
	}

	return nil
}

func fromEthAddress(ethAdr common.Address, contract bool) (bool, crypto.Address) {

	var addr crypto.Address
	var err error
	if contract {
		addr, err = crypto.ContractAddress(ethAdr.Bytes())
		if err != nil {
			return false, addr
		}
	} else {
		if bytes.Equal(ethAdr.Bytes(), crypto.GlobalAddress.RawBytes()[2:22]) {
			addr = crypto.GlobalAddress
		} else {
			addr, err = crypto.AccountAddress(ethAdr.Bytes())
			if err != nil {
				return false, addr
			}
		}
	}

	return true, addr
}
