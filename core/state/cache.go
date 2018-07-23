package state

import (
	"errors"
	"sync"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
)

var (
	errValidatorChanged = errors.New("Validator has changed before in this height")
)

const (
	addToPool       = 0 /// Bonding transaction
	removeFromPool  = 1 /// Unbonding transaction
	addToSet        = 2 /// Sortition transaction
	updateValidator = 3 /// No transaction for this state change, but it is deterministic and enforce by the protocol
)

type Cache struct {
	sync.Mutex
	name       string
	st         *State
	valChanges map[crypto.Address]*validatorInfo
	accChanges map[crypto.Address]*accountInfo
}

type validatorInfo struct {
	status int
	val    *validator.Validator
}

type accountInfo struct {
	acc *account.Account
}

type CacheOption func(*Cache)

func NewCache(st *State, options ...CacheOption) *Cache {
	ch := &Cache{
		st:         st,
		valChanges: make(map[crypto.Address]*validatorInfo),
		accChanges: make(map[crypto.Address]*accountInfo),
	}
	for _, option := range options {
		option(ch)
	}

	return ch
}

func Name(name string) CacheOption {
	return func(ch *Cache) {
		ch.name = name
	}
}

func (ch *Cache) Reset() {
	for a := range ch.accChanges {
		delete(ch.accChanges, a)
	}

	for v := range ch.valChanges {
		delete(ch.valChanges, v)
	}
}

//
func (ch *Cache) Flush() error {
	ch.Lock()
	defer ch.Unlock()
	for _, i := range ch.accChanges {
		if err := ch.st.UpdateAccount(i.acc); err != nil {
			return err
		}
	}
	/*
		for _, valInfo := range ch.valChanges {
			switch valInfo.status {
			case addToSet:
				if err := pool.set.Join(valInfo.validator); err != nil {
					return err
				}

			case updateValidator, addToPool:
				bytes, err := valInfo.validator.Encode()
				if err != nil {
					return err
				}

				if !pool.tree.Set(validatorKey(valInfo.validator.Address()), bytes) {
					return fmt.Errorf("Unable to set validator to tree")
				}

			case removeFromPool:
				if err := pool.set.ForceLeave(valInfo.validator); err != nil {
					/// when the node is byzantine
					return err
				}

				if _, removed := pool.tree.Remove(validatorKey(valInfo.validator.Address())); !removed {
					return fmt.Errorf("Unable to remove validator from tree")
				}
			}
		}

		pool.clear()
	*/
	return nil
}

func (ch *Cache) GetAccount(addr crypto.Address) *account.Account {
	ch.Lock()
	defer ch.Unlock()

	i, ok := ch.accChanges[addr]
	if ok {
		return i.acc
	}

	return ch.st.GetAccount(addr)
}

func (ch *Cache) UpdateAccount(acc *account.Account) error {
	ch.Lock()
	defer ch.Unlock()

	ch.accChanges[acc.Address()] = &accountInfo{acc: acc}
	return nil

}

func (ch *Cache) GetValidator(addr crypto.Address) *validator.Validator {
	ch.Lock()
	defer ch.Unlock()
	/*
		valInfo, ok := pool.changes[addr]
		if ok {
			return valInfo.validator
		}

		_, bytes := pool.tree.Get(validatorKey(addr))
		if bytes == nil {
			return nil
		}
		val, err := validator.ValidatorFromBytes(bytes)
		if err != nil {
			panic("Unable to decode encoded validator")
		}

		return val
	*/
	return nil
}

func (ch *Cache) HasPermissions(acc *account.Account, perm account.Permissions) bool {
	return false
}

func (ch *Cache) AddToPool(val *validator.Validator) error {
	ch.Lock()
	defer ch.Unlock()
	/*
		address := validator.Address()
		_, ok := pool.changes[address]
		if ok {
			return errValidatorChanged
		}

		pool.changes[address] = &validatorInfo{
			status:    addToPool,
			validator: validator,
		}
	*/
	return nil
}

func (ch *Cache) AddToSet(val *validator.Validator) error {
	ch.Lock()
	defer ch.Unlock()
	/*
		address := validator.Address()
		_, ok := pool.changes[address]
		if ok {
			return errValidatorChanged
		}

		pool.changes[address] = &validatorInfo{
			status:    addToSet,
			validator: validator,
		}
	*/
	return nil
}

func (ch *Cache) RemoveFromPool(val *validator.Validator) error {
	ch.Lock()
	defer ch.Unlock()
	/*
		address := validator.Address()
		_, ok := pool.changes[address]
		if ok {
			return errValidatorChanged
		}

		pool.changes[address] = &validatorInfo{
			status:    removeFromPool,
			validator: validator,
		}
	*/
	return nil
}

func (ch *Cache) UpdateValidator(val *validator.Validator) error {
	ch.Lock()
	defer ch.Unlock()
	/*
		address := validator.Address()
		_, ok := pool.changes[address]
		if ok {
			return errValidatorChanged
		}

		pool.changes[address] = &validatorInfo{
			status:    updateValidator,
			validator: validator,
		}
	*/
	return nil
}
