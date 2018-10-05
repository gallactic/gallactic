package state

import (
	"fmt"
	"sync"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/proposal"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
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
	tree   *iavl.MutableTree
	logger *logging.Logger
}

// NewState creates a new instance of State object
func NewState(db dbm.DB, logger *logging.Logger) *State {
	tree := iavl.NewMutableTree(db, defaultCacheCapacity)
	st := &State{
		db:     db,
		tree:   tree,
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
		return nil, err
	}
	if ver <= 0 {
		return nil, fmt.Errorf("Trying to load state from non-positive version. version %v, hash: %X", ver, hash)
	}

	treeVer, err := st.tree.LoadVersion(ver)
	if err != nil {
		return nil, err
	}
	if treeVer != ver {
		return nil, fmt.Errorf("Trying to load state version %v for state hash %X but loaded version %v", ver, hash, treeVer)
	}

	return st, nil
}

// UpdateGenesisState updates state at genesis time
func (st *State) UpdateGenesisState(gen *proposal.Genesis) error {
	// Make accounts state tree
	for _, acc := range gen.Accounts() {
		if err := st.updateAccount(acc); err != nil {
			return err
		}
	}

	for _, val := range gen.Validators() {
		if err := st.updateValidator(val); err != nil {
			return err
		}
	}

	return nil
}

func (st *State) SaveState() ([]byte, error) {
	st.Lock()
	defer st.Unlock()

	hash, version, err := st.tree.SaveVersion()
	if err != nil {
		return nil, err
	}
	if hash == nil {
		return nil, fmt.Errorf("IVAL root is not set")
	}

	// Provide a reference to load this version in the future from the state hash
	st.setVersion(hash, version)

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

// -------
// ACCOUNT

func (st *State) GlobalAccount() *account.Account {
	gAcc, _ := st.GetAccount(crypto.GlobalAddress)
	return gAcc
}

func (st *State) HasAccount(addr crypto.Address) bool {
	return st.tree.Has(accountKey(addr))
}

func (st *State) GetAccount(addr crypto.Address) (*account.Account, error) {
	st.Lock()
	defer st.Unlock()

	_, bs := st.tree.Get(accountKey(addr))
	if bs == nil {
		return nil, fmt.Errorf("There is no account with this address %s", addr.String())
	}
	acc, err := account.AccountFromBytes(bs)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode account: %v", err)
	}

	return acc, nil
}

func (st *State) AccountCount() int {
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

// HasPermissions ensures that an account has required permissions
func (st *State) HasPermissions(acc *account.Account, perm account.Permissions) bool {
	if !permission.EnsureValid(perm) {
		return false
	}

	gAcc := st.GlobalAccount()
	if gAcc.HasPermissions(perm) {
		return true
	}

	if acc.HasPermissions(perm) {
		return true
	}

	return false
}

func (st *State) HasSendPermission(acc *account.Account) bool {
	return st.HasPermissions(acc, permission.Send)
}

func (st *State) HasCallPermission(acc *account.Account) bool {
	return st.HasPermissions(acc, permission.Call)
}

func (st *State) HasCreateContractPermission(acc *account.Account) bool {
	return st.HasPermissions(acc, permission.CreateContract)
}

func (st *State) HasCreateAccountPermission(acc *account.Account) bool {
	return st.HasPermissions(acc, permission.CreateAccount)
}

func (st *State) HasBondPermission(acc *account.Account) bool {
	return st.HasPermissions(acc, permission.Bond)
}

func (st *State) HasModifyPermission(acc *account.Account) bool {
	return st.HasPermissions(acc, permission.ModifyPermission)
}

// ---------
// VALIDATOR

func (st *State) HasValidator(addr crypto.Address) bool {
	return st.tree.Has(validatorKey(addr))
}

func (st *State) GetValidator(addr crypto.Address) (*validator.Validator, error) {
	st.Lock()
	defer st.Unlock()

	_, bytes := st.tree.Get(validatorKey(addr))
	if bytes == nil {
		return nil, fmt.Errorf("There is no validator with this address %s", addr.String())
	}
	val, err := validator.ValidatorFromBytes(bytes)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode validator: %v", err)
	}

	return val, nil
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

// -------
// STORAGE

func (st *State) GetStorage(addr crypto.Address, key binary.Word256) (binary.Word256, error) {
	_, value := st.tree.Get(prefixedKey(storagePrefix, addr.RawBytes(), key.Bytes()))
	return binary.LeftPadWord256(value), nil
}

func (st *State) IterateStorage(addr crypto.Address,
	consumer func(key, value binary.Word256) (stop bool)) (stopped bool, err error) {
	stopped = st.tree.IterateRange(storageStart, storageEnd, true, func(key []byte, value []byte) (stop bool) {
		// Note: no left padding should occur unless there is a bug and non-words have been writte to this storage tree
		if len(key) != binary.Word256Length {
			err = fmt.Errorf("key '%X' stored for account %s is not a %v-byte word",
				key, addr, binary.Word256Length)
			return true
		}
		if len(value) != binary.Word256Length {
			err = fmt.Errorf("value '%X' stored for account %s is not a %v-byte word",
				key, addr, binary.Word256Length)
			return true
		}
		return consumer(binary.LeftPadWord256(key), binary.LeftPadWord256(value))
	})
	return
}

func (st *State) ByzantineValidator(addr crypto.Address) error {
	return st.removeValidator(addr)
}

func (st *State) IncentivizeValidator(addr crypto.Address, fee uint64) error {
	val, err := st.GetValidator(addr)
	if err != nil {
		return fmt.Errorf("Could not get proposer information: %v", err)
	}
	val.AddToStake(fee)
	return st.updateValidator(val)
}

/// -----------------------------
/// Modifier methods are private.
func (st *State) updateAccount(acc *account.Account) error {
	st.Lock()
	defer st.Unlock()

	bs, err := acc.Encode()
	if err != nil {
		return err
	}

	st.tree.Set(accountKey(acc.Address()), bs)
	return nil
}

func (st *State) updateValidator(val *validator.Validator) error {
	st.Lock()
	defer st.Unlock()

	bs, err := val.Encode()
	if err != nil {
		return err
	}

	st.tree.Set(validatorKey(val.Address()), bs)
	return nil
}

func (st *State) removeValidator(addr crypto.Address) error {
	st.Lock()
	defer st.Unlock()

	st.tree.Remove(validatorKey(addr))
	return nil
}

func (st *State) setStorage(addr crypto.Address, key, value binary.Word256) error {
	if value == binary.Zero256 {
		st.tree.Remove(key.Bytes())
	} else {
		st.tree.Set(prefixedKey(storagePrefix, addr.RawBytes(), key.Bytes()), value.Bytes())
	}
	return nil
}
