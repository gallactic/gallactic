package sputnikvm

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereumproject/go-ethereum/common"
	///e "github.com/gallactic/gallactic/errors"
	"github.com/gallactic/sputnikvm-ffi/go/sputnikvm"
)

/// TODO: Gheis
/// Please look here as an example of implementing the SputnikVM
/// https://github.com/ethereumproject/go-ethereum/blob/master/core/multivm_processor.go
/// Fix all TODOs and also create a `receipt` object
func Execute(adapter Adapter) (Output, error) {
	var out Output
	var retError error

	if adapter.calleeAddress() == nil && len(adapter.GetData()) == 0 {
		// TODO: Returning without calling vm.Free()
		// TODO: use ErrInternalEvm as an error code
		///return out, e.Errorf(e.ErrInternalEvm, "Zero Bytes of Codes")
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

	vm := sputnikvm.NewGallactic(&transaction, &header)

	defer vm.Free()

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
					//adapter.createContractAccount(require.Address())
					//vm.CommitAccount(require.Address(), new(big.Int).SetUint64(0), new(big.Int).SetUint64(0), adapter.GetData())
					vm.CommitNonexist(require.Address())
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
			// TODO: get block nuber
			//blockNumber := require.BlockNumber()
			blockNumber := new(big.Int).SetUint64(adapter.LastBlockNumber())
			var blockHash common.Hash
			// TODO: Get block hash and check if the block exists....
			// (!Not only last block number)
			blockHash.SetBytes(adapter.LastBlockHash())
			vm.CommitBlockhash(blockNumber, blockHash)

		default:
			// TODO: Returning without calling vm.Free()
			///return out, e.Errorf(e.ErrInternalEvm, "Zero Bytes of Codes")
			return out, errors.New("Not Supperted Requirement by Sputnik")
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
			//TODO: removeAccount(changedAcc.Address())

		case sputnikvm.AccountChangeFull:
			// TODO: Update balance, nonce, code
			changeStorage := changedAcc.ChangedStorage()
			if len(changeStorage) > 0 {
				for i := 0; i < len(changeStorage); i++ {
					key := changeStorage[i].Key
					value := changeStorage[i].Value
					adapter.updateStorage(changedAcc.Address(), key, value)
				}
			}

		case sputnikvm.AccountChangeCreate:
			// TODO: Update balance, nonce, code
			adapter.createContractAccount(changedAcc.Address())

			changeStorage := changedAcc.Storage()
			if len(changeStorage) > 0 {
				for i := 0; i < len(changeStorage); i++ {
					key := changeStorage[i].Key
					value := changeStorage[i].Value
					adapter.updateStorage(changedAcc.Address(), key, value)
				}
			}
			adapter.setCode(changedAcc.Address(), changedAcc.Code())

			if adapter.calleeAddress() == nil {
				adapter.setCalleeAddress(changedAcc.Address())
			}

		default:
			//Return error :unreachable!
			// TODO: Returning without calling vm.Free()
			///return out, e.Errorf(e.ErrInternalEvm, "Zero Bytes of Codes")
			return out, errors.New("unreachable")
		}

	}

	//Extract logs and events
	for _, log := range vm.Logs() {
		adapter.log(log.Address, log.Topics, log.Data)
	}

	if vm.Failed() {
		out.Failed = true
		// TODO: Returning without calling vm.Free()
		///return out, e.Errorf(e.ErrInternalEvm, "Zero Bytes of Codes")
		return out, fmt.Errorf("VM Failed")
	}
	out.UsedGas = vm.UsedGas().Uint64()
	out.Output = make([]uint8, vm.OutLen())
	copy(out.Output, vm.Output())

	//vm.Free()
	return out, retError
}
