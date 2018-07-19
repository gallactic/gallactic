package core

import (
	"context"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gallactic/gallactic/common/process"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/config"
	"github.com/gallactic/gallactic/core/consensus/tendermint"
	"github.com/gallactic/gallactic/core/consensus/tendermint/query"
	tmv "github.com/gallactic/gallactic/core/consensus/tendermint/validator" // TODO:::
	"github.com/gallactic/gallactic/core/execution"
	"github.com/gallactic/gallactic/core/genesis"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/rpc"
	"github.com/gallactic/gallactic/txs"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/hyperledger/burrow/logging/structure"

	kitlog "github.com/go-kit/kit/log"
	dbm "github.com/tendermint/tendermint/libs/db"
)

const (
	cooldownMilliseconds              = 1000
	serverShutdownTimeoutMilliseconds = 1000
	loggingCallerDepth                = 5
)

// Kernel is the root structure of Gallactic
type Kernel struct {
	// Expose these public-facing interfaces to allow programmatic extension of the Kernel by other projects
	st             *state.State
	bc             *blockchain.Blockchain
	logger         *logging.Logger
	launchers      []process.Launcher
	processes      map[string]process.Process
	shutdownNotify chan struct{}
	shutdownOnce   sync.Once
}

func NewKernel(ctx context.Context, gen *genesis.Genesis, conf *config.Config, passphrase string) (*Kernel, error) {

	if gen == nil {
		return nil, fmt.Errorf("no Genesis defined, cannot make Kernel")
	}
	if conf == nil {
		return nil, fmt.Errorf("no Config defined, cannot make Kernel")
	}
	if err := conf.Validator.Address.EnsureValid(); err != nil {
		return nil, fmt.Errorf("Not a valid address provided for validator, cannot make Kernel: %v", err)
	}
	logger, err := lifecycle.NewLoggerFromLoggingConfig(&conf.Logging)
	if err != nil {
		return nil, fmt.Errorf("could not generate logger from logging config: %v", err)
	}

	logger = logger.WithScope("NewKernel()").With(structure.TimeKey, kitlog.DefaultTimestampUTC)
	tmLogger := logger.With(structure.CallerKey, kitlog.Caller(loggingCallerDepth+1))
	logger = logger.WithInfo(structure.CallerKey, kitlog.Caller(loggingCallerDepth))

	tmGenesisDoc := tendermint.DeriveGenesisDoc(gen)
	tmConfig := tendermint.DeriveConfig(conf)

	stateDB := dbm.NewDB("gallactic_state", dbm.GoLevelDBBackend, tmConfig.DBDir())

	bc, err := blockchain.LoadOrNewBlockchain(stateDB, gen, logger)
	if err != nil {
		/// TODO: return proper error with code
		return nil, fmt.Errorf("error creating or loading blockchain state: %v", err)
	}

	signer, err := validator.NewPrivateValidator("", passphrase)
	if err != nil {
		return nil, fmt.Errorf("could not create the signer from keystore: %v", err)
	}
	privValidator := tmv.NewPrivValidatorMemory(signer)

	txCodec := txs.NewAminoCodec()
	checker := execution.NewBatchChecker(bc, logger)
	committer := execution.NewBatchCommitter(bc, logger)

	tmNode, err := tendermint.NewNode(tmConfig, privValidator, tmGenesisDoc, bc, checker, committer, txCodec,
		tmLogger)

	if err != nil {
		return nil, err
	}

	transactor := execution.NewTransactor(tmNode.MempoolReactor().BroadcastTx, txCodec, logger)

	service := rpc.NewService(ctx, bc, transactor, query.NewNodeView(tmNode, txCodec), logger)

	launchers := []process.Launcher{
		{
			Name:    "Database",
			Enabled: true,
			Launch: func() (process.Process, error) {
				// Just close database
				return process.ShutdownFunc(func(ctx context.Context) error {
					stateDB.Close()
					return nil
				}), nil
			},
		},
		{
			Name:    "Tendermint",
			Enabled: true,
			Launch: func() (process.Process, error) {
				err := tmNode.Start()
				if err != nil {
					return nil, fmt.Errorf("error starting Tendermint node: %v", err)
				}
				if err != nil {
					return nil, fmt.Errorf("could not subscribe to Tendermint events: %v", err)
				}
				return process.ShutdownFunc(func(ctx context.Context) error {
					err := tmNode.Stop()
					// Close tendermint database connections using our wrapper
					defer tmNode.Close()
					if err != nil {
						return err
					}
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-tmNode.Quit():
						logger.InfoMsg("Tendermint Node has quit, closing DB connections...")
						return nil
					}
					return err
				}), nil
			},
		},
		{
			Name:    "RPC",
			Enabled: conf.RPC.Enabled,
			Launch: func() (process.Process, error) {
				codec := rpc.NewTCodec()
				jsonServer := rpc.NewJSONServer(rpc.NewJSONService(codec, service, logger))
				serveProcess, err := rpc.NewServeProcess(conf.RPC.Server, logger, jsonServer)
				if err != nil {
					return nil, err
				}
				err = serveProcess.Start()
				if err != nil {
					return nil, err
				}
				return serveProcess, nil
			},
		},
	}

	return &Kernel{
		launchers:      launchers,
		bc:             bc,
		logger:         logger,
		processes:      make(map[string]process.Process),
		shutdownNotify: make(chan struct{}),
	}, nil
}

