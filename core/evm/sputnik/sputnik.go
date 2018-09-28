package sputnik

import (
	"bytes"
	"encoding/hex"
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

func Execute(bc *blockchain.Blockchain, cache *state.Cache, caller, callee *account.Account, tx *tx.CallTx, gas *uint64, isDeploying bool) ([]uint8, error) {

	//var ret []byte
	var retError error

	var addrCaller common.Address

	var addrCallee common.Address

	callerBytes := caller.Address().RawBytes()[2:22]
	addrCaller.SetBytes(callerBytes)
	fmt.Println(addrCaller.Bytes())
	calleeBytes := callee.Address().RawBytes()[2:22]
	addrCallee.SetBytes(calleeBytes)
	fmt.Println(addrCallee.Bytes())

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
			fmt.Println("\n RequireAccount:", require.Address().Bytes())
			if require.Address().IsEmpty() {
				fmt.Println("Commit None Exist")
				vm.CommitNonexist(require.Address())
			} else {
				fmt.Println("Found!")
				acc := GetAccount(cache, require.Address())
				fmt.Println("BAL:",acc.Balance())
				vm.CommitAccount(require.Address(), new(big.Int).SetUint64(acc.Sequence()), new(big.Int).SetUint64(acc.Balance()), acc.Code())
			}

		case sputnikvm.RequireAccountCode:
			fmt.Println("\n RequireAccountCode:", require.Address().Bytes())
			acc := GetAccount(cache, require.Address())
			vm.CommitAccountCode(require.Address(), acc.Code())

		case sputnikvm.RequireAccountStorage:
			fmt.Println("\n RequireAccountStorage:", require.Address().Bytes())
			converted, addr := fromEthAddress(require.Address(), true)
			if !converted {
				fmt.Println("Commit None Exist")
				vm.CommitNonexist(require.Address())
				break
			}
			key := binary.Uint64ToWord256(require.StorageKey().Uint64())
			fmt.Println("Address: ", addr)
			fmt.Printf("Storage key: %d\n", require.StorageKey().Uint64())
			storage, err := cache.GetStorage(addr, key)
			if err != nil {
				vm.CommitAccountStorage(require.Address(), require.StorageKey(), new(big.Int).SetUint64(0))
				break
			}
			var value big.Int
			value.SetUint64(binary.Uint64FromWord256(storage))
			fmt.Printf("Storage value: %d\n", binary.Uint64FromWord256(storage))
			vm.CommitAccountStorage(require.Address(), require.StorageKey(), &value)

		case sputnikvm.RequireBlockhash:
			fmt.Println("\n RequireBlockhash:", require.Address().Bytes())
			blockNumber := new(big.Int).SetUint64(bc.LastBlockHeight())
			var hash common.Hash
			hash.SetBytes(bc.LastBlockHash())
			vm.CommitBlockhash(blockNumber, hash)

		default:
			fmt.Println("\n default...")
			//panic("Panic : unreachable!")
			return nil, retError
		}
	}

	fmt.Println("\n\nVM Successfuly : ", !vm.Failed())
	fmt.Printf("VM Output Len : %d\n", vm.OutLen())
	fmt.Printf("VM Output : %s\n\n", hex.EncodeToString(vm.Output()))

	if vm.Failed() {
		fmt.Println("\n\nOh! Failed :(")
		return []byte{}, retError
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
			fmt.Println("AccountChangeIncreaseBalance")
			amount := acc1.ChangedAmount()
			fmt.Println("Amount:", amount)
			fmt.Println("Address:", acc1.Address())
			targetAcc := GetAccount(cache, acc1.Address())
			fmt.Println("Address:", targetAcc.Address())
			targetAcc.AddToBalance(amount.Uint64())
			fmt.Println("Added Successfully!")
			cache.UpdateAccount(targetAcc)

		case sputnikvm.AccountChangeDecreaseBalance:
			fmt.Println("AccountChangeDecreaseBalance")
			amount := acc1.ChangedAmount()
			targetAcc := GetAccount(cache, acc1.Address())
			targetAcc.SubtractFromBalance(amount.Uint64())
			cache.UpdateAccount(targetAcc)

		case sputnikvm.AccountChangeRemoved:
			fmt.Println("AccountChangeRemoved")
			//removeAccount(acc1.Address())

		case sputnikvm.AccountChangeFull, sputnikvm.AccountChangeCreate:
			fmt.Println("AccountChangeFull - AccountChangeCreate")
			cod := hex.EncodeToString(acc1.Code())
			targetAcc := GetAccount(cache, acc1.Address())
			targetAcc.SetCode(acc1.Code())
			cache.UpdateAccount(targetAcc)
			if len(cod) > 0 {
				fmt.Println("\naddress is: ", hex.EncodeToString(acc1.Address().Bytes()))
				fmt.Println("code is: ", cod)
				println()
			}

			if acc1.Typ() == sputnikvm.AccountChangeFull {
				changeStorage := acc1.ChangedStorage()
				if len(changeStorage) > 0 {
					fmt.Println("Size of changed storage: ", len(changeStorage))
					for i := 0; i < len(changeStorage); i++ {
						fmt.Println("Key: ", common.BigToHash(changeStorage[i].Key).Hex(), "=", common.BigToHash(changeStorage[i].Value).Hex())
						key := binary.Uint64ToWord256(changeStorage[i].Key.Uint64())
						value := binary.Uint64ToWord256(changeStorage[i].Value.Uint64())
						fmt.Println("Key: ", key.Bytes(), "=", value.Bytes())
						cache.SetStorage(targetAcc.Address(), key, value)
						cache.UpdateAccount(targetAcc)

						v,_:=cache.GetStorage(targetAcc.Address(),key)
						fmt.Println("After Set: ", v.Bytes())

					}
					cache.UpdateAccount(targetAcc)
					println("Storage is set successfully!")
				}
			} else {
				/*
					createAccount(acc1.Address())
					changeStorage := acc1.Storage()
					if len(changeStorage) > 0 {
						fmt.Println("Size of changed storage: ", len(changeStorage))
						for i := 0; i < len(changeStorage); i++ {
							fmt.Println("Key: ", common.BigToHash(changeStorage[i].Key).Hex(), "=", common.BigToHash(changeStorage[i].Value).Hex())
							addToAccountStorage(acc1.Address(), *changeStorage[i].Key, *changeStorage[i].Value)
						}
						println()
					}
				*/
			}

		default:
			panic("Panic :unreachable!")
		}

	}

	if !vm.Failed() && isDeploying {
		callee.SetCode(vm.Output())
	}

	for _, log := range vm.Logs() {
		for i := 0; i < len(log.Topics); i++ {
			println("LOG address: ", log.Address.Str())
			ltopics := len(log.Topics)
			if ltopics > 0 {
				for nt := 0; nt < ltopics; nt++ {
					topic := log.Topics[nt]
					fmt.Println("LOG Topic ", nt, ": ", topic.Hex())
				}
			}
			println("LOG data: ", hex.EncodeToString(log.Data))
			println()
		}
	}

	*gas = vm.UsedGas().Uint64()

	out := make([]uint8, vm.OutLen())
	copy(out, vm.Output())

	vm.Free()
	//cache.Flush()
	return out, retError
}

