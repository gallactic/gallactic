package state

import (
	"fmt"
	"sync"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/crypto"
	"github.com/tendermint/iavl"
)

type AccountPool struct {
	sync.Mutex
	tree    *iavl.Tree
	changes map[crypto.Address]*account.Account
}

func NewAccountPool(tree *iavl.Tree) *AccountPool {
	return &AccountPool{
		tree:    tree,
		changes: make(map[crypto.Address]*account.Account),
	}
}

func (pool *AccountPool) UpdateTree(tree *iavl.Tree) error {
	if len(pool.changes) > 0 {
		return fmt.Errorf("There are changes are waiting for commit in AccountPool")
	}

	pool.tree = tree
	return nil
}

func (pool *AccountPool) clear() {

	for c := range pool.changes {
		delete(pool.changes, c)
	}
}

//
func (pool *AccountPool) flush() error {
	pool.Lock()
	defer pool.Unlock()

	for _, acc := range pool.changes {
		bytes, err := acc.Encode()
		if err != nil {
			return err
		}

		pool.tree.Set(accountKey(acc.Address()), bytes)
	}

	pool.clear()

	return nil
}

func (pool *AccountPool) GetAccount(address crypto.Address) *account.Account {
	pool.Lock()
	defer pool.Unlock()

	acc, ok := pool.changes[address]
	if ok {
		return acc
	}

	_, bytes := pool.tree.Get(accountKey(address))
	if bytes == nil {
		return nil
	}
	acc, err := account.AccountFromBytes(bytes)
	if err != nil {
		panic("Unable to decode encoded Account")
	}

	return acc
}

func (pool *AccountPool) UpdateAccount(acc *account.Account) error {
	pool.Lock()
	defer pool.Unlock()

	address := acc.Address()

	pool.changes[address] = acc

	return nil
}

func (pool *AccountPool) Count() int {
	count := 0
	pool.Iterate(func(validator *account.Account) (stop bool) {
		count++
		return false
	})
	return count
}

func (pool *AccountPool) Iterate(consumer func(*account.Account) (stop bool)) (stopped bool, err error) {
	stopped = pool.tree.IterateRange(accountsStart, accountsEnd, true, func(key, bs []byte) bool {
		acc, err := account.AccountFromBytes(bs)
		if err != nil {
			return true
		}
		return consumer(acc)
	})
	return
}
