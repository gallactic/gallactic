package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/genesis"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/hyperledger/burrow/logging"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var stateKey = []byte("BlockchainState")

type Blockchain struct {
	chainID     string
	genesisHash []byte
	db          dbm.DB
	state       *state.State
	data        *blockchainData
}

type blockchainData struct {
	Genesis         *genesis.Genesis `json:"genesisDoc"`
	LastAppHash     []byte           `json:"lastAppHash"`
	LastBlockHash   []byte           `json:"lastBlockHash"`
	LastBlockHeight uint64           `json:"lastBlockHeight"`
	LastBlockTime   time.Time        `json:"lastBlockTime"`
}

func LoadOrNewBlockchain(db dbm.DB, gen *genesis.Genesis, logger *logging.Logger) (*Blockchain, error) {

	logger = logger.WithScope("LoadOrNewBlockchain")
	logger.InfoMsg("Trying to load blockchain state from database",
		"database_key", stateKey)
	bc, err := loadBlockchain(db, logger)
	if err != nil {
		return nil, e.Errorf(e.ErrGeneric, "error loading blockchain state from database: %v", err)
	}
	if bc != nil {
		dbHash := bc.genesisHash
		argHash := gen.Hash()
		if !bytes.Equal(dbHash, argHash) {
			return nil, fmt.Errorf("Genesis passed to LoadOrNewBlockchain has hash: 0x%X, which does not "+
				"match the one found in database: 0x%X", argHash, dbHash)
		}
		return bc, nil
	}

	logger.InfoMsg("No existing blockchain state found in database, making new blockchain")
	return newBlockchain(db, gen, logger)
}

// Pointer to blockchain state initialized from genesis
func newBlockchain(db dbm.DB, gen *genesis.Genesis, logger *logging.Logger) (*Blockchain, error) {
	if len(gen.Validators()) == 0 {
		return nil, fmt.Errorf("The genesis file has no validators")
	}

	if gen.GenesisTime().IsZero() {
		return nil, fmt.Errorf("Genesis time didn't set inside genesis doc")
	}

	st := state.NewState(db, logger)

	// Make accounts state tree
	for _, acc := range gen.Accounts() {
		err := st.UpdateAccount(acc)
		if err != nil {
			return nil, err
		}
	}

	gAcc, _ := account.NewAccount(crypto.GlobalAddress)
	gAcc.SetPermissions(gen.GlobalPermissions())

	err := st.UpdateAccount(gAcc)
	if err != nil {
		return nil, err
	}

	// We need to save at least once so that readTree points at a non-working-state tree
	_, err = st.SaveState()
	if err != nil {
		return nil, err
	}

	bc := &Blockchain{
		chainID:     gen.ChainID(),
		genesisHash: gen.Hash(),
		db:          db,
		state:       st,
		data: &blockchainData{
			Genesis:       gen,
			LastBlockTime: gen.GenesisTime(),
			LastAppHash:   gen.Hash(),
		},
	}
	return bc, nil
}

func loadBlockchain(db dbm.DB, logger *logging.Logger) (*Blockchain, error) {
	buf := db.Get(stateKey)
	if len(buf) == 0 {
		return nil, nil
	}
	data := new(blockchainData)

	err := json.Unmarshal(buf, data)
	if err != nil {
		return nil, err
	}

	st, err := state.LoadState(db, data.LastAppHash, logger)
	if err != nil {
		return nil, fmt.Errorf("could not load persisted execution state at hash 0x%X: %v", data.LastAppHash, err)
	}

	bc := &Blockchain{
		chainID:     data.Genesis.ChainID(),
		genesisHash: data.Genesis.Hash(),
		db:          db,
		state:       st,
		data:        data,
	}

	return bc, nil
}

func (bc *Blockchain) State() *state.State {
	return bc.state
}

func (bc *Blockchain) ChainID() string           { return bc.chainID }
func (bc *Blockchain) GenesisHash() []byte       { return bc.genesisHash }
func (bc *Blockchain) Genesis() *genesis.Genesis { return bc.data.Genesis }
func (bc *Blockchain) LastBlockHeight() uint64   { return bc.data.LastBlockHeight }
func (bc *Blockchain) LastBlockTime() time.Time  { return bc.data.LastBlockTime }
func (bc *Blockchain) LastBlockHash() []byte     { return bc.data.LastBlockHash }
func (bc *Blockchain) LastAppHash() []byte       { return bc.data.LastAppHash }

func (bc *Blockchain) CommitBlock(blockTime time.Time, blockHash []byte) ([]byte, error) {
	// Checkpoint on the _previous_ block. If we die, this is where we will resume since we know it must have been
	// committed since we are committing the next block. If we fall over we can resume a safe committed state and
	// Tendermint will catch us up
	err := bc.save()
	if err != nil {
		return nil, err
	}

	appHash, err := bc.state.SaveState()
	if err != nil {
		return nil, err
	}

	bc.data.LastBlockHeight++
	bc.data.LastBlockTime = blockTime
	bc.data.LastBlockHash = blockHash
	bc.data.LastAppHash = appHash
	return appHash, nil
}

func (bc *Blockchain) save() error {
	if bc.db != nil {
		bytes, err := json.Marshal(&bc.data)
		if err != nil {
			return err
		}
		bc.db.SetSync(stateKey, bytes)
	}
	return nil
}
