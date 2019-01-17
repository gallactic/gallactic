package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gallactic/gallactic/core/consensus/tendermint/codes"
	"github.com/gallactic/gallactic/events"
	"github.com/gallactic/gallactic/txs"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	abciTypes "github.com/tendermint/tendermint/abci/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

const (
	blockingTimeout = 100 * time.Second
)

// Transactor is the controller/middleware for the v0 RPC
type Transactor struct {
	broadcastTxAsync func(tx tmTypes.Tx, cb func(*abciTypes.Response)) error
	logger           *logging.Logger
	eventBus         *events.EventBus
}

func NewTransactor(broadcastTxAsync func(tx tmTypes.Tx, cb func(*abciTypes.Response)) error,
	eventBus *events.EventBus, logger *logging.Logger) *Transactor {

	return &Transactor{
		broadcastTxAsync: broadcastTxAsync,
		eventBus:         eventBus,
		logger:           logger.With(structure.ComponentKey, "Transactor"),
	}
}

func (trans *Transactor) BroadcastTxAsyncRaw(txBytes []byte, callback func(res *abciTypes.Response)) error {
	return trans.broadcastTxAsync(txBytes, callback)
}

func (trans *Transactor) BroadcastTxAsync(txEnv *txs.Envelope, callback func(res *abciTypes.Response)) error {
	txBytes, err := txEnv.Encode()
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

	txBytes, err := txEnv.Encode()
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

func (trans *Transactor) BroadcastTxSync(txEnv *txs.Envelope, callback func(res *abciTypes.Response)) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), blockingTimeout)
	defer cancel()
	txBytes, err := txEnv.Encode()
	if err != nil {
		return 0, fmt.Errorf("error encoding transaction: %v", err)
	}

	// Subscribe before submitting to mempool
	txHash := txEnv.Hash()
	subID := events.GenSubID()
	out := make(chan interface{}, 1)
	q := events.QueryForTxExecution(txHash)
	err = trans.eventBus.Subscribe(ctx, subID, q, out)

	if err != nil {
		// We do not want to hold the lock with a defer so we must
		return 0, err
	}
	defer trans.eventBus.UnsubscribeAll(ctx, subID)

	txReceipt, err := trans.BroadcastTxRaw(txBytes)
	if err != nil {
		return 0, err
	}
	fmt.Print(txReceipt)

	return 0, nil

	// Get all the execution events for this Tx
	// select {
	// case <-ctx.Done():
	// 	return nil, ctx.Err()
	// case <-timer.C:
	// 	return nil, fmt.Errorf("timed out waiting for transaction with hash %v timed out after %v",
	// 		checkTxReceipt.TxHash, BlockingTimeout)
	// case msg := <-out:
	// 	txe := msg.(*exec.TxExecution)
	// 	callError := txe.CallError()
	// 	if callError != nil && callError.ErrorCode() != errors.ErrorCodeExecutionReverted {
	// 		return nil, errors.Wrap(callError, "exception during transaction execution")
	// 	}
	// 	return txe, nil
	// }
}
