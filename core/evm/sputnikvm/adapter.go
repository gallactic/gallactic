package sputnikvm

import (
	"math/big"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/evm"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/sputnikvm-ffi/go/sputnikvm"
)

type Output struct {
	Failed          bool
	UsedGas         uint64
	Output          []uint8
	Logs            []evm.Log
	ContractAddress *crypto.Address
}

type Adapter interface {
	callerAddress() common.Address
	calleeAddress() *common.Address

	GetGasLimit() uint64
	GetAmount() uint64
	GetData() []byte
	GetNonce() uint64

	setCalleeAddress(address common.Address)

	updateAccount(account *account.Account)
	getAccount(address common.Address) *account.Account

	removeAccount(address common.Address)
	createContractAccount(address common.Address) *account.Account

	updateStorage(address common.Address, key *big.Int, value *big.Int)
	getStorage(address common.Address, key *big.Int) (*big.Int, error)

	addBalance(address common.Address, amount uint64)
	subBalance(address common.Address, amount uint64)
	setBalance(address common.Address, amount uint64)
	setNonce(address common.Address, nonce uint64)
	setCode(address common.Address, code []byte)

	TimeStamp() uint64
	LastBlockNumber() uint64
	LastBlockHash() []byte

	ConvertLog(log sputnikvm.Log) evm.Log
}
