package rpc

import (
	"encoding/json"
	"testing"

	"fmt"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResultBroadcastTx(t *testing.T) {
	result := BroadcastTxOutput{
		Receipt: txs.Receipt{
			TxHash: []byte("foo"),
		},
	}

	jsonResult, err := json.Marshal(result)
	require.NoError(t, err)
	assert.Equal(t, `{"TxHash":"Zm9v"}`, string(jsonResult))
}

func TestListUnconfirmedTxs(t *testing.T) {

	caller, err := crypto.AddressFromString("ac8KfZqAKYayEWsc6vuwfLu5GDBaCUvoH8B")
	require.NoError(t, err)

	callee, err := crypto.AddressFromString("acTqSGVw94xP1myXrnCm3rBWgzcJ5uEbB1f")
	require.NoError(t, err)

	callTx, err := tx.NewCallTx(caller, callee, 1, nil, 1, 100, 12)
	fmt.Println("CallTx :\n", callTx)

	result := &UnconfirmedTxsOutput{
		Count: 1,
		Txs: []*txs.Envelope{
			txs.Enclose("testChain", callTx),
		},
	}

	jsonResult, err := json.Marshal(result)
	require.NoError(t, err)
	expected := "{\"Count\":1,\"Txs\":[{\"chainId\":\"testChain\",\"type\":\"CallTx\",\"tx\":{\"caller\":{\"address\":\"ac8KfZqAKYayEWsc6vuwfLu5GDBaCUvoH8B\",\"amount\":112,\"sequence\":1},\"callee\":{\"address\":\"acTqSGVw94xP1myXrnCm3rBWgzcJ5uEbB1f\",\"amount\":100},\"gas_limit\":1}}]}"
	assert.Equal(t, expected, string(jsonResult))
}

func TestResultListAccounts(t *testing.T) {
	acc := account.NewAccountFromSecret("This is sercret!")
	result := AccountsOutput{
		Accounts:    []*account.Account{acc},
		BlockHeight: 2,
	}

	jsonResult, err := json.Marshal(result)
	require.NoError(t, err)
	resultOut := new(AccountsOutput)
	json.Unmarshal(jsonResult, resultOut)
	jsonResultOut, err := json.Marshal(resultOut)
	require.NoError(t, err)
	assert.Equal(t, string(jsonResult), string(jsonResultOut))
}

/*
func TestResultCall_MarshalJSON(t *testing.T) {
	res := ResultCall{
		Call: execution.Call{
			Return:  []byte("hi"),
			GasUsed: 1,
		},
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)

	resOut := new(ResultCall)
	json.Unmarshal(bs, resOut)
	bsOut, err := json.Marshal(resOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}

/*
func TestResultEvent(t *testing.T) {
	eventDataNewBlock := tmTypes.EventDataNewBlock{
		Block: &tmTypes.Block{
			Header: &tmTypes.Header{
				ChainID: "chainy",
				Count:  30,
			},
			LastCommit: &tmTypes.Commit{
				Precommits: []*tmTypes.Vote{
					{
						Signature: tmCrypto.SignatureEd25519{1, 2, 3},
					},
				},
			},
		},
	}
	res := ResultEvent{
		Tendermint: &TendermintEvent{
			TMEventData: &eventDataNewBlock,
		},
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)

	resOut := new(ResultEvent)
	require.NoError(t, json.Unmarshal(bs, resOut))
	bsOut, err := json.Marshal(resOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
	//fmt.Println(string(bs))
	//fmt.Println(string(bsOut))
}

func TestResultGetBlock(t *testing.T) {
	res := &GetBlockOutput{
		Block: &Block{&tmTypes.Block{
			LastCommit: &tmTypes.Commit{
				Precommits: []*tmTypes.Vote{
					{
						Signature: tmCrypto.SignatureEd25519{1, 2, 3},
					},
				},
			},
		},
		},
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)
	resOut := new(GetBlockOutput)
	require.NoError(t, json.Unmarshal([]byte(bs), resOut))
	bsOut, err := json.Marshal(resOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}

func TestResultDumpConsensusState(t *testing.T) {
	res := &DumpConsensusStateOutput{
		RoundState: types.RoundStateSimple{
			HeightRoundStep: "34/0/3",
			Votes:           json.RawMessage(`[{"i'm a json": "32"}]`),
			LockedBlockHash: common.HexBytes{'b', 'y', 't', 'e', 's'},
		},
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)
	resOut := new(DumpConsensusStateOutput)
	require.NoError(t, json.Unmarshal([]byte(bs), resOut))
	bsOut, err := json.Marshal(resOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}

func TestResultLastBlockInfo(t *testing.T) {
	res := &GetLastBlockInfoOutput{
		LastBlockTime:   time.Now(),
		LastBlockHash:   binary.HexBytes{3, 4, 5, 6},
		LastBlockHeight: 2343,
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)
	fmt.Println(string(bs))

}
*/
