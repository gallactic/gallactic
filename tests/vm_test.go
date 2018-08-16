package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"

	"github.com/gallactic/gallactic/core/evm/burrow"
	. "github.com/hyperledger/burrow/execution/evm/asm"
	. "github.com/hyperledger/burrow/execution/evm/asm/bc"

	"github.com/stretchr/testify/require"
)

var defaultGas uint64 = 100000

func callAndCheck(t *testing.T, errorCode int, contractCode []byte, contractBalance uint64, bytecode, input []byte, value, gas uint64) (output []byte, err error) {
	caller := getAccountByName(t, "vbuterin")
	callee, _ := makeContractAccount(t, contractCode, contractBalance, permission.Call)

	caller.SetCode(bytecode)

	start := time.Now()
	output, err = burrow.CallCode(tBC, caller, callee, input, value, 0, 2100, &gas)
	fmt.Printf("Output: %v Error: %v\n", output, err)
	fmt.Println("Call took:", time.Since(start))
	if errorCode != e.ErrNone {
		require.Equal(t, e.Code(err), errorCode)
	} else {
		require.NoError(t, err)
	}
	return output, err
}

// convenience function for contract that calls a given address
func callContractCode(contractAddr crypto.Address, amt byte) []byte {
	// calldatacopy into mem and use as input to call
	memOff, inputOff := byte(0x0), byte(0x0)
	value := amt /// amount to transfer
	inOff := byte(0x0)
	retOff, retSize := byte(0x0), byte(0x20)

	// this is the code we want to run (call a contract and return)
	return MustSplice(CALLDATASIZE, PUSH1, inputOff, PUSH1, memOff,
		CALLDATACOPY, PUSH1, retSize, PUSH1, retOff, CALLDATASIZE, PUSH1, inOff,
		PUSH1, value, PUSH20, contractAddr.RawBytes(),
		// Zeno loves us - call with half of the available gas each time we CALL
		PUSH1, 2, GAS, DIV, CALL,
		PUSH1, 32, PUSH1, 0, RETURN)
}

// wrap a contract in create code
func wrapContractForCreateCode(contractCode []byte) []byte {
	// the is the code we need to return the contractCode when the contract is initialized
	lenCode := len(contractCode)
	// push code to the stack
	code := append([]byte{0x7f}, binary.RightPadWord256(contractCode).Bytes()...)
	// store it in memory
	code = append(code, []byte{0x60, 0x0, 0x52}...)
	// return whats in memory
	code = append(code, []byte{0x60, byte(lenCode), 0x60, 0x0, 0xf3}...)
	// return init code, contract code, expected return
	return code
}

// convenience function for contract that is a factory for the code that comes as call data
func createContractCode() []byte {
	// TODO: gas ...

	// calldatacopy the calldatasize
	memOff, inputOff := byte(0x0), byte(0x0)
	contractCode := []byte{0x60, memOff, 0x60, inputOff, 0x36, 0x37}

	// create
	value := byte(0x1)
	contractCode = append(contractCode, []byte{0x60, value, 0x36, 0x60, memOff, 0xf0}...)
	return contractCode
}