// Boot the kernel starting Tendermint and RPC layers
func (kern *Kernel) Boot() error {
	for _, launcher := range kern.launchers {
		if launcher.Enabled {
			srvr, err := launcher.Launch()
			if err != nil {
				return fmt.Errorf("error launching %s server: %v", launcher.Name, err)
			}

			kern.processes[launcher.Name] = srvr
		}
	}
	go kern.supervise()
	return nil
}

// Wait for a graceful shutdown
func (kern *Kernel) WaitForShutdown() {
	// Supports multiple goroutines waiting for shutdown since channel is closed
	<-kern.shutdownNotify
}

// Supervise kernel once booted
func (kern *Kernel) supervise() {
	// TODO: Consider capturing kernel panics from boot and sending them here via a channel where we could
	// perform disaster restarts of the kernel; rejoining the network as if we were a new node.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sig := <-signals
	kern.logger.InfoMsg(fmt.Sprintf("Caught %v signal so shutting down", sig),
		"signal", sig.String())
	kern.Shutdown(context.Background())
}

// Stop the kernel allowing for a graceful shutdown of components in order
func (kern *Kernel) Shutdown(ctx context.Context) (err error) {
	kern.shutdownOnce.Do(func() {
		logger := kern.logger.WithScope("Shutdown")
		logger.InfoMsg("Attempting graceful shutdown...")
		logger.InfoMsg("Shutting down servers")
		ctx, cancel := context.WithTimeout(ctx, serverShutdownTimeoutMilliseconds*time.Millisecond)
		defer cancel()
		// Shutdown servers in reverse order to boot
		for i := len(kern.launchers) - 1; i >= 0; i-- {
			name := kern.launchers[i].Name
			srvr, ok := kern.processes[name]
			if ok {
				logger.InfoMsg("Shutting down server", "server_name", name)
				sErr := srvr.Shutdown(ctx)
				if sErr != nil {
					logger.InfoMsg("Failed to shutdown server",
						"server_name", name,
						structure.ErrorKey, sErr)
					if err == nil {
						err = sErr
					}
				}
			}
		}
		logger.InfoMsg("Shutdown complete")
		structure.Sync(kern.logger.Info)
		structure.Sync(kern.logger.Trace)
		// We don't want to wait for them, but yielding for a cooldown Let other goroutines flush
		// potentially interesting final output (e.g. log messages)
		time.Sleep(time.Millisecond * cooldownMilliseconds)
		close(kern.shutdownNotify)
	})
	return
}
