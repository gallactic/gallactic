package state

import (
	"bytes"
	"errors"
	"sync"

	"github.com/gallactic/gallactic/common/orderedmap"

	"github.com/gallactic/gallactic/common/binary"
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
	valChanges *orderedmap.OrderedMap
	accChanges *orderedmap.OrderedMap
}

type validatorInfo struct {
	status    int
	validator *validator.Validator
}

type accountInfo struct {
	account  *account.Account
	storages *orderedmap.OrderedMap
	removed  bool
}

type CacheOption func(*Cache)

func lessFn(l, r interface{}) bool {
	return bytes.Compare(l.(crypto.Address).RawBytes(), r.(crypto.Address).RawBytes()) < 0
}

func lessFn2(l, r interface{}) bool {
	return bytes.Compare(l.(binary.Word256).Bytes(), r.(binary.Word256).Bytes()) < 0
}
func NewCache(state *State) *Cache {
	ch := &Cache{
		state:      state,
		valChanges: orderedmap.NewMap(lessFn),
		accChanges: orderedmap.NewMap(lessFn),
	}
	return ch
}

func (c *Cache) Reset() {
	c.Lock()
	defer c.Unlock()

	c.accChanges = orderedmap.NewMap(lessFn)
	c.valChanges = orderedmap.NewMap(lessFn)
}

//
func (c *Cache) Flush(set *validator.ValidatorSet) error {
	c.Lock()
	defer c.Unlock()

	c.accChanges.Iter(func(key, value interface{}) (more bool) {
		addr := key.(crypto.Address)
		i := value.(*accountInfo)
		if i.removed {
			if err := c.state.removeAccount(addr); err != nil {
				panic(err)
			}
		} else {
			if err := c.state.updateAccount(i.account); err != nil {
				panic(err)
			}

			if i.storages != nil {
				i.storages.Iter(func(k, v interface{}) (more bool) {
					if err := c.state.setStorage(i.account.Address(), k.(binary.Word256), v.(binary.Word256)); err != nil {
						panic(err)
					}
					return true
				})
			}
		}
		return true
	})

	c.valChanges.Iter(func(key, value interface{}) (more bool) {
		addr := key.(crypto.Address)
		i := value.(*validatorInfo)

		switch i.status {
		case addToSet:
			if err := set.Join(i.validator); err != nil {
				panic(err)
			}

		case updateValidator, addToPool:
			if err := c.state.updateValidator(i.validator); err != nil {
				panic(err)
			}

		case removeFromPool:
			if err := set.ForceLeave(addr); err != nil {
				/// when the node is byzantine
				panic(err)
			}

			if err := c.state.removeValidator(addr); err != nil {
				panic(err)
			}
		}
		return true
	})

	/// reset cache
	c.accChanges = orderedmap.NewMap(lessFn)
	c.valChanges = orderedmap.NewMap(lessFn)

	return nil
}

func (c *Cache) HasAccount(addr crypto.Address) bool {
	c.Lock()
	defer c.Unlock()

	_, ok := c.accChanges.GetOk(addr)
	if ok {
		return true
	}

	return c.state.HasAccount(addr)
}

func (c *Cache) GetAccount(addr crypto.Address) (*account.Account, error) {
	c.Lock()
	defer c.Unlock()

	i, ok := c.accChanges.GetOk(addr)
	if ok {
		return i.(*accountInfo).account, nil
	}

	return c.state.GetAccount(addr)
}

func (c *Cache) UpdateAccount(acc *account.Account) error {
	c.Lock()
	defer c.Unlock()

	addr := acc.Address()
	i, ok := c.accChanges.GetOk(addr)
	if ok {
		i.(*accountInfo).account = acc
	} else {
		c.accChanges.Set(addr, &accountInfo{account: acc})
	}

	return nil
}

func (c *Cache) RemoveAccount(addr crypto.Address) error {
	c.Lock()
	defer c.Unlock()

	_, ok := c.accChanges.GetOk(addr)
	if ok {
		c.accChanges.Unset(addr) /// simply remove it from cache
	} else {
		c.accChanges.Set(addr, &accountInfo{removed: true})
	}

	return nil
}

func (c *Cache) HasPermissions(acc *account.Account, perm account.Permissions) bool {
	return c.state.HasPermissions(acc, perm)
}

func (c *Cache) HasValidator(addr crypto.Address) bool {
	c.Lock()
	defer c.Unlock()

	_, ok := c.valChanges.GetOk(addr)
	if ok {
		return true
	}

	return c.state.HasValidator(addr)
}

func (c *Cache) GetValidator(addr crypto.Address) (*validator.Validator, error) {
	c.Lock()
	defer c.Unlock()

	i, ok := c.valChanges.GetOk(addr)
	if ok {
		return i.(*validatorInfo).validator, nil
	}

	return c.state.GetValidator(addr)
}

func (c *Cache) AddToPool(val *validator.Validator) error {
	c.Lock()
	defer c.Unlock()

	addr := val.Address()
	_, ok := c.valChanges.GetOk(addr)
	if ok {
		return errValidatorChanged
	}

	c.valChanges.Set(addr, &validatorInfo{addToPool, val})
	return nil
}

func (c *Cache) AddToSet(val *validator.Validator) error {
	c.Lock()
	defer c.Unlock()

	addr := val.Address()
	_, ok := c.valChanges.GetOk(addr)
	if ok {
		return errValidatorChanged
	}

	c.valChanges.Set(addr, &validatorInfo{addToSet, val})
	return nil
}

func (c *Cache) RemoveFromPool(val *validator.Validator) error {
	c.Lock()
	defer c.Unlock()

	addr := val.Address()
	_, ok := c.valChanges.GetOk(addr)
	if ok {
		return errValidatorChanged
	}

	c.valChanges.Set(addr, &validatorInfo{removeFromPool, val})
	return nil
}

func (c *Cache) UpdateValidator(val *validator.Validator) error {
	c.Lock()
	defer c.Unlock()

	addr := val.Address()
	_, ok := c.valChanges.GetOk(addr)
	if ok {
		return errValidatorChanged
	}

	c.valChanges.Set(addr, &validatorInfo{updateValidator, val})
	return nil
}

func (c *Cache) GetStorage(addr crypto.Address, key binary.Word256) (binary.Word256, error) {
	c.Lock()
	defer c.Unlock()

	i, ok := c.accChanges.GetOk(addr)
	if ok {
		if i.(*accountInfo).storages != nil {
			s, ok := i.(*accountInfo).storages.GetOk(key)
			if ok {
				return s.(binary.Word256), nil
			}
		}
	}

	return c.state.GetStorage(addr, key)
}

func (c *Cache) SetStorage(addr crypto.Address, key, value binary.Word256) error {
	c.Lock()
	defer c.Unlock()

	i, ok := c.accChanges.GetOk(addr)
	if !ok {
		acc, _ := c.state.GetAccount(addr)
		if acc == nil {
			acc, _ = account.NewContractAccount(addr)
		}

		i = &accountInfo{account: acc}
		c.accChanges.Set(addr, i)
	}

	if i.(*accountInfo).storages == nil {
		i.(*accountInfo).storages = orderedmap.NewMap(lessFn2)
	}

	i.(*accountInfo).storages.Set(key, value)
	return nil
}
