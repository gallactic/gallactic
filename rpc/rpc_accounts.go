package rpc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/gallactic/gallactic/core/account"
)

func NewAccountFilterFactory() *FilterFactory {
	filterFactory := NewFilterFactory()

	filterFactory.RegisterFilterPool("code", &sync.Pool{
		New: func() interface{} {
			return &AccountCodeFilter{}
		},
	})

	filterFactory.RegisterFilterPool("balance", &sync.Pool{
		New: func() interface{} {
			return &AccountBalanceFilter{}
		},
	})

	return filterFactory
}

// Filter for account code.
// Ops: == or !=
// Could be used to match against nil, to see if an account is a contract account.
type AccountCodeFilter struct {
	op    string
	value []byte
	match func([]byte, []byte) bool
}

func (acf *AccountCodeFilter) Configure(fd *FilterData) error {
	op := fd.Op
	val, err := hex.DecodeString(fd.Value)

	if err != nil {
		return fmt.Errorf("Wrong value type.")
	}
	if op == "==" {
		acf.match = func(a, b []byte) bool {
			return bytes.Equal(a, b)
		}
	} else if op == "!=" {
		acf.match = func(a, b []byte) bool {
			return !bytes.Equal(a, b)
		}
	} else {
		return fmt.Errorf("Op: " + acf.op + " is not supported for 'code' filtering")
	}
	acf.op = op
	acf.value = val
	return nil
}

func (acf *AccountCodeFilter) Match(v interface{}) bool {
	acc, ok := v.(*account.Account)
	if !ok {
		return false
	}
	return acf.match(acc.Code(), acf.value)
}

// Filter for account balance.
// Ops: All
type AccountBalanceFilter struct {
	op    string
	value uint64
	match func(uint64, uint64) bool
}

func (abf *AccountBalanceFilter) Configure(fd *FilterData) error {
	val, err := ParseNumberValue(fd.Value)
	if err != nil {
		return err
	}
	match, err2 := GetRangeFilter(fd.Op, "balance")
	if err2 != nil {
		return err2
	}
	abf.match = match
	abf.op = fd.Op
	abf.value = uint64(val)
	return nil
}

func (abf *AccountBalanceFilter) Match(v interface{}) bool {
	acc, ok := v.(*account.Account)
	if !ok {
		return false
	}
	return abf.match(acc.Balance(), abf.value)
}
