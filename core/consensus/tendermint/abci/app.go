package abci

import (
	"fmt"
	"sync"
	"time"

	"encoding/json"

	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/consensus/tendermint/codes"
	"github.com/gallactic/gallactic/core/execution"
	"github.com/gallactic/gallactic/txs"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"

	"github.com/pkg/errors"
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

const responseInfoName = "Burrow"

type App struct {
	// State
	blockchain    *blockchain.Blockchain
	checker       execution.Executor
	deliverer     execution.Executor
	mempoolLocker sync.Locker
	// We need to cache these from BeginBlock for when we need actually need it in Commit
	block *abciTypes.RequestBeginBlock
	// Utility
	txDecoder txs.Decoder
	// Logging
	logger *logging.Logger
}

var _ abciTypes.Application = &App{}

func NewApp(bc *blockchain.Blockchain, checker, deliverer execution.Executor,
	txDecoder txs.Decoder, logger *logging.Logger) *App {
	return &App{
		blockchain: bc,
		checker:    checker,
		deliverer:  deliverer,
		txDecoder:  txDecoder,
		logger:     logger.WithScope("abci.NewApp").With(structure.ComponentKey, "ABCI_App"),
	}
}

// Provide the Mempool lock. When provided we will attempt to acquire this lock in a goroutine during the Commit. We
// will keep the checker cache locked until we are able to acquire the mempool lock which signals the end of the commit
// and possible recheck on Tendermint's side.
func (app *App) SetMempoolLocker(mempoolLocker sync.Locker) {
	app.mempoolLocker = mempoolLocker
}

func (app *App) Info(info abciTypes.RequestInfo) abciTypes.ResponseInfo {
	return abciTypes.ResponseInfo{
		Data:             responseInfoName,
		Version:          "0.0.0", /// TODO
		LastBlockHeight:  int64(app.blockchain.LastBlockHeight()),
		LastBlockAppHash: app.blockchain.LastAppHash(),
	}
}

func (app *App) SetOption(option abciTypes.RequestSetOption) (respSetOption abciTypes.ResponseSetOption) {
	respSetOption.Log = "SetOption not supported"
	respSetOption.Code = codes.UnsupportedRequestCode
	return
}

func (app *App) Query(reqQuery abciTypes.RequestQuery) (respQuery abciTypes.ResponseQuery) {
	respQuery.Log = "Query not supported"
	respQuery.Code = codes.UnsupportedRequestCode
	return
}

func (app *App) CheckTx(txBytes []byte) abciTypes.ResponseCheckTx {
	txEnv, err := app.txDecoder.DecodeTx(txBytes)
	if err != nil {
		app.logger.TraceMsg("CheckTx decoding error",
			"tag", "CheckTx",
			structure.ErrorKey, err)
		return abciTypes.ResponseCheckTx{
			Code: codes.EncodingErrorCode,
			Log:  fmt.Sprintf("Encoding error: %s", err),
		}
	}
	receipt := txEnv.GenerateReceipt()

	err = app.checker.Execute(txEnv)
	if err != nil {
		app.logger.TraceMsg("CheckTx execution error",
			structure.ErrorKey, err,
			"tag", "CheckTx",
			"tx_hash", receipt.TxHash)
		return abciTypes.ResponseCheckTx{
			Code: codes.EncodingErrorCode,
			Log:  fmt.Sprintf("CheckTx could not execute transaction: %s, error: %v", txEnv, err),
		}
	}

	receiptBytes, err := json.Marshal(receipt)
	if err != nil {
		return abciTypes.ResponseCheckTx{
			Code: codes.TxExecutionErrorCode,
			Log:  fmt.Sprintf("CheckTx could not serialize receipt: %s", err),
		}
	}
	app.logger.TraceMsg("CheckTx success",
		"tag", "CheckTx",
		"tx_hash", receipt.TxHash)
	return abciTypes.ResponseCheckTx{
		Code: codes.TxExecutionSuccessCode,
		Log:  "CheckTx success - receipt in data",
		Data: receiptBytes,
	}
}

func (app *App) InitChain(chain abciTypes.RequestInitChain) (respInitChain abciTypes.ResponseInitChain) {
	// Could verify agreement on initial validator set here
	return
}

func (app *App) BeginBlock(block abciTypes.RequestBeginBlock) (respBeginBlock abciTypes.ResponseBeginBlock) {
	//app.mempoolLocker.Lock()

	app.blockchain.State().ClearChanges()
	app.block = &block

	return
}

func (app *App) DeliverTx(txBytes []byte) abciTypes.ResponseDeliverTx {
	txEnv, err := app.txDecoder.DecodeTx(txBytes)
	if err != nil {
		app.logger.TraceMsg("DeliverTx decoding error",
			"tag", "DeliverTx",
			structure.ErrorKey, err)

		app.mempoolLocker.Unlock()
		return abciTypes.ResponseDeliverTx{
			Code: codes.EncodingErrorCode,
			Log:  fmt.Sprintf("Encoding error: %s", err),
		}
	}

	receipt := txEnv.GenerateReceipt()
	err = app.deliverer.Execute(txEnv)
	if err != nil {
		app.logger.TraceMsg("DeliverTx execution error",
			structure.ErrorKey, err,
			"tag", "DeliverTx",
			"tx_hash", receipt.TxHash)
		app.mempoolLocker.Unlock()
		return abciTypes.ResponseDeliverTx{
			Code: codes.TxExecutionErrorCode,
			Log:  fmt.Sprintf("DeliverTx could not execute transaction: %s, error: %s", txEnv, err),
		}
	}

	app.logger.TraceMsg("DeliverTx success",
		"tag", "DeliverTx",
		"tx_hash", receipt.TxHash)
	receiptBytes, err := json.Marshal(receipt)
	if err != nil {
		app.mempoolLocker.Unlock()
		return abciTypes.ResponseDeliverTx{
			Code: codes.TxExecutionErrorCode,
			Log:  fmt.Sprintf("DeliverTx could not serialize receipt: %s", err),
		}
	}

	return abciTypes.ResponseDeliverTx{
		Code: codes.TxExecutionSuccessCode,
		Log:  "DeliverTx success - receipt in data",
		Data: receiptBytes,
	}
}

func (app *App) EndBlock(reqEndBlock abciTypes.RequestEndBlock) abciTypes.ResponseEndBlock {
	//defer app.mempoolLocker.Unlock()

	return abciTypes.ResponseEndBlock{}
}

func (app *App) Commit() abciTypes.ResponseCommit {
	app.logger.InfoMsg("Committing block",
		"tag", "Commit",
		structure.ScopeKey, "Commit()",
		"height", app.block.Header.Height,
		"hash", app.block.Hash,
		"txs", app.block.Header.NumTxs,
		"block_time", app.block.Header.Time, // [CSK] this sends a fairly non-sensical number; should be human readable
		"last_block_time", app.blockchain.LastBlockTime(),
		"last_block_hash", app.blockchain.LastBlockHash())

	// Commit to our blockchain state which will checkpoint the previous app hash by saving it to the database
	// (we know the previous app hash is safely committed because we are about to commit the next)
	appHash, err := app.blockchain.CommitBlock(time.Unix(int64(app.block.Header.Time), 0), app.block.Hash)
	if err != nil {
		panic(errors.Wrap(err, "could not commit block to blockchain state"))
	}

	// Perform a sanity check our block height
	if app.blockchain.LastBlockHeight() != uint64(app.block.Header.Height) {
		app.logger.InfoMsg("Burrow block height disagrees with Tendermint block height",
			structure.ScopeKey, "Commit()",
			"burrow_height", app.blockchain.LastBlockHeight(),
			"tendermint_height", app.block.Header.Height)

		panic(fmt.Errorf("burrow has recorded a block height of %v, "+
			"but Tendermint reports a block height of %v, and the two should agree",
			app.blockchain.LastBlockHeight(), app.block.Header.Height))
	}
	return abciTypes.ResponseCommit{
		Data: appHash,
	}
}
