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
	cache := &Cache{
		st:         st,
		valChanges: make(map[crypto.Address]*validatorInfo),
		accChanges: make(map[crypto.Address]*accountInfo),
	}
	for _, option := range options {
		option(cache)
	}

	return cache
}

func Name(name string) CacheOption {
	return func(cache *Cache) {
		cache.name = name
	}
}

func (c *Cache) Reset() {
	for a := range c.accChanges {
		delete(c.accChanges, a)
	}

	for v := range c.valChanges {
		delete(c.valChanges, v)
	}
}

//
func (c *Cache) Flush() error {
	c.Lock()
	defer c.Unlock()
	for _, i := range c.accChanges {
		if err := c.st.UpdateAccount(i.acc); err != nil {
			return err
		}
	}
	/*
		for _, valInfo := range c.valChanges {
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

func (c *Cache) GetAccount(addr crypto.Address) *account.Account {
	c.Lock()
	defer c.Unlock()

	i, ok := c.accChanges[addr]
	if ok {
		return i.acc
	}

	return c.st.GetAccount(addr)
}

func (c *Cache) UpdateAccount(acc *account.Account) error {
	c.Lock()
	defer c.Unlock()

	c.accChanges[acc.Address()] = &accountInfo{acc: acc}
	return nil

}

func (c *Cache) GetValidator(addr crypto.Address) *validator.Validator {
	c.Lock()
	defer c.Unlock()
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

func (c *Cache) AddToPool(val *validator.Validator) error {
	c.Lock()
	defer c.Unlock()
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

func (c *Cache) AddToSet(val *validator.Validator) error {
	c.Lock()
	defer c.Unlock()
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

func (c *Cache) RemoveFromPool(val *validator.Validator) error {
	c.Lock()
	defer c.Unlock()
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

func (c *Cache) UpdateValidator(val *validator.Validator) error {
	c.Lock()
	defer c.Unlock()
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
