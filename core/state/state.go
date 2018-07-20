package state

import (
	"fmt"
	"sync"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/hyperledger/burrow/logging"
	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tendermint/libs/db"
)

const (
	defaultCacheCapacity = 1024

	// Version by state hash
	versionPrefix = "v/"

	// Prefix of keys in state tree
	accountPrefix   = "a/"
	storagePrefix   = "s/"
	validatorPrefix = "i/"
	eventPrefix     = "e/"
)

var (
	accountsStart, accountsEnd   []byte = prefixKeyRange(accountPrefix)
	validatorStart, validatorEnd []byte = prefixKeyRange(validatorPrefix)
	storageStart, storageEnd     []byte = prefixKeyRange(storagePrefix)
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

func accountKey(addr crypto.Address) []byte {
	return prefixedKey(accountPrefix, addr.RawBytes())
}

func validatorKey(addr crypto.Address) []byte {
	return prefixedKey(validatorPrefix, addr.RawBytes())
}

type State struct {
	sync.Mutex
	db     dbm.DB
	vTree  *iavl.VersionedTree
	tree   *iavl.Tree
	logger *logging.Logger
}

// NewState creates a new instance of State object
func NewState(db dbm.DB, logger *logging.Logger) *State {
	vTree := iavl.NewVersionedTree(db, defaultCacheCapacity)
	st := &State{
		db:     db,
		vTree:  vTree,
		tree:   vTree.Tree(),
		logger: logger,
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

	treeVer, err := st.vTree.LoadVersion(ver)
	if err != nil {
		return nil, e.ErrorE(e.ErrLoadingState, err)
	}
	if treeVer != ver {
		return nil, e.Errorf(e.ErrLoadingState, "Trying to load state version %v for state hash %X but loaded version %v", ver, hash, treeVer)
	}

	st.tree = st.vTree.Tree()

	return st, nil
}

func (st *State) SaveState() ([]byte, error) {
	st.Lock()
	defer st.Unlock()

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
	st.tree = st.vTree.Tree()

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

	if addr.IsAccountAddress() {
		return st.GetAccount(addr)
	} else if addr.IsValidatorAddress() {
		return st.GetValidator(addr)
	}

	return nil
}

// -------
// ACCOUNT

func (st *State) GlobalAccount() *account.Account {
	return st.GetAccount(crypto.GlobalAddress)
}

func (st *State) GetAccount(addr crypto.Address) *account.Account {
	st.Lock()
	defer st.Unlock()

	_, bs := st.tree.Get(accountKey(addr))
	if bs == nil {
		return nil
	}
	acc, err := account.AccountFromBytes(bs)
	if err != nil {
		panic("Unable to decode encoded Account")
	}

	return acc
}

func (st *State) UpdateAccount(acc *account.Account) error {
	st.Lock()
	defer st.Unlock()

	bs, err := acc.Encode()
	if err != nil {
		return err
	}

	st.tree.Set(accountKey(acc.Address()), bs)
	return nil
}

func (st *State) Count() int {
	count := 0
	st.IterateAccounts(func(validator *account.Account) (stop bool) {
		count++
		return false
	})
	return count
}

func (st *State) IterateAccounts(consumer func(*account.Account) (stop bool)) (stopped bool, err error) {
	stopped = st.tree.IterateRange(accountsStart, accountsEnd, true, func(key, bs []byte) bool {
		acc, err := account.AccountFromBytes(bs)
		if err != nil {
			return true
		}
		return consumer(acc)
	})
	return
}

// ---------
// VALIDATOR

func (st *State) GetValidator(addr crypto.Address) *validator.Validator {
	st.Lock()
	defer st.Unlock()

	_, bytes := st.tree.Get(validatorKey(addr))
	if bytes == nil {
		return nil
	}
	val, err := validator.ValidatorFromBytes(bytes)
	if err != nil {
		panic("Unable to decode encoded validator")
	}

	return val
}

func (st *State) UpdateValidator(val *validator.Validator) error {
	st.Lock()
	defer st.Unlock()

	bs, err := val.Encode()
	if err != nil {
		return err
	}

	st.tree.Set(accountKey(val.Address()), bs)

	return nil
}

func (st *State) ValidatorCount() int {
	count := 0
	st.IterateValidators(func(val *validator.Validator) (stop bool) {
		count++
		return false
	})
	return count
}

func (st *State) IterateValidators(consumer func(*validator.Validator) (stop bool)) (stopped bool, err error) {
	return st.tree.IterateRange(validatorStart, validatorEnd, true, func(key []byte, bs []byte) (stop bool) {
		validator, err := validator.ValidatorFromBytes(bs)
		if err != nil {
			return true
		}
		return consumer(validator)
	}), nil
}

func (s *State) IterateStorage(address crypto.Address,
	consumer func(key, value binary.Word256) (stop bool)) (stopped bool, err error) {
	stopped = s.tree.IterateRange(storageStart, storageEnd, true, func(key []byte, value []byte) (stop bool) {
		// Note: no left padding should occur unless there is a bug and non-words have been writte to this storage tree
		if len(key) != binary.Word256Length {
			err = fmt.Errorf("key '%X' stored for account %s is not a %v-byte word",
				key, address, binary.Word256Length)
			return true
		}
		if len(value) != binary.Word256Length {
			err = fmt.Errorf("value '%X' stored for account %s is not a %v-byte word",
				key, address, binary.Word256Length)
			return true
		}
		return consumer(binary.LeftPadWord256(key), binary.LeftPadWord256(value))
	})
	return
}

func (s *State) GetStorage(address crypto.Address, key binary.Word256) (binary.Word256, error) {
	_, value := s.tree.Get(prefixedKey(storagePrefix, address.RawBytes(), key.Bytes()))
	return binary.LeftPadWord256(value), nil
}

func (s *State) SetStorage(address crypto.Address, key, value binary.Word256) error {
	if value == binary.Zero256 {
		s.tree.Remove(key.Bytes())
	} else {
		s.tree.Set(prefixedKey(storagePrefix, address.RawBytes(), key.Bytes()), value.Bytes())
	}
	return nil
}
