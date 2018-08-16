package execution

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/gallactic/gallactic/core/blockchain"
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

type BatchExecutor interface {
	// Provides access to write lock for a BatchExecutor so reads can be prevented for the duration of a commit
	sync.Locker

	// Execute transaction against block cache (i.e. block buffer)
	Executor

	// Reset executor to underlying State
	Reset() error
}

// Executes transactions
type BatchCommitter interface {
	BatchExecutor

	// Commit execution results to underlying State and provide opportunity
	// to mutate state before it is saved
	Commit() (err error)
}

type executor struct {
	sync.RWMutex
	logger      *logging.Logger
	bc          *blockchain.Blockchain
	cache       *state.Cache
	txExecutors map[tx.Type]Executor
}

var _ BatchExecutor = (*executor)(nil)

// Wraps a cache of what is variously known as the 'check cache' and 'mempool'
func NewBatchChecker(bc *blockchain.Blockchain, logger *logging.Logger) BatchExecutor {
	return newExecutor("CheckCache", false, bc, logger.WithScope("NewBatchExecutor"))
}

func NewBatchCommitter(bc *blockchain.Blockchain, logger *logging.Logger) BatchCommitter {
	return newExecutor("CommitCache", true, bc, logger.WithScope("NewBatchCommitter"))
}

func newExecutor(name string, committing bool, bc *blockchain.Blockchain, logger *logging.Logger) *executor {

	exe := &executor{
		bc:     bc,
		cache:  state.NewCache(bc.State(), state.Name(name)),
		logger: logger.With(structure.ComponentKey, "Executor"),
	}

	exe.txExecutors = map[tx.Type]Executor{
		tx.TypeSend: &executors.SendContext{
			Committing: committing,
			Cache:      exe.cache,
			Logger:     exe.logger,
		},
		tx.TypeCall: &executors.CallContext{
			Committing: committing,
			BC:         bc,
			Cache:      exe.cache,
			Logger:     exe.logger,
		},
		tx.TypePermissions: &executors.PermissionContext{
			Committing: committing,
			Cache:      exe.cache,
			Logger:     exe.logger,
		},
		tx.TypeBond: &executors.BondContext{
			Committing: committing,
			BC:         bc,
			Cache:      exe.cache,
			Logger:     exe.logger,
		},
		tx.TypeUnbond: &executors.UnbondContext{
			Committing: committing,
			Cache:      exe.cache,
			Logger:     exe.logger,
		},
		tx.TypeSortition: &executors.SortitionContext{
			Committing: committing,
			BC:         bc,
			Cache:      exe.cache,
			Logger:     exe.logger,
		},
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

	if err = txEnv.Verify(); err != nil {
		return err
	}

	if err = txEnv.Tx.EnsureValid(); err != nil {
		return err
	}

	if txExecutor, ok := exe.txExecutors[txEnv.Tx.Type()]; ok {
		return txExecutor.Execute(txEnv)
	}
	return fmt.Errorf("unknown transaction type: %v", txEnv.Tx.Type())
}

func (exe *executor) Commit() (err error) {
	// The write lock to the executor is controlled by the caller (e.g. abci.App) so we do not acquire it here to avoid
	// deadlock
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in executor.Commit(): %v\n%v", r, debug.Stack())
		}
	}()

	return exe.cache.Flush(exe.bc.ValidatorSet()) /// TODO: better way???
}

func (exe *executor) Reset() error {
	// As with Commit() we do not take the write lock here
	exe.cache.Reset()
	return nil
}
