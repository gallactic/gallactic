package sputnikvm

import (
	"bytes"
	"math/big"

	"github.com/gallactic/gallactic/common/binary"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/blockchain"
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

func (ga *GallacticAdapter) GetGasLimit() uint64 {
	return ga.GasLimit
}

func (ga *GallacticAdapter) GetAmount() uint64 {
	return ga.Amount
}

func (ga *GallacticAdapter) GetData() []byte {
	return ga.Data
}

func (ga *GallacticAdapter) GetNonce() uint64 {
	return ga.Nonce
}


func (ga *GallacticAdapter) updateAccount(acc *account.Account) {
	ga.Cache.UpdateAccount(acc)
}

func (ga *GallacticAdapter) updateStorage(address common.Address, key *big.Int, value *big.Int) {
	_, addr := fromEthAddress(address, true)
	wKey := binary.LeftPadWord256(key.Bytes())
	wValue := binary.LeftPadWord256(value.Bytes())
	ga.Cache.SetStorage(addr, wKey, wValue)
}

func (ga *GallacticAdapter) getStorage(address common.Address, key *big.Int) (*big.Int, error) {
	_, addr := fromEthAddress(address, true)
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
	_, addr := fromEthAddress(address, true)
	acc, _ := account.NewContractAccount(addr)
	ga.Cache.UpdateAccount(acc)
	return acc
}

func (ga *GallacticAdapter) setCalleeAddress(address common.Address) {
	if ga.Callee == nil {
		ga.Callee = ga.getAccount(address)
	}
}

func (ga *GallacticAdapter) removeAccount(address common.Address) {
	//TODO: We should implement removeAccount() in cache
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

func (ga *GallacticAdapter) setAccount(address common.Address, balance uint64, code []byte, nonce uint64) {
	acc := ga.getAccount(address)
	acc.SetBalance(balance)
	acc.SetSequence(nonce)
	acc.SetCode(code)
	ga.Cache.UpdateAccount(acc)
}

func (ga *GallacticAdapter) log(address common.Address, topics []common.Hash, data []byte) {
	//TODO: We should handle events inside this method
}

func (ga *GallacticAdapter) getAccount(ethAddr common.Address) *account.Account {
	converted, addr := fromEthAddress(ethAddr, false)

	if converted {
		acc, _ := ga.Cache.GetAccount(addr)
		if acc != nil {
			return acc
		}
	}

	converted, addr = fromEthAddress(ethAddr, true)

	if converted {
		acc, _ := ga.Cache.GetAccount(addr)
		return acc
	}

	return nil
}

func (ga *GallacticAdapter) LastBlockNumber() uint64 {
	return ga.BlockChain.LastBlockHeight()
}

func (ga *GallacticAdapter) LastBlockHash() []byte {
	return ga.BlockChain.LastBlockHash()
}

func (ga *GallacticAdapter) TimeStamp() uint64 {
	return uint64(ga.BlockChain.LastBlockTime().Unix())
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
