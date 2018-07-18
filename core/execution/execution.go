package execution

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/gallactic/gallactic/core/execution/executors"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/txs"
	"github.com/gallactic/gallactic/txs/tx"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
)

type Executor interface {
	Execute(txEnv *txs.Envelope) error
}

type executor struct {
	sync.RWMutex
	logger      *logging.Logger
	txExecutors map[tx.Type]Executor
}

// Wraps a cache of what is variously known as the 'check cache' and 'mempool'
func NewChecker(st *state.State, logger *logging.Logger) Executor {
	return newExecutor(false, st, logger.WithScope("NewExecutor"))
}

func NewDeliverer(st *state.State, logger *logging.Logger) Executor {
	return newExecutor(true, st, logger.WithScope("Deliverer"))
}

func newExecutor(deliver bool, st *state.State, logger *logging.Logger) *executor {

	exe := &executor{
		logger: logger.With(structure.ComponentKey, "Executor"),
	}

	exe.txExecutors = map[tx.Type]Executor{
		tx.TypeSend: &executors.SendContext{
			Deliver: deliver,
			State:   st,
			Logger:  exe.logger,
		}, /*
			tx.TypeCall: &executors.CallContext{
				Blockchain:     blockchain,
				StateWriter:    exe.stateCache,
				EventPublisher: publisher,
				RunCall:        runCall,
				VMOptions:      exe.vmOptions,
				Logger:         exe.logger,
			},
			tx.TypePermissions: &executors.PermissionsContext{
				Blockchain:     blockchain,
				StateWriter:    exe.stateCache,
				EventPublisher: publisher,
				Logger:         exe.logger,
			},
		*/
	}
	return exe
}

// If the tx is invalid, an error will be returned.
// Unlike ExecBlock(), state will not be altered.
func (exe *executor) Execute(txEnv *txs.Envelope) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in executor.Execute(%s): %v\n%s", txEnv.String(), r,
				debug.Stack())
		}
	}()

	logger := exe.logger.WithScope("executor.Execute(tx txs.Tx)").With("tx_hash", txEnv.Hash())
	logger.TraceMsg("Executing transaction", "tx", txEnv.String())

	// Verify transaction signature against inputs
	err = txEnv.Verify()
	if err != nil {
		return err
	}

	if txExecutor, ok := exe.txExecutors[txEnv.Tx.Type()]; ok {
		return txExecutor.Execute(txEnv)
	}
	return fmt.Errorf("unknown transaction type: %v", txEnv.Tx.Type())
}
