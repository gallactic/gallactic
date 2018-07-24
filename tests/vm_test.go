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

/*


func callAndCheck(t *testing.T, errorCode int, contractCode []byte, contractBalance uint64, bytecode, input []byte, value, gas uint64) (output []byte, err error) {
	caller := getAccount(t, "vbuterin")
	callee, _ := makeContractAccount(t, contractCode, contractBalance, permission.Call)

	start := time.Now()
	output, err = evm1.Call(evm1Cache, caller, callee, bytecode, input, value, &gas)
	fmt.Printf("Output: %v Error: %v\n", output, err)
	fmt.Println("Call took:", time.Since(start))
	if errorCode != e.ErrNone {
		require.Equal(t, e.Code(err), errorCode)
	} else {
		require.NoError(t, err)
	}
	return output, err
}

// Subscribes to an AccCall, runs the vm, returns the output any direct exception
// and then waits for any exceptions transmitted by Data in the AccCall
// event (in the case of no direct error from call we will block waiting for
// at least 1 AccCall event)
func runVMWaitError(t *testing.T, caller, callee *acm.Account, subscribeAddr crypto.Address,
	contractCode []byte, gas uint64) ([]byte, error) {
	eventCh := make(chan *events.EventDataCall)
	output, err := runVM(t, eventCh, caller, callee, subscribeAddr, contractCode, gas)
	if err != nil {
		return output, err
	}
	select {
	case eventDataCall := <-eventCh:
		if eventDataCall.Exception != nil {
			return output, eventDataCall.Exception
		}
		return output, nil
	}
}

// Subscribes to an AccCall, runs the vm, returns the output and any direct
// exception
func runVM(t *testing.T, eventCh chan<- *events.EventDataCall, caller, callee *acm.Account,
	subscribeAddr crypto.Address, contractCode []byte, gas uint64) ([]byte, error) {

	// we need to catch the event from the CALL to check for exceptions
	emitter := event.NewEmitter(logging.NewNoopLogger())
	fmt.Printf("subscribe to %s\n", subscribeAddr)

	err := events.SubscribeAccountCall(context.Background(), emitter, "test", subscribeAddr,
		nil, -1, eventCh)
	if err != nil {
		return nil, err
	}
	evc := event.NewCache()
	evm1.SetPublisher(evc)
	start := time.Now()
	output, err := evm1.Call(evm1Cache, caller, callee, contractCode, []byte{}, 0, &gas)
	fmt.Printf("Output: %v Error: %v\n", output, err)
	fmt.Println("Call took:", time.Since(start))
	evc.Flush(emitter)
	return output, err
}

// convenience function for contract that calls a given address
func callContractCode(contractAddr crypto.Address, amount byte) []byte {
	// calldatacopy into mem and use as input to call
	memOff, inputOff := byte(0x0), byte(0x0)
	value := amount /// amount to ransffer
	inOff := byte(0x0)
	retOff, retSize := byte(0x0), byte(0x20)

	// this is the code we want to run (call a contract and return)
	return MustSplice(CALLDATASIZE, PUSH1, inputOff, PUSH1, memOff,
		CALLDATACOPY, PUSH1, retSize, PUSH1, retOff, CALLDATASIZE, PUSH1, inOff,
		PUSH1, value, PUSH20, contractAddr,
		// Zeno loves us - call with half of the available gas each time we CALL
		PUSH1, 2, GAS, DIV, CALL,
		PUSH1, 32, PUSH1, 0, RETURN)
}

// wrap a contract in create code
func wrapContractForCreateCode(contractCode []byte) []byte {
	// the is the code we need to return the contractCode when the contract is initialized
	lenCode := len(contractCode)
	// push code to the stack
	code := append([]byte{0x7f}, RightPadWord256(contractCode).Bytes()...)
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

// Runs a basic loop
func TestVM(t *testing.T) {
	bytecode := MustSplice(PUSH1, 0x00, PUSH1, 0x20, MSTORE, JUMPDEST, PUSH2, 0x0F, 0x0F, PUSH1, 0x20, MLOAD,
		SLT, ISZERO, PUSH1, 0x1D, JUMPI, PUSH1, 0x01, PUSH1, 0x20, MLOAD, ADD, PUSH1, 0x20,
		MSTORE, PUSH1, 0x05, JUMP, JUMPDEST)

	callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
}

func TestSHL(t *testing.T) {
	//Shift left 0
	bytecode := MustSplice(PUSH1, 0x01, PUSH1, 0x00, SHL, return1())

	output, _ := callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value := []uint8([]byte{0x1})
	expected := LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift left 0
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x00, SHL, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Shift left 1
	bytecode = MustSplice(PUSH1, 0x01, PUSH1, 0x01, SHL, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x2})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift left 1
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x01, SHL, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift left 1
	bytecode = MustSplice(PUSH32, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x01, SHL, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Shift left 255
	bytecode = MustSplice(PUSH1, 0x01, PUSH1, 0xFF, SHL, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x80})
	expected = RightPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift left 255
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0xFF, SHL, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x80})
	expected = RightPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Shift left 256 (overflow)
	bytecode = MustSplice(PUSH1, 0x01, PUSH2, 0x01, 0x00, SHL, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift left 256 (overflow)
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH2, 0x01, 0x00, SHL,
		return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Shift left 257 (overflow)
	bytecode = MustSplice(PUSH1, 0x01, PUSH2, 0x01, 0x01, SHL, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)
}

func TestSHR(t *testing.T) {
	//Shift right 0
	bytecode := MustSplice(PUSH1, 0x01, PUSH1, 0x00, SHR, return1())
	output, _ := callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value := []uint8([]byte{0x1})
	expected := LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift right 0
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x00, SHR, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Shift right 1
	bytecode = MustSplice(PUSH1, 0x01, PUSH1, 0x01, SHR, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift right 1
	bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1, 0x01, SHR, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x40})
	expected = RightPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift right 1
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x01, SHR, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	expected = []uint8([]byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Shift right 255
	bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1, 0xFF, SHR, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x1})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift right 255
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0xFF, SHR, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x1})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Shift right 256 (underflow)
	bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH2, 0x01, 0x00, SHR,
		return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift right 256 (underflow)
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH2, 0x01, 0x00, SHR,
		return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Shift right 257 (underflow)
	bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH2, 0x01, 0x01, SHR,
		return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)
}

func TestSAR(t *testing.T) {
	//Shift arith right 0
	bytecode := MustSplice(PUSH1, 0x01, PUSH1, 0x00, SAR, return1())
	output, _ := callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value := []uint8([]byte{0x1})
	expected := LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative arith shift right 0
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x00, SAR, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Shift arith right 1
	bytecode = MustSplice(PUSH1, 0x01, PUSH1, 0x01, SAR, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift arith right 1
	bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1, 0x01, SAR, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0xc0})
	expected = RightPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift arith right 1
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x01, SAR, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Shift arith right 255
	bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1, 0xFF, SAR, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift arith right 255
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0xFF, SAR, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift arith right 255
	bytecode = MustSplice(PUSH32, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0xFF, SAR, return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x00})
	expected = RightPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Shift arith right 256 (reset)
	bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH2, 0x01, 0x00, SAR,
		return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Alternative shift arith right 256 (reset)
	bytecode = MustSplice(PUSH32, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH2, 0x01, 0x00, SAR,
		return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	//Shift arith right 257 (reset)
	bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH2, 0x01, 0x01, SAR,
		return1())
	output, _ = callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)
}

//Test attempt to jump to bad destination (position 16)
func TestJumpErr(t *testing.T) {
	bytecode := MustSplice(PUSH1, 0x10, JUMP)

	var err error
	ch := make(chan struct{})
	go func() {
		_, err = callAndCheck(t, e.ErrVMInvalidJumpDest, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
		ch <- struct{}{}
	}()
	tick := time.NewTicker(time.Second * 2)
	select {
	case <-tick.C:
		t.Fatal("VM ended up in an infinite loop from bad jump dest (it took too long!)")
	case <-ch:
		if err == nil {
			t.Fatal("Expected invalid jump dest err")
		}
	}
}

// Tests the code for a subcurrency contract compiled by serpent
func TestSubcurrency(t *testing.T) {
	bytecode := MustSplice(PUSH3, 0x0F, 0x42, 0x40, CALLER, SSTORE, PUSH29, 0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1,
		0x00, CALLDATALOAD, DIV, PUSH4, 0x15, 0xCF, 0x26, 0x84, DUP2, EQ, ISZERO, PUSH2,
		0x00, 0x46, JUMPI, PUSH1, 0x04, CALLDATALOAD, PUSH1, 0x40, MSTORE, PUSH1, 0x40,
		MLOAD, SLOAD, PUSH1, 0x60, MSTORE, PUSH1, 0x20, PUSH1, 0x60, RETURN, JUMPDEST,
		PUSH4, 0x69, 0x32, 0x00, 0xCE, DUP2, EQ, ISZERO, PUSH2, 0x00, 0x87, JUMPI, PUSH1,
		0x04, CALLDATALOAD, PUSH1, 0x80, MSTORE, PUSH1, 0x24, CALLDATALOAD, PUSH1, 0xA0,
		MSTORE, CALLER, SLOAD, PUSH1, 0xC0, MSTORE, CALLER, PUSH1, 0xE0, MSTORE, PUSH1,
		0xA0, MLOAD, PUSH1, 0xC0, MLOAD, SLT, ISZERO, ISZERO, PUSH2, 0x00, 0x86, JUMPI,
		PUSH1, 0xA0, MLOAD, PUSH1, 0xC0, MLOAD, SUB, PUSH1, 0xE0, MLOAD, SSTORE, PUSH1,
		0xA0, MLOAD, PUSH1, 0x80, MLOAD, SLOAD, ADD, PUSH1, 0x80, MLOAD, SSTORE, JUMPDEST,
		JUMPDEST, POP, JUMPDEST, PUSH1, 0x00, RETURN)

	data, _ := hex.DecodeString("693200CE0000000000000000000000004B4363CDE27C2EB05E66357DB05BC5C88F850C1A0000000000000000000000000000000000000000000000000000000000000005")
	output, _ := callAndCheck(t, e.ErrNone, []byte{}, 0, bytecode, data, 0, defaultGas)
	fmt.Println(output)
}

//This test case is taken from EIP-140 (https://github.com/ethereum/EIPs/blob/master/EIPS/eip-140.md);
//it is meant to test the implementation of the REVERT opcode
func TestRevert(t *testing.T) {
	key, value := []byte{0x00}, []byte{0x00}
	evm1Cache.SetStorage(accountPool["satoshi"].Address(), LeftPadWord256(key), LeftPadWord256(value))

	bytecode := MustSplice(PUSH13, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x65, 0x64, 0x20, 0x64, 0x61, 0x74, 0x61,
		PUSH1, 0x00, SSTORE, PUSH32, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		PUSH1, 0x00, MSTORE, PUSH1, 0x0E, PUSH1, 0x00, REVERT)

	// bytecode := MustSplice(PUSH32, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61,
	// 0x67, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// 0x00, 0x00, 0x00, PUSH1, 0x00, MSTORE, PUSH1, 0x0E, PUSH1, 0x00, REVERT)

	output, _ := callAndCheck(t, e.ErrVMExecutionReverted, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)

	storageVal, err := evm1Cache.GetStorage(accountPool["satoshi"].Address(), LeftPadWord256(key))
	assert.Equal(t, LeftPadWord256(value), storageVal)

	t.Logf("Output: %v Error: %v\n", output, err)
}

// Test sending tokens from a contract to another account
func TestSendCall(t *testing.T) {
	// Create accounts
	account1, _ := makeAccount(t, 0, permission.Call)
	account2, _ := makeAccount(t, 0, permission.Call)
	account3, _ := makeAccount(t, 0, 0)

	// account1 will call account2 which will trigger CALL opcode to account3
	addr := account3.Address()
	contractCode := callContractCode(addr, 100)

	//----------------------------------------------
	// account2 has insufficient balance, should fail
	_, err := runVMWaitError(t, account1, account2, addr, contractCode, 1000)
	assert.Error(t, err, "Expected insufficient balance error")

	//----------------------------------------------
	// give account2 sufficient balance, should pass
	err = account2.AddToBalance(100000)
	require.NoError(t, err)
	_, err = runVMWaitError(t, account1, account2, addr, contractCode, 1000)
	assert.NoError(t, err, "Should have sufficient balance")

	//----------------------------------------------
	// insufficient gas, should fail
	err = account2.AddToBalance(100000)
	require.NoError(t, err)
	_, err = runVMWaitError(t, account1, account2, addr, contractCode, 100)
	assert.NoError(t, err, "Expected insufficient gas error")
}

// This test was introduced to cover an issues exposed in our handling of the
// gas limit passed from caller to callee on various forms of CALL.
// The idea of this test is to implement a simple DelegateCall in EVM code
// We first run the DELEGATECALL with _just_ enough gas expecting a simple return,
// and then run it with 1 gas unit less, expecting a failure
func TestDelegateCallGas(t *testing.T) {
	inOff := 0
	inSize := 0 // no call data
	retOff := 0
	retSize := 32
	calleeReturnValue := int64(20)

	// DELEGATECALL(retSize, refOffset, inSize, inOffset, addr, gasLimit)
	// 6 pops
	delegateCallCost := evm.GasStackOp * 6
	// 1 push
	gasCost := evm.GasStackOp
	// 2 pops, 1 push
	subCost := evm.GasStackOp * 3
	pushCost := evm.GasStackOp

	costBetweenGasAndDelegateCall := gasCost + subCost + delegateCallCost + pushCost

	// Do a simple operation using 1 gas unit
	code := MustSplice(PUSH1, calleeReturnValue, return1())
	calleeAccount, calleeAddress := makeContractAccount(t, code, 10000, permission.Call)

	// Here we split up the caller code so we can make a DELEGATE call with
	// different amounts of gas. The value we sandwich in the middle is the amount
	// we subtract from the available gas (that the caller has available), so:
	// code := MustSplice(callerCodePrefix, <amount to subtract from GAS> , callerCodeSuffix)
	// gives us the code to make the call
	callerCodePrefix := MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize,
		PUSH1, inOff, PUSH20, calleeAddress, PUSH1)
	callerCodeSuffix := MustSplice(GAS, SUB, DELEGATECALL, returnWord())

	// Perform a delegate call
	code = MustSplice(callerCodePrefix,
		// Give just enough gas to make the DELEGATECALL
		costBetweenGasAndDelegateCall,
		callerCodeSuffix)
	callerAccount, _ := makeContractAccount(t, code, 10000, permission.ZeroPermissions)

	// Should pass
	output, err := runVMWaitError(t, callerAccount, calleeAccount, calleeAddress,
		callerAccount.Code(), 100)
	assert.NoError(t, err, "Should have sufficient funds for call")
	assert.Equal(t, Int64ToWord256(calleeReturnValue).Bytes(), output)

	callerAccount.SetCode(MustSplice(callerCodePrefix,
		// Shouldn't be enough gas to make call
		costBetweenGasAndDelegateCall-1,
		callerCodeSuffix))

	// Should fail
	_, err = runVMWaitError(t, callerAccount, calleeAccount, calleeAddress, callerAccount.Code(), 100)
	assert.Equal(t, e.Code(err), e.ErrVMInsufficientGas, "Should have insufficient gas for call")
}

func TestMemoryBounds(t *testing.T) {

	memoryProvider := func() evm.Memory {
		return evm.NewDynamicMemory(1024, 2048)
	}
	ourVM := evm.NewVM(newParams(), crypto.ZeroAddress, nil, nopLogger, evm.MemoryProvider(memoryProvider))
	caller, _ := makeContractAccount(t, nil, 10000, permission.ZeroPermissions)
	callee, _ := makeContractAccount(t, nil, 10000, permission.ZeroPermissions)

	// This attempts to store a value at the memory boundary and return it
	word := One256
	output, err := ourVM.Call(evm1Cache, caller, callee,
		MustSplice(pushWord(word), storeAtEnd(), MLOAD, storeAtEnd(), returnAfterStore()),
		nil, 0, &defaultGas)
	assert.NoError(t, err)
	assert.Equal(t, word.Bytes(), output)

	// Same with number
	word = Int64ToWord256(232234234432)
	output, err = ourVM.Call(evm1Cache, caller, callee,
		MustSplice(pushWord(word), storeAtEnd(), MLOAD, storeAtEnd(), returnAfterStore()),
		nil, 0, &defaultGas)
	assert.NoError(t, err)
	assert.Equal(t, word.Bytes(), output)

	// Now test a series of boundary stores
	code := pushWord(word)
	for i := 0; i < 10; i++ {
		code = MustSplice(code, storeAtEnd(), MLOAD)
	}
	output, err = ourVM.Call(evm1Cache, caller, callee, MustSplice(code, storeAtEnd(), returnAfterStore()),
		nil, 0, &defaultGas)
	assert.NoError(t, err)
	assert.Equal(t, word.Bytes(), output)

	// Same as above but we should breach the upper memory limit set in memoryProvider
	code = pushWord(word)
	for i := 0; i < 100; i++ {
		code = MustSplice(code, storeAtEnd(), MLOAD)
	}
	output, err = ourVM.Call(evm1Cache, caller, callee, MustSplice(code, storeAtEnd(), returnAfterStore()),
		nil, 0, &defaultGas)
	assert.Error(t, err, "Should hit memory out of bounds")
}

func TestMsgSender(t *testing.T) {
	//
	//	pragma solidity ^0.4.0;
	//
	//	contract SimpleStorage {
	//     function get() public constant returns (address) {
	//	      return msg.sender;
	//	   }
	//  }
	//

	// This bytecode is compiled from Solidity contract above using remix.ethereum.org online compiler
	code, err := hex.DecodeString("6060604052341561000f57600080fd5b60ca8061001d6000396000f30060606040526004361060" +
		"3f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680636d4ce63c14604457" +
		"5b600080fd5b3415604e57600080fd5b60546096565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ff" +
		"ffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6000339050905600a165627a" +
		"7a72305820b9ebf49535372094ae88f56d9ad18f2a79c146c8f56e7ef33b9402924045071e0029")
	require.NoError(t, err)

	// Run the contract initialisation code to obtain the contract code that would be mounted at account2
	contractCode, _ := callAndCheck(t, e.ErrNone, []byte{}, 0, code, code, 0, defaultGas)

	// Not needed for this test (since contract code is passed as argument to vm), but this is what an execution
	// framework must do

	// Input is the function hash of `get()`
	input, err := hex.DecodeString("6d4ce63c")

	output, _ := callAndCheck(t, e.ErrNone, contractCode, 0, contractCode, input, 0, defaultGas)

	assert.Equal(t, getAccount(t, "vbuterin").Address().Word256().Bytes(), output)

}

func TestInvalid(t *testing.T) {
	bytecode := MustSplice(PUSH32, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61,
		0x67, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, PUSH1, 0x00, MSTORE, PUSH1, 0x0E, PUSH1, 0x00, INVALID)

	output, err := callAndCheck(t, e.ErrVMExecutionAborted, []byte{}, 0, bytecode, []byte{}, 0, defaultGas)
	t.Logf("Output: %v Error: %v\n", output, err)
}

func TestReturnDataSize(t *testing.T) {
	callcode := MustSplice(PUSH32, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61,
		0x67, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, PUSH1, 0x00, MSTORE, PUSH1, 0x0E, PUSH1, 0x00, RETURN)

	accountName := "account2addresstests"
	address, _ := crypto.AddressFromBytes([]byte(accountName)) ///0x6163636F756E7432616464726573737465737473
	acc := acm.NewAccount(address)
	acc.SetPermissions(permission.Call)
	acc.SetCode(callcode)
	evm1Cache.UpdateAccount(acc)

	gas1, gas2 := byte(0x1), byte(0x1)
	value := byte(0x69)
	inOff, inSize := byte(0x0), byte(0x0) // no call data
	retOff, retSize := byte(0x0), byte(0x0E)

	bytecode := MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1, inOff, PUSH1, value, PUSH20,
		0x61, 0x63, 0x63, 0x6F, 0x75, 0x6E, 0x74, 0x32, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x74, 0x65,
		0x73, 0x74, 0x73, PUSH2, gas1, gas2, CALL, RETURNDATASIZE, PUSH1, 0x00, MSTORE, PUSH1, 0x20, PUSH1, 0x00, RETURN)

	expected := LeftPadBytes([]byte{0x0E}, 32)

	output, _ := callAndCheck(t, e.ErrNone, []byte{}, 1000, bytecode, []byte{}, 0, defaultGas)
	assert.Equal(t, expected, output)
}

func TestReturnDataCopy(t *testing.T) {
	callcode := MustSplice(PUSH32, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61,
		0x67, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, PUSH1, 0x00, MSTORE, PUSH1, 0x0E, PUSH1, 0x00, RETURN)

	accountName := "account2addresstests"
	address, _ := crypto.AddressFromBytes([]byte(accountName)) ///0x6163636F756E7432616464726573737465737473
	acc := acm.NewAccount(address)
	acc.SetCode(callcode)
	evm1Cache.UpdateAccount(acc)

	gas1, gas2 := byte(0x1), byte(0x1)
	value := byte(0x69)
	inOff, inSize := byte(0x0), byte(0x0) // no call data
	retOff, retSize := byte(0x0), byte(0x0E)

	bytecode := MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1, inOff, PUSH1, value, PUSH20,
		0x61, 0x63, 0x63, 0x6F, 0x75, 0x6E, 0x74, 0x32, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x74, 0x65,
		0x73, 0x74, 0x73, PUSH2, gas1, gas2, CALL, RETURNDATASIZE, PUSH1, 0x00, PUSH1, 0x00, RETURNDATACOPY,
		RETURNDATASIZE, PUSH1, 0x00, RETURN)

	expected := []byte{0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65}

	output, _ := callAndCheck(t, e.ErrNone, []byte{}, 1000, bytecode, []byte{}, 0, defaultGas)

	assert.Equal(t, expected, output)

}

// These code segment helpers exercise the MSTORE MLOAD MSTORE cycle to test
// both of the memory operations. Each MSTORE is done on the memory boundary
// (at MSIZE) which Solidity uses to find guaranteed unallocated memory.

// storeAtEnd expects the value to be stored to be on top of the stack, it then
// stores that value at the current memory boundary
func storeAtEnd() []byte {
	// Pull in MSIZE (to carry forward to MLOAD), swap in value to store, store it at MSIZE
	return MustSplice(MSIZE, SWAP1, DUP2, MSTORE)
}

func returnAfterStore() []byte {
	return MustSplice(PUSH1, 32, DUP2, RETURN)
}

// Store the top element of the stack (which is a 32-byte word) in memory
// and return it. Useful for a simple return value.
func return1() []byte {
	return MustSplice(PUSH1, 0, MSTORE, returnWord())
}

func returnWord() []byte {
	// PUSH1 => return size, PUSH1 => return offset, RETURN
	return MustSplice(PUSH1, 32, PUSH1, 0, RETURN)
}

func pushInt64(i int64) []byte {
	return pushWord(Int64ToWord256(i))
}

// Produce bytecode for a PUSH<N>, b_1, ..., b_N where the N is number of bytes
// contained in the unpadded word
func pushWord(word Word256) []byte {
	leadingZeros := byte(0)
	for leadingZeros < 32 {
		if word[leadingZeros] == 0 {
			leadingZeros++
		} else {
			return MustSplice(byte(PUSH32)-leadingZeros, word[leadingZeros:])
		}
	}
	return MustSplice(PUSH1, 0)
}
*/
