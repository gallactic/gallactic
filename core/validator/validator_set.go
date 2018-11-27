package validator

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/gallactic/gallactic/crypto"
	"github.com/hyperledger/burrow/logging"
	tmRPC "github.com/tendermint/tendermint/rpc/core"
	tmRPCTypes "github.com/tendermint/tendermint/rpc/core/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

const maximumTendermintNode = 90
const minimumTendermintNode = 6

type validatorListProxy interface {
	validators(height int64) (*tmRPCTypes.ResultValidators, error)
}

type _validatorListProxy struct{}

func (vlp _validatorListProxy) validators(height int64) (*tmRPCTypes.ResultValidators, error) {
	result, err := tmRPC.Validators(&height)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type ValidatorSet struct {
	proxy        validatorListProxy
	maximumPower int
	leavers      map[crypto.Address]*Validator
	validators   map[crypto.Address]*Validator
	logger       *logging.Logger
}

func NewValidatorSet(validators map[crypto.Address]*Validator, maximumPower int, logger *logging.Logger) *ValidatorSet {
	set := &ValidatorSet{
		validators:   validators,
		leavers:      make(map[crypto.Address]*Validator),
		maximumPower: maximumPower,
		proxy:        _validatorListProxy{},
		logger:       logger,
	}
	return set
}

// TotalPower equals to the number of validator in the set
func (set *ValidatorSet) TotalPower() int {
	return len(set.validators)
}

func (set *ValidatorSet) UpdateMaximumPower(maximumPower int) {
	if maximumPower > maximumTendermintNode {
		maximumPower = maximumTendermintNode
	}

	if maximumPower < minimumTendermintNode {
		maximumPower = minimumTendermintNode
	}

	set.maximumPower = maximumPower
}

func (set *ValidatorSet) MaximumPower() int {
	return set.maximumPower
}

func (set *ValidatorSet) AdjustPower(height int64) error {
	/// first clear the slice
	for k := range set.leavers {
		delete(set.leavers, k)
	}

	dif := set.TotalPower() - set.maximumPower
	if dif <= 0 {
		return nil
	}

	limit := set.maximumPower/3 - 1
	if dif > limit {
		dif = limit
	}

	/// copy of validator set in round m
	var vals1 []*Validator
	var vals2 []*tmTypes.Validator

	/// initialize the validator slice
	vals1 = make([]*Validator, len(set.validators))
	i := 0
	for _, v := range set.validators {
		vals1[i] = v
		i++
	}

	sort.SliceStable(vals1, func(i, j int) bool {
		return bytes.Compare(vals1[i].Address().RawBytes(), vals1[j].Address().RawBytes()) < 0
	})

	for {
		height--

		if height > 0 {
			result, err := set.proxy.validators(height)
			if err != nil {
				return err
			}

			/// copy of validator set in round n (n<m)
			vals2 = result.Validators
			sort.SliceStable(vals2, func(i, j int) bool {
				return bytes.Compare(vals2[i].Address.Bytes(), vals2[j].Address.Bytes()) < 0
			})
		} else {
			/// genesis validators
			vals2 = vals2[1:]
		}

		r := make([]int, 0)
		var i, j int = 0, 0
		for i < len(vals1) && j < len(vals2) {
			val1 := vals1[i]
			val2 := vals2[j]

			cmp := bytes.Compare(val1.Address().RawBytes(), val2.Address.Bytes())
			if cmp == 0 {
				i++
				j++
			} else if cmp < 0 {
				r = append(r, i)
				i++
			} else {
				j++
			}
		}

		/// if at the end of slice_a there are some elements bigger than last element in slice_b
		for z := i; z < len(vals1); z++ {
			r = append(r, i)
		}

		// println(fmt.Sprintf("%v", vals1))
		// println(fmt.Sprintf("%v", vals2))
		// println(fmt.Sprintf("%v", r))

		var n int
		for _, m := range r {
			vals1 = append(vals1[:m-n], vals1[m-n+1:]...)
			n++

			/// Not removing more than requested
			if len(vals1) == dif {
				break
			}
		}

		if len(vals1) == dif {
			break
		}
	}

	// println(fmt.Sprintf("%v", vals1))
	for _, v := range vals1 {
		a := v.Address()
		v, ok := set.validators[a]
		if ok {
			delete(set.validators, a)
			set.leavers[a] = v
		}
	}

	return nil
}

func (set *ValidatorSet) Validators() map[crypto.Address]*Validator {
	return set.validators
}

func (set *ValidatorSet) Leavers() map[crypto.Address]*Validator {
	return set.leavers
}

func (set *ValidatorSet) Join(val *Validator) error {
	if set.Contains(val.Address()) {
		return fmt.Errorf("This validator currently is in the set: %v", val.Address())
	}

	/// Welcome to the party!
	set.validators[val.Address()] = val
	return nil
}

func (set *ValidatorSet) ForceLeave(addr crypto.Address) error {
	if !set.Contains(addr) {
		return fmt.Errorf("This validator currently is not in the set: %v", addr)
	}

	_, ok := set.validators[addr]
	if ok {
		delete(set.validators, addr)
	}

	return nil
}

func (set *ValidatorSet) Contains(addr crypto.Address) bool {
	_, ok := set.validators[addr]
	return ok
}