func toWord256(inp *big.Int) binary.Word256 {
	inpBytes := inp.Bytes()
	var ret binary.Word256
	copy(ret[:], inpBytes)
	return ret
}

func GetAccount(cache *state.Cache, ethAddr common.Address) *account.Account {
	converted, addr := fromEthAddress(ethAddr, false)
	//fmt.Println ("Converted:",converted)
	//fmt.Println ("Address:",addr.RawBytes())

	if converted {
		acc, _ := cache.GetAccount(addr)
		if (acc!=nil){
			return acc
		}
	}

	converted, addr = fromEthAddress(ethAddr, true)
	//fmt.Println ("Converted:",converted)
	//fmt.Println ("Address:",addr.RawBytes())
	if converted {
		acc, _ := cache.GetAccount(addr)
		return acc
	}

	return nil
}

func toEthAddress(addr crypto.Address) common.Address {
	var ethAddr common.Address
	ethAddr.SetBytes(addr.RawBytes()[2:22])
	return ethAddr
}

func fromEthAddress(ethAdr common.Address, contract bool) (bool, crypto.Address) {

	var addr crypto.Address
	var err error
	if contract {
		addr, err = crypto.ContractAddress(ethAdr.Bytes())
		if err != nil {
			return false, addr
		}
		return true, addr
	} else {
		if bytes.Equal(ethAdr.Bytes(), crypto.GlobalAddress.RawBytes()[2:22]) {
			addr = crypto.GlobalAddress
			return true, addr
		} else {
			addr, err = crypto.AccountAddress(ethAdr.Bytes())
			if err != nil {
				return false, addr
			}
			return true, addr
		}
	}

	return false, addr
}
