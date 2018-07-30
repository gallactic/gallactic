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
	state      *State
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

func NewCache(state *State, options ...CacheOption) *Cache {
	ch := &Cache{
		state:      state,
		valChanges: make(map[crypto.Address]*validatorInfo),
		accChanges: make(map[crypto.Address]*accountInfo),
	}
	for _, option := range options {
		option(ch)
	}

	return ch
}

func Name(name string) CacheOption {
	return func(c *Cache) {
		c.name = name
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
func (c *Cache) Flush(set *validator.ValidatorSet) error {
	c.Lock()
	defer c.Unlock()
	for _, i := range c.accChanges {
		if err := c.state.UpdateAccount(i.acc); err != nil {
			return err
		}
	}
	for _, i := range c.valChanges {
		switch i.status {
		case addToSet:
			if err := set.Join(i.val); err != nil {
				return err
			}

		case updateValidator, addToPool:
			if err := c.state.UpdateValidator(i.val); err != nil {
				return err
			}

		case removeFromPool:
			if err := set.ForceLeave(i.val); err != nil {
				/// when the node is byzantine
				return err
			}

			if err := c.state.RemoveValidator(i.val); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Cache) GetAccount(addr crypto.Address) (*account.Account, error) {
	c.Lock()
	defer c.Unlock()

	i, ok := c.accChanges[addr]
	if ok {
		return i.acc, nil
	}

	return c.state.GetAccount(addr)
}

func (c *Cache) UpdateAccount(acc *account.Account) error {
	c.Lock()
	defer c.Unlock()

	c.accChanges[acc.Address()] = &accountInfo{acc: acc}
	return nil
}

func (c *Cache) HasPermissions(acc *account.Account, perm account.Permissions) bool {
	return c.state.HasPermissions(acc, perm)
}

func (c *Cache) GetValidator(addr crypto.Address) (*validator.Validator, error) {
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
	return nil, nil
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
