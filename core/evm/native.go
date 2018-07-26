// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package evm

import (
	"crypto/sha256"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/hyperledger/burrow/logging"
	"golang.org/x/crypto/ripemd160"
)

var registeredNativeContracts = make(map[binary.Word256]NativeContract)

func IsRegisteredNativeContract(address binary.Word256) bool {
	_, ok := registeredNativeContracts[address]
	return ok
}

func RegisterNativeContract(addr binary.Word256, fn NativeContract) bool {
	_, exists := registeredNativeContracts[addr]
	if exists {
		return false
	}
	registeredNativeContracts[addr] = fn
	return true
}

func init() {
	registerNativeContracts()
	registerSNativeContracts()
}

func registerNativeContracts() {
	// registeredNativeContracts[Int64ToWord256(1)] = ecrecoverFunc
	registeredNativeContracts[binary.Int64ToWord256(2)] = sha256Func
	registeredNativeContracts[binary.Int64ToWord256(3)] = ripemd160Func
	registeredNativeContracts[binary.Int64ToWord256(4)] = identityFunc
}

//-----------------------------------------------------------------------------

func ExecuteNativeContract(addr binary.Word256, st *state.State, caller *account.Account, input []byte, gas *uint64,
	logger *logging.Logger) ([]byte, error) {
	contract, ok := registeredNativeContracts[addr]
	if !ok {
		addr, _ := crypto.AddressFromWord256(addr)
		return nil, e.Errorf(e.ErrNativeFunction, "no native contract registered at address: %v", addr)
	}
	output, err := contract(st, caller, input, gas, logger)
	if err != nil {
		return nil, e.Errorf(e.ErrNativeFunction, err.Error())
	}
	return output, nil
}

type NativeContract func(st *state.State, caller *account.Account, input []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error)

func sha256Func(st *state.State, caller *account.Account, input []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {
	// Deduct gas
	gasRequired := uint64((len(input)+31)/32)*GasSha256Word + GasSha256Base
	if *gas < gasRequired {
		return nil, e.Error(e.ErrInsufficientGas)
	} else {
		*gas -= gasRequired
	}
	// Hash
	hasher := sha256.New()
	// CONTRACT: this does not err
	hasher.Write(input)
	return hasher.Sum(nil), nil
}

func ripemd160Func(st *state.State, caller *account.Account, input []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {
	// Deduct gas
	gasRequired := uint64((len(input)+31)/32)*GasRipemd160Word + GasRipemd160Base
	if *gas < gasRequired {
		return nil, e.Error(e.ErrInsufficientGas)
	} else {
		*gas -= gasRequired
	}
	// Hash
	hasher := ripemd160.New()
	// CONTRACT: this does not err
	hasher.Write(input)
	return binary.LeftPadBytes(hasher.Sum(nil), 32), nil
}

func identityFunc(st *state.State, caller *account.Account, input []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {
	// Deduct gas
	gasRequired := uint64((len(input)+31)/32)*GasIdentityWord + GasIdentityBase
	if *gas < gasRequired {
		return nil, e.Error(e.ErrInsufficientGas)
	}

	*gas -= gasRequired

	// Return identity
	return input, nil
}
