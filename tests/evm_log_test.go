package tests

import (
	"encoding/hex"
	"testing"

	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/txs"

	"github.com/gallactic/gallactic/crypto"
	e "github.com/gallactic/gallactic/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
 * Scenario: Alice Deploy Contract and Bob Call Contract Event Method
 * Test: Event Log must be recorded, caller and contract address recorded must match
 */
func TestCallEventMethod1(t *testing.T) {
	setPermissions(t, "alice", permission.CreateContract|permission.Call)
	setPermissions(t, "bob", permission.Call)

	//Test Smart Contract
	/*
		pragma solidity ^0.4.19;
		contract ExampleContract { event ReturnValue(address indexed _from, int256 _value); function foo (int256 _value) public returns (int256) { ReturnValue(msg.sender, _value); return _value; } }
	*/
	// This byte code is compiled from Solidity contract above using remix.ethereum.org online compiler
	code, err := hex.DecodeString("6080604052348015600f57600080fd5b5060d28061001e6000396000f300608060405260043610603e5763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416634c970b2f81146043575b600080fd5b348015604e57600080fd5b506058600435606a565b60408051918252519081900360200190f35b60408051828152905160009133917f8c9591f069b095a3c90b9bc514b93f273d8ae392658e13676958539141377da29181900360200190a250905600a165627a7a72305820f305ceb0fb7128776960a15b6333b755c341609f417383f090abeb76ab4cbff10029")

	require.NoError(t, err)

	// A single input, having the permission, should succeed
	var _fee uint64 = 10
	seq1 := getAccountByName(t, "alice").Sequence()
	// alice deploy contract
	tx1 := makeCallTx(t, "alice", crypto.Address{}, code, 0, _fee)
	_, rec1 := signAndExecute(t, e.ErrNone, tx1, "alice")
	contractAddr := *rec1.ContractAddress
	assert.Equal(t, rec1.Status, txs.Ok)

	seq2 := getAccountByName(t, "alice").Sequence()
	// ensure the contract is there
	assert.Equal(t, seq2, seq1+1)

	// Input is the function hash of `foo()`: 120
	input2, _ := hex.DecodeString("4c970b2f0000000000000000000000000000000000000000000000000000000000000078")
	// bob call contract
	tx2 := makeCallTx(t, "bob", contractAddr, input2, 0, 100000)
	_, rec2 := signAndExecute(t, e.ErrNone, tx2, "bob")

	bob := getAccountByName(t, "bob")
	assert.Equal(t, rec2.Status, txs.Ok)
	// assert log should not be empty
	assert.NotNil(t, rec2.Logs)
	// assert log output data
	assert.Equal(t, rec2.Logs[0].Address, contractAddr)
	assert.Equal(t, rec2.Logs[0].Data.Bytes(), input2[4:])

	topic1 := rec2.Logs[0].Topics[1][12:]
	assert.Equal(t, topic1.Bytes(), bob.Address().RawBytes()[2:22])
}
