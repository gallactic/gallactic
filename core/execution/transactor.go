package execution

import (
	"encoding/json"
	"fmt"

	"github.com/gallactic/gallactic/core/consensus/tendermint/codes"
	"github.com/gallactic/gallactic/txs"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	abciTypes "github.com/tendermint/abci/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

// Transactor is the controller/middleware for the v0 RPC
type Transactor struct {
	broadcastTxAsync func(tx tmTypes.Tx, cb func(*abciTypes.Response)) error
	txEncoder        txs.Encoder
	logger           *logging.Logger
}

func NewTransactor(broadcastTxAsync func(tx tmTypes.Tx, cb func(*abciTypes.Response)) error, txEncoder txs.Encoder,
	logger *logging.Logger) *Transactor {

	return &Transactor{
		broadcastTxAsync: broadcastTxAsync,
		txEncoder:        txEncoder,
		logger:           logger.With(structure.ComponentKey, "Transactor"),
	}
}

func (trans *Transactor) BroadcastTxAsyncRaw(txBytes []byte, callback func(res *abciTypes.Response)) error {
	return trans.broadcastTxAsync(txBytes, callback)
}

func (trans *Transactor) BroadcastTxAsync(txEnv *txs.Envelope, callback func(res *abciTypes.Response)) error {
	txBytes, err := trans.txEncoder.EncodeTx(txEnv)
	if err != nil {
		return fmt.Errorf("error encoding transaction: %v", err)
	}
	return trans.BroadcastTxAsyncRaw(txBytes, callback)
}

// Broadcast a transaction and waits for a response from the mempool. Transactions to BroadcastTx will block during
// various mempool operations (managed by Tendermint) including mempool Reap, Commit, and recheckTx.
func (trans *Transactor) BroadcastTx(txEnv *txs.Envelope) (*txs.Receipt, error) {
	trans.logger.Trace.Log("method", "BroadcastTx",
		"tx_hash", txEnv.Hash(),
		"tx", txEnv.String())

	txBytes, err := trans.txEncoder.EncodeTx(txEnv)
	if err != nil {
		return nil, err
	}
	return trans.BroadcastTxRaw(txBytes)
}

func (trans *Transactor) BroadcastTxRaw(txBytes []byte) (*txs.Receipt, error) {
	responseCh := make(chan *abciTypes.Response, 1)
	err := trans.BroadcastTxAsyncRaw(txBytes, func(res *abciTypes.Response) {
		responseCh <- res
	})

	if err != nil {
		return nil, err
	}
	response := <-responseCh
	checkTxResponse := response.GetCheckTx()
	if checkTxResponse == nil {
		return nil, fmt.Errorf("application did not return CheckTx response")
	}

	switch checkTxResponse.Code {
	case codes.TxExecutionSuccessCode:
		receipt := new(txs.Receipt)
		err := json.Unmarshal(checkTxResponse.Data, receipt)
		if err != nil {
			return nil, fmt.Errorf("could not deserialise transaction receipt: %s", err)
		}
		return receipt, nil
	default:
		return nil, fmt.Errorf("error returned by Tendermint in BroadcastTxSync "+
			"ABCI code: %v, ABCI log: %v", checkTxResponse.Code, checkTxResponse.Log)
	}
}
