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
	status    int
	validator *validator.Validator
}

type accountInfo struct {
	account *account.Account
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
		if err := c.state.UpdateAccount(i.account); err != nil {
			return err
		}
	}
	for addr, i := range c.valChanges {
		switch i.status {
		case addToSet:
			if err := set.Join(i.validator); err != nil {
				return err
			}

		case updateValidator, addToPool:
			if err := c.state.UpdateValidator(i.validator); err != nil {
				return err
			}

		case removeFromPool:
			if err := set.ForceLeave(addr); err != nil {
				/// when the node is byzantine
				return err
			}

			if err := c.state.RemoveValidator(addr); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Cache) HasAccount(addr crypto.Address) bool {
	c.Lock()
	defer c.Unlock()

	_, ok := c.accChanges[addr]
	if ok {
		return true
	}

	return c.state.HasAccount(addr)
}

func (c *Cache) GetAccount(addr crypto.Address) (*account.Account, error) {
	c.Lock()
	defer c.Unlock()

	i, ok := c.accChanges[addr]
	if ok {
		return i.account, nil
	}

	return c.state.GetAccount(addr)
}

func (c *Cache) UpdateAccount(acc *account.Account) error {
	c.Lock()
	defer c.Unlock()

	c.accChanges[acc.Address()] = &accountInfo{account: acc}
	return nil
}

func (c *Cache) HasPermissions(acc *account.Account, perm account.Permissions) bool {
	return c.state.HasPermissions(acc, perm)
}

func (c *Cache) HasValidator(addr crypto.Address) bool {
	c.Lock()
	defer c.Unlock()

	_, ok := c.valChanges[addr]
	if ok {
		return true
	}

	return c.state.HasValidator(addr)
}

func (c *Cache) GetValidator(addr crypto.Address) (*validator.Validator, error) {
	c.Lock()
	defer c.Unlock()

	i, ok := c.valChanges[addr]
	if ok {
		return i.validator, nil
	}

	return c.state.GetValidator(addr)
}

func (c *Cache) AddToPool(val *validator.Validator) error {
	c.Lock()
	defer c.Unlock()

	a := val.Address()
	_, ok := c.valChanges[a]
	if ok {
		return errValidatorChanged
	}

	c.valChanges[a] = &validatorInfo{
		status:    addToPool,
		validator: val,
	}
	return nil
}

func (c *Cache) AddToSet(val *validator.Validator) error {
	c.Lock()
	defer c.Unlock()

	a := val.Address()
	_, ok := c.valChanges[a]
	if ok {
		return errValidatorChanged
	}

	c.valChanges[a] = &validatorInfo{
		status:    addToSet,
		validator: val,
	}
	return nil
}

func (c *Cache) RemoveFromPool(val *validator.Validator) error {
	c.Lock()
	defer c.Unlock()

	a := val.Address()
	_, ok := c.valChanges[a]
	if ok {
		return errValidatorChanged
	}

	c.valChanges[a] = &validatorInfo{
		status:    removeFromPool,
		validator: val,
	}
	return nil
}

func (c *Cache) UpdateValidator(val *validator.Validator) error {
	c.Lock()
	defer c.Unlock()

	a := val.Address()
	_, ok := c.valChanges[a]
	if ok {
		return errValidatorChanged
	}

	c.valChanges[a] = &validatorInfo{
		status:    updateValidator,
		validator: val,
	}
	return nil
}
