package sputnikvm

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/sputnikvm-ffi/go/sputnikvm"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/evm"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/crypto"
)

type GallacticAdapter struct {
	BlockChain *blockchain.Blockchain
	Cache      *state.Cache
	Caller     *account.Account
	Callee     *account.Account
	GasLimit   uint64
	Amount     uint64
	Data       []byte
	Nonce      uint64
}

func (ga *GallacticAdapter) calleeAddress() *common.Address {
	var addr *common.Address
	if ga.Callee != nil {
		addr = new(common.Address)
		addr.SetBytes(toEthAddress(ga.Callee.Address()).Bytes())
	}
	return addr
}

func (ga *GallacticAdapter) callerAddress() common.Address {
	if ga.Caller == nil {
		return common.Address{}
	}
	return toEthAddress(ga.Caller.Address())
}

func (ga *GallacticAdapter) GetGasLimit() *big.Int {
	return bigint(ga.GasLimit)
}

func (ga *GallacticAdapter) GetGasPrice() *big.Int {
	return bigint(0) /// TODO: Using testnet is free
}

func (ga *GallacticAdapter) GetAmount() *big.Int {
	return bigint(ga.Amount)
}

func (ga *GallacticAdapter) GetData() []byte {
	return ga.Data
}

func (ga *GallacticAdapter) GetNonce() *big.Int {
	return bigint(ga.Nonce)
}

func (ga *GallacticAdapter) createAccount(address common.Address) *account.Account {
	addr := fromEthAddress(address, false)
	acc, _ := account.NewAccount(addr)
	return acc
}

func (ga *GallacticAdapter) updateAccount(acc *account.Account) {
	ga.Cache.UpdateAccount(acc)
}

func (ga *GallacticAdapter) updateStorage(address common.Address, key *big.Int, value *big.Int) {
	addr := fromEthAddress(address, true)
	wKey := binary.LeftPadWord256(key.Bytes())
	wValue := binary.LeftPadWord256(value.Bytes())
	ga.Cache.SetStorage(addr, wKey, wValue)
}

func (ga *GallacticAdapter) getStorage(address common.Address, key *big.Int) (*big.Int, error) {
	addr := fromEthAddress(address, true)
	wKey := binary.LeftPadWord256(key.Bytes())
	wValue, err := ga.Cache.GetStorage(addr, wKey)
	var value big.Int
	if err != nil {
		value.SetUint64(0)
	} else {
		value.SetBytes(wValue.Bytes())
	}
	return &value, err
}

func (ga *GallacticAdapter) createContractAccount(address common.Address) *account.Account {
	addr := fromEthAddress(address, true)
	acc, _ := account.NewContractAccount(addr)
	return acc
}

func (ga *GallacticAdapter) setCalleeAddress(address common.Address) {
	if ga.Callee == nil {
		ga.Callee = ga.getAccount(address)
	}
}

func (ga *GallacticAdapter) removeAccount(address common.Address) {
	addr := fromEthAddress(address, true)
	ga.Cache.RemoveAccount(addr)
}

func (ga *GallacticAdapter) addBalance(address common.Address, amount uint64) {
	acc := ga.getAccount(address)
	acc.AddToBalance(amount)
	ga.Cache.UpdateAccount(acc)
}

func (ga *GallacticAdapter) subBalance(address common.Address, amount uint64) {
	acc := ga.getAccount(address)
	acc.SubtractFromBalance(amount)
	ga.Cache.UpdateAccount(acc)
}

func (ga *GallacticAdapter) setBalance(address common.Address, amount uint64) {
	acc := ga.getAccount(address)
	acc.SetBalance(amount)
	ga.Cache.UpdateAccount(acc)
}

func (ga *GallacticAdapter) setNonce(address common.Address, nonce uint64) {
	acc := ga.getAccount(address)
	acc.SetSequence(nonce)
	ga.Cache.UpdateAccount(acc)
}

func (ga *GallacticAdapter) setCode(address common.Address, code []byte) {
	acc := ga.getAccount(address)
	acc.SetCode(code)
	ga.Cache.UpdateAccount(acc)
}

func (ga *GallacticAdapter) getAccount(ethAddr common.Address) *account.Account {
	accAddr := fromEthAddress(ethAddr, false)
	acc, _ := ga.Cache.GetAccount(accAddr)
	if acc != nil {
		return acc
	}

	ctrAddr := fromEthAddress(ethAddr, true)
	ctr, _ := ga.Cache.GetAccount(ctrAddr)
	if ctr != nil {
		return ctr
	}

	fmt.Printf("Not such a address: %s, %s\n", accAddr, ctrAddr)
	return nil
}

func (ga *GallacticAdapter) LastBlockNumber() *big.Int {
	return bigint(ga.BlockChain.LastBlockHeight())
}

func (ga *GallacticAdapter) LastBlockHash() []byte {
	return ga.BlockChain.LastBlockHash()
}

func (ga *GallacticAdapter) TimeStamp() uint64 {
	return uint64(ga.BlockChain.LastBlockTime().Unix())
}

func (ga *GallacticAdapter) ConvertLog(log sputnikvm.Log) evm.Log {
	var l evm.Log
	for _, t := range log.Topics {
		l.Topics = append(l.Topics, t.Bytes())

	}
	l.Address = fromEthAddress(log.Address, true)
	l.Data = log.Data
	return l
}

func toEthAddress(addr crypto.Address) common.Address {
	var ethAddr common.Address
	ethAddr.SetBytes(addr.RawBytes()[2:22])
	return ethAddr
}

func fromEthAddress(ethAdr common.Address, contract bool) crypto.Address {
	var addr crypto.Address

	if contract {
		addr, _ = crypto.ContractAddress(ethAdr.Bytes())
	} else {
		if bytes.Equal(ethAdr.Bytes(), crypto.GlobalAddress.RawBytes()[2:22]) {
			addr = crypto.GlobalAddress
		} else {
			addr, _ = crypto.AccountAddress(ethAdr.Bytes())
		}
	}

	return addr
}

func bigint(val uint64) *big.Int {
	return new(big.Int).SetUint64(val)
}
