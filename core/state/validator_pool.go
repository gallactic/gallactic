package state

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/tendermint/iavl"
)

var (
	errValidatorChanged = errors.New("Validator has changed before in this height")
)

const (
	addToPool       = iota /// Bonding transaction
	removeFromPool         /// Unbonding transaction
	addToSet               /// Sortition transaction
	updateValidator        /// No transaction for this state change, but it is deterministic and enforce by the protocol
)

type ValidatorPool struct {
	sync.Mutex
	tree    *iavl.Tree
	set     *ValidatorSet
	changes map[crypto.Address]*validatorInfo
}

type validatorInfo struct {
	status    int
	validator *validator.Validator
}

func NewValidatorPool(tree *iavl.Tree) *ValidatorPool {
	return &ValidatorPool{
		tree:    tree,
		changes: make(map[crypto.Address]*validatorInfo),
	}
}

func (pool *ValidatorPool) UpdateTree(tree *iavl.Tree) error {
	if len(pool.changes) > 0 {
		return fmt.Errorf("There are changes are waiting for commit in ValidatorPool")
	}

	pool.tree = tree
	return nil
}

func (pool *ValidatorPool) clear() {
	for c := range pool.changes {
		delete(pool.changes, c)
	}
}

//
func (pool *ValidatorPool) flush() error {
	pool.Lock()
	defer pool.Unlock()

	for _, valInfo := range pool.changes {
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

	return nil
}

func (pool *ValidatorPool) GetValidator(address crypto.Address) *validator.Validator {
	pool.Lock()
	defer pool.Unlock()

	valInfo, ok := pool.changes[address]
	if ok {
		return valInfo.validator
	}

	_, bytes := pool.tree.Get(validatorKey(address))
	if bytes == nil {
		return nil
	}
	val, err := validator.ValidatorFromBytes(bytes)
	if err != nil {
		panic("Unable to decode encoded validator")
	}

	return val
}

func (pool *ValidatorPool) AddToPool(validator *validator.Validator) error {
	pool.Lock()
	defer pool.Unlock()

	address := validator.Address()
	_, ok := pool.changes[address]
	if ok {
		return errValidatorChanged
	}

	pool.changes[address] = &validatorInfo{
		status:    addToPool,
		validator: validator,
	}
	return nil
}

func (pool *ValidatorPool) AddToSet(validator *validator.Validator) error {
	pool.Lock()
	defer pool.Unlock()

	address := validator.Address()
	_, ok := pool.changes[address]
	if ok {
		return errValidatorChanged
	}

	pool.changes[address] = &validatorInfo{
		status:    addToSet,
		validator: validator,
	}
	return nil
}

func (pool *ValidatorPool) RemoveFromPool(validator *validator.Validator) error {
	pool.Lock()
	defer pool.Unlock()

	address := validator.Address()
	_, ok := pool.changes[address]
	if ok {
		return errValidatorChanged
	}

	pool.changes[address] = &validatorInfo{
		status:    removeFromPool,
		validator: validator,
	}
	return nil
}

func (pool *ValidatorPool) UpdateValidator(validator *validator.Validator) error {
	pool.Lock()
	defer pool.Unlock()

	address := validator.Address()
	_, ok := pool.changes[address]
	if ok {
		return errValidatorChanged
	}

	pool.changes[address] = &validatorInfo{
		status:    updateValidator,
		validator: validator,
	}
	return nil
}

func (pool *ValidatorPool) Count() int {
	count := 0
	pool.Iterate(func(validator *validator.Validator) (stop bool) {
		count++
		return false
	})
	return count
}

func (pool *ValidatorPool) Iterate(consumer func(*validator.Validator) (stop bool)) (stopped bool, err error) {
	return pool.tree.IterateRange(validatorStart, validatorEnd, true, func(key []byte, bs []byte) (stop bool) {
		validator, err := validator.ValidatorFromBytes(bs)
		if err != nil {
			return true
		}
		return consumer(validator)
	}), nil

}
