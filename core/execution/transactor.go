package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gallactic/gallactic/core/consensus/tendermint/codes"
	"github.com/gallactic/gallactic/core/events"
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
	broadcastTxFunc func(tx tmTypes.Tx, cb func(*abciTypes.Response)) error
	logger          *logging.Logger
	eventBus        events.EventBus
}

func NewTransactor(broadcastTxFunc func(tx tmTypes.Tx, cb func(*abciTypes.Response)) error,
	eventBus events.EventBus, logger *logging.Logger) *Transactor {

	return &Transactor{
		broadcastTxFunc: broadcastTxFunc,
		eventBus:        eventBus,
		logger:          logger.With(structure.ComponentKey, "Transactor"),
	}
}

func (trans *Transactor) BroadcastTxSync(txEnv *txs.Envelope) (*txs.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), blockingTimeout)
	defer cancel()

	// Subscribe before submitting to mempool
	txHash := txEnv.Hash()
	subID := events.GenSubID()
	out := make(chan interface{}, 1)
	q := events.QueryForTx(txHash)

	if err := trans.eventBus.Subscribe(ctx, subID, q, out); err != nil {
		// We do not want to hold the lock with a defer so we must
		return nil, err
	}
	defer trans.eventBus.UnsubscribeAll(ctx, subID)

	receipt, err := trans.broadcastTxRaw(txEnv)
	if err != nil {
		return receipt, err
	}

	// Get all the execution events for this Tx
	select {
	case <-ctx.Done():
		return receipt, ctx.Err()

	case msg := <-out:
		receipt2 := msg.(*txs.Receipt)
		return receipt2, nil
	}
}

func (trans *Transactor) BroadcastTxAsync(txEnv *txs.Envelope) (*txs.Receipt, error) {
	trans.logger.Trace.Log("method", "BroadcastTxAsync",
		"tx_hash", txEnv.Hash(),
		"tx", txEnv.String())

	return trans.broadcastTxRaw(txEnv)
}

func (trans *Transactor) broadcastTxRaw(txEnv *txs.Envelope) (*txs.Receipt, error) {
	txBytes, err := txEnv.Encode()
	if err != nil {
		return nil, err
	}

	responseCh := make(chan *abciTypes.Response, 1)
	err = trans.broadcastTxFunc(txBytes, func(res *abciTypes.Response) {
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
		if err := json.Unmarshal(checkTxResponse.Data, receipt); err != nil {
			return nil, fmt.Errorf("could not deserialise transaction receipt: %s", err)
		}
		return receipt, nil
	default:
		return nil, fmt.Errorf("error returned by Tendermint in BroadcastTxSync "+
			"ABCI code: %v, ABCI log: %v", checkTxResponse.Code, checkTxResponse.Log)
	}
}
