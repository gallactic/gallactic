package state

import (
	"fmt"
	"sync"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/hyperledger/burrow/logging"
	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tendermint/libs/db"
)

const (
	defaultCacheCapacity = 1024
	// Age of state versions in blocks before we remove them. This has us keeping a little over an hour's worth of blocks
	// in principle we could manage with 2. Ideally we would lift this limit altogether but IAVL leaks memory on access
	// to previous tree versions since it lazy loads values (nice) but gives no ability to unload them (see SaveBranch)
	defaultVersionExpiry = 2048

	// Version by state hash
	versionPrefix = "v/"

	// Prefix of keys in state tree
	accountPrefix   = "a/"
	storagePrefix   = "s/"
	validatorPrefix = "i/"
	/// AHMAD:::
	///eventPrefix        = "e/"
)

var (
	accountsStart, accountsEnd   []byte = prefixKeyRange(accountPrefix)
	storageStart, storageEnd     []byte = prefixKeyRange(storagePrefix)
	validatorStart, validatorEnd []byte = prefixKeyRange(validatorPrefix)
)

func prefixedKey(prefix string, suffices ...[]byte) []byte {
	key := []byte(prefix)
	for _, suffix := range suffices {
		key = append(key, suffix...)
	}
	return key
}

// Returns the start key equal to the bs of prefix and the end key which lexicographically above any key beginning
// with prefix
func prefixKeyRange(prefix string) (start, end []byte) {
	start = []byte(prefix)
	for i := len(start) - 1; i >= 0; i-- {
		c := start[i]
		if c < 0xff {
			end = make([]byte, i+1)
			copy(end, start)
			end[i]++
			return
		}
	}
	return
}

func accountKey(address crypto.Address) []byte {
	return prefixedKey(accountPrefix, address.RawBytes())
}

func validatorKey(address crypto.Address) []byte {
	return prefixedKey(validatorPrefix, address.RawBytes())
}

type State struct {
	sync.Mutex
	AccountPool   *AccountPool
	ValidatorPool *ValidatorPool
	db            dbm.DB
	vTree         *iavl.VersionedTree
	logger        *logging.Logger
}

// NewState creates a new instance of State object
func NewState(db dbm.DB, logger *logging.Logger) *State {
	vTree := iavl.NewVersionedTree(db, defaultCacheCapacity)
	st := &State{
		db:            db,
		vTree:         vTree,
		logger:        logger,
		AccountPool:   NewAccountPool(vTree.Tree()),
		ValidatorPool: NewValidatorPool(vTree.Tree()),
	}

	return st
}

// LoadState tries to load the execution state from DB, returns nil with no error if no state found
func LoadState(db dbm.DB, hash []byte, logger *logging.Logger) (*State, error) {
	st := NewState(db, logger)

	// Get the version associated with this state hash
	ver, err := st.getVersion(hash)
	if err != nil {
		return nil, e.ErrorE(e.ErrLoadingState, err)
	}
	if ver <= 0 {
		return nil, e.Errorf(e.ErrLoadingState, "Trying to load state from non-positive version. version %v, hash: %X", ver, hash)
	}

	// Load previous version for readTree
	treeVer, err := st.vTree.LoadVersion(ver)
	if err != nil {
		return nil, e.ErrorE(e.ErrLoadingState, err)
	}
	if treeVer != ver {
		return nil, e.Errorf(e.ErrLoadingState, "Trying to load state version %v for state hash %X but loaded version %v", ver, hash, treeVer)
	}
	if err := st.AccountPool.UpdateTree(st.vTree.Tree()); err != nil {
		return nil, e.ErrorE(e.ErrLoadingState, err)
	}

	if err := st.ValidatorPool.UpdateTree(st.vTree.Tree()); err != nil {
		return nil, e.ErrorE(e.ErrLoadingState, err)
	}

	return st, nil
}

func (st *State) SaveState() ([]byte, error) {
	st.Lock()
	defer st.Unlock()

	if err := st.AccountPool.flush(); err != nil {
		return nil, e.Errorf(e.ErrSavingState, err.Error())
	}

	if err := st.ValidatorPool.flush(); err != nil {
		return nil, e.Errorf(e.ErrSavingState, err.Error())
	}

	// save state at a new version may still be orphaned before we save the version against the hash
	hash, version, err := st.vTree.SaveVersion()
	if err != nil {
		return nil, e.Errorf(e.ErrSavingState, err.Error())
	}
	if hash == nil {
		return nil, e.Errorf(e.ErrSavingState, "The root is not set")
	}

	// Provide a reference to load this version in the future from the state hash
	st.setVersion(hash, version)

	// Update tree
	tree := st.vTree.Tree()
	if err := st.AccountPool.UpdateTree(tree); err != nil {
		return nil, e.ErrorE(e.ErrSavingState, err)
	}
	if err := st.ValidatorPool.UpdateTree(tree); err != nil {
		return nil, e.ErrorE(e.ErrSavingState, err)
	}

	return hash, nil
}

// GetVersion gets a previously saved tree version stored by state hash
func (st *State) getVersion(hash []byte) (int64, error) {
	bs := st.db.Get(prefixedKey(versionPrefix, hash))
	if bs == nil {
		return -1, fmt.Errorf("Could not retrieve version corresponding to state hash '%X' in database", hash)
	}
	return binary.GetInt64BE(bs), nil
}

// SetVersion sets the tree version associated with a particular hash
func (st *State) setVersion(hash []byte, version int64) {
	bs := make([]byte, 8)
	binary.PutInt64BE(bs, version)
	st.db.SetSync(prefixedKey(versionPrefix, hash), bs)
}

func (st *State) GetObj(addr crypto.Address) StateObj {
	st.Lock()
	defer st.Unlock()

	if addr.IsAccountAddress() {
		return st.AccountPool.GetAccount(addr)
	} else if addr.IsValidatorAddress() {
		return st.ValidatorPool.GetValidator(addr)
	}

	return nil
}

func (st *State) GlobalAccount() *account.Account {
	st.Lock()
	defer st.Unlock()

	return st.AccountPool.GetAccount(crypto.GlobalAddress)
}

func (st *State) ClearChanges() {
	st.AccountPool.clear()
	st.ValidatorPool.clear()
}
