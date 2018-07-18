package state

import (
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/crypto"
)

func (st *State) GlobalAccount() *account.Account {
	st.Lock()
	defer st.Unlock()

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
