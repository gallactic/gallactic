package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gallactic/gallactic/core/proposal"
	"github.com/gallactic/gallactic/core/sortition"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/hyperledger/burrow/logging"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var stateKey = []byte("BlockchainState")

type Blockchain struct {
	chainID      string
	genesisHash  []byte
	db           dbm.DB
	state        *state.State
	data         *blockchainData
	validatorSet *validator.ValidatorSet
	sortition    *sortition.Sortition
	logger       *logging.Logger
}

type blockchainData struct {
	Genesis         *proposal.Genesis `json:"genesis"`
	LastBlockTime   time.Time         `json:"lastBlockTime"`
	LastBlockHeight uint64            `json:"lastBlockHeight"`
	LastBlockHash   []byte            `json:"lastBlockHash"`
	LastAppHash     []byte            `json:"lastAppHash"`
	LastValidators  []crypto.Address  `json:"lastValidators"`
	MaximumPower    int               `json:"maximumPower"`
}

func LoadOrNewBlockchain(db dbm.DB, gen *proposal.Genesis, myVal crypto.Signer, logger *logging.Logger) (*Blockchain, error) {
	logger = logger.WithScope("LoadOrNewBlockchain")
	logger.InfoMsg("Trying to load blockchain state from database",
		"database_key", stateKey)
	bc, err := loadBlockchain(db, logger)
	if err != nil {
		return nil, fmt.Errorf("error loading blockchain state from database: %v", err)
	}

	if bc != nil {
		dbHash := bc.genesisHash
		argHash := gen.Hash()
		if !bytes.Equal(dbHash, argHash) {
			return nil, fmt.Errorf("Genesis passed to LoadOrNewBlockchain has hash: 0x%X, which does not "+
				"match the one found in database: 0x%X", argHash, dbHash)
		}
	} else {

		logger.InfoMsg("No existing blockchain state found in database, making new blockchain")

		bc, err = newBlockchain(db, gen, logger)
		if err != nil {
			return nil, fmt.Errorf("error creating blockchain from genesis doc: %v", err)
		}
	}

	/// set logger
	bc.logger = logger

	if err := bc.loadValidatorSet(); err != nil {
		return nil, err
	}

	if err := bc.createSortition(myVal); err != nil {
		return nil, err
	}

	return bc, nil
}

// Pointer to blockchain state initialized from genesis
func newBlockchain(db dbm.DB, gen *proposal.Genesis, logger *logging.Logger) (*Blockchain, error) {
	if len(gen.Validators()) == 0 {
		return nil, fmt.Errorf("The genesis file has no validators")
	}

	if gen.GenesisTime().IsZero() {
		return nil, fmt.Errorf("Genesis time didn't set inside genesis doc")
	}

	st := state.NewState(db, logger)

	// Make accounts state tree
	for _, acc := range gen.Accounts() {
		if err := st.UpdateAccount(acc); err != nil {
			return nil, err
		}
	}

	var vals []crypto.Address
	for _, val := range gen.Validators() {
		if err := st.UpdateValidator(val); err != nil {
			return nil, err
		}
		vals = append(vals, val.Address())
	}

	// We need to save at least once so that readTree points at a non-working-state tree
	_, err := st.SaveState()
	if err != nil {
		return nil, err
	}

	bc := &Blockchain{
		chainID:     gen.ChainID(),
		genesisHash: gen.Hash(),
		db:          db,
		state:       st,
		data: &blockchainData{
			Genesis:        gen,
			LastBlockTime:  gen.GenesisTime(),
			MaximumPower:   gen.MaximumPower(),
			LastAppHash:    gen.Hash(),
			LastValidators: vals,
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

func (bc *Blockchain) ChainID() string            { return bc.chainID }
func (bc *Blockchain) GenesisHash() []byte        { return bc.genesisHash }
func (bc *Blockchain) Genesis() *proposal.Genesis { return bc.data.Genesis }
func (bc *Blockchain) LastBlockHeight() uint64    { return bc.data.LastBlockHeight }
func (bc *Blockchain) LastBlockTime() time.Time   { return bc.data.LastBlockTime }
func (bc *Blockchain) LastBlockHash() []byte      { return bc.data.LastBlockHash }
func (bc *Blockchain) LastAppHash() []byte        { return bc.data.LastAppHash }
func (bc *Blockchain) MaximumPower() int          { return bc.data.MaximumPower }

func (bc *Blockchain) ValidatorSet() *validator.ValidatorSet {
	return bc.validatorSet
}

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

	vals := make([]crypto.Address, 0)
	for addr, _ := range bc.validatorSet.Validators() {
		vals = append(vals, addr)
	}

	bc.data.LastBlockHeight++
	bc.data.LastBlockTime = blockTime
	bc.data.LastBlockHash = blockHash
	bc.data.LastAppHash = appHash
	bc.data.LastValidators = vals
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

func (bc *Blockchain) loadValidatorSet() error {
	valMap := make(map[crypto.Address]*validator.Validator)
	for _, addr := range bc.data.LastValidators {
		val, err := bc.state.GetValidator(addr)
		if err != nil {
			return err
		}

		valMap[addr] = val
	}

	bc.validatorSet = validator.NewValidatorSet(valMap, bc.data.MaximumPower, bc.logger)
	return nil
}

func (bc *Blockchain) createSortition(myVal crypto.Signer) error {
	bc.sortition = sortition.NewSortition(myVal, bc.chainID, bc.logger)
	return nil
}

func (bc *Blockchain) EvaluateSortition(blockHeight uint64, blockHash []byte) {

	// check if this validator is in set or not
	if bc.validatorSet.Contains(bc.sortition.Address()) {
		return
	}

	// this validator is not in the set
	go bc.sortition.Evaluate(blockHeight, blockHash)
}

func (bc *Blockchain) VerifySortition(blockHash []byte, publicKey crypto.PublicKey, info uint64, proof []byte) bool {
	return bc.sortition.Verify(blockHash, publicKey, info, proof)
}
