package tendermint

import (
	"os"
	"path"

	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/config"
	"github.com/gallactic/gallactic/core/consensus/tendermint/abci"
	"github.com/gallactic/gallactic/core/execution"
	"github.com/gallactic/gallactic/core/proposal"
	"github.com/gallactic/gallactic/txs"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	tmConfig "github.com/tendermint/tendermint/config"
	tmEd25519 "github.com/tendermint/tendermint/crypto/ed25519"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	tmTypes "github.com/tendermint/tendermint/types"
)

// Node serves as a wrapper around the Tendermint node's closeable resources (database connections)
type Node struct {
	*node.Node
	closers []interface {
		Close()
	}
}

func DBProvider(ID string, backendType dbm.DBBackendType, dbDir string) dbm.DB {
	return dbm.NewDB(ID, backendType, dbDir)
}

func (n *Node) DBProvider(ctx *node.DBContext) (dbm.DB, error) {
	db := DBProvider(ctx.ID, dbm.DBBackendType(ctx.Config.DBBackend), ctx.Config.DBDir())
	n.closers = append(n.closers, db)
	return db, nil
}

// Close closes DB connections. Since Tendermint doesn't close its DB connections
func (n *Node) Close() {
	for _, closer := range n.closers {
		closer.Close()
	}
}

func NewNode(conf *tmConfig.Config, privValidator tmTypes.PrivValidator, gen *tmTypes.GenesisDoc,
	bc *blockchain.Blockchain, checker execution.BatchExecutor, committer execution.BatchCommitter,
	txDecoder txs.Decoder, logger *logging.Logger) (*Node, error) {

	var err error

	err = os.MkdirAll(path.Dir(conf.NodeKeyFile()), 0777)
	if err != nil {
		return nil, err
	}

	// metricsProvider function
	metricsProvider := node.DefaultMetricsProvider(&tmConfig.InstrumentationConfig{
		Prometheus:           false,
		PrometheusListenAddr: "",
	})

	tmLogger := NewLogger(logger.WithPrefix(structure.ComponentKey, "Tendermint").With(structure.ScopeKey, "tendermint.NewNode"))
	n := &Node{}
	app := abci.NewApp(bc, checker, committer, txDecoder, logger)
	client := proxy.NewLocalClientCreator(app)
	conf.NodeKeyFile()
	n.Node, err = node.NewNode(conf, privValidator, client,
		func() (*tmTypes.GenesisDoc, error) {
			return gen, nil
		},
		n.DBProvider,
		metricsProvider,
		tmLogger)

	if err != nil {
		return nil, err
	}
	app.SetMempoolLocker(n.MempoolReactor().Mempool)
	return n, nil
}

func DeriveGenesisDoc(gen *proposal.Genesis) *tmTypes.GenesisDoc {
	validators := make([]tmTypes.GenesisValidator, len(gen.Validators()))
	for i, validator := range gen.Validators() {
		tm := tmEd25519.PubKeyEd25519{}
		copy(tm[:], validator.PublicKey().RawBytes())
		validators[i] = tmTypes.GenesisValidator{
			PubKey: tm,
			Power:  int64(validator.Power()),
		}
	}
	return &tmTypes.GenesisDoc{
		ChainID:         gen.ChainID(),
		GenesisTime:     gen.GenesisTime(),
		Validators:      validators,
		AppHash:         gen.Hash(),
		ConsensusParams: tmTypes.DefaultConsensusParams(),
	}
}

func DeriveConfig(conf *config.Config) *tmConfig.Config {

	tmConf := tmConfig.DefaultConfig()

	tmConf.SetRoot(conf.Tendermint.TendermintRoot)
	tmConf.P2P.Seeds = conf.Tendermint.Seeds
	tmConf.P2P.PersistentPeers = conf.Tendermint.PersistentPeers
	tmConf.P2P.ListenAddress = conf.Tendermint.ListenAddress
	tmConf.Moniker = conf.Tendermint.Moniker
	tmConf.RPC.ListenAddress = "tcp://0.0.0.0:0" /// TODO: change it to localhost:0

	return tmConf
}
