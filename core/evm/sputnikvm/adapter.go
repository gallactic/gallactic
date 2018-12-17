package sputnikvm

import (
	"math/big"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/gallactic/gallactic/core/account"
)

type Output struct {
	Failed  bool
	UsedGas uint64
	Output  []uint8
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
	setAccount(address common.Address, balance uint64, code []byte, nonce uint64)

	log(address common.Address, topics []common.Hash, data []byte)

	TimeStamp() uint64
	LastBlockNumber() uint64
	LastBlockHash() []byte
}
