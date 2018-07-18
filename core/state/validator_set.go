package state

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	tmRPC "github.com/tendermint/tendermint/rpc/core"
	tmRPCTypes "github.com/tendermint/tendermint/rpc/core/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

var stateKey = []byte("ValidatorSet")

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
	data    validatorSetData
	proxy   validatorListProxy
	leavers []*validator.Validator /// TODO: change to map
}

type validatorSetData struct {
	MaximumPower int                    `json:"maximumPower"`
	Validators   []*validator.Validator `json:"validators"`
}

func NewValidatorSet(validators []*validator.Validator) *ValidatorSet {
	set := &ValidatorSet{
		data: validatorSetData{
			Validators:   validators,
			MaximumPower: minimumTendermintNode,
		},
		proxy: _validatorListProxy{},
	}
	return set
}

// TotalPower equals to the number of validator in the set
func (set *ValidatorSet) TotalPower() int {
	return len(set.data.Validators)
}

func (set *ValidatorSet) SetMaximumPower(maximumPower int) {
	if maximumPower > maximumTendermintNode {
		maximumPower = maximumTendermintNode
	}

	if maximumPower < minimumTendermintNode {
		maximumPower = minimumTendermintNode
	}

	set.data.MaximumPower = maximumPower
}

func (set *ValidatorSet) MaximumPower() int {
	return set.data.MaximumPower
}

func (set *ValidatorSet) AdjustPower(height int64) error {
	/// first clear the slice
	set.leavers = set.leavers[:0]

	dif := set.TotalPower() - set.data.MaximumPower
	if dif <= 0 {
		return nil
	}

	limit := set.data.MaximumPower/3 - 1
	if dif > limit {
		dif = limit
	}

	/// copy of validator set in round m
	var vals1 []*validator.Validator
	var vals2 []*tmTypes.Validator

	vals1 = make([]*validator.Validator, len(set.data.Validators))
	copy(vals1, set.data.Validators)
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
			vals2 = vals2[1:len(vals2)]
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
	for _, v1 := range vals1 {
		for i, v2 := range set.data.Validators {
			if v1.Address().EqualsTo(v2.Address()) {
				set.data.Validators = append(set.data.Validators[:i], set.data.Validators[i+1:]...)
				set.leavers = append(set.leavers, v2)
				break
			}
		}
	}

	return nil
}

func (set *ValidatorSet) Validators() []*validator.Validator {
	return set.data.Validators
}

func (set *ValidatorSet) Leavers() []*validator.Validator {
	return set.leavers
}

func (set *ValidatorSet) Join(validator *validator.Validator) error {
	if true == set.Contains(validator.Address()) {
		return fmt.Errorf("This validator currently is in the set: %v", validator.Address())
	}

	/// Welcome to the party!
	set.data.Validators = append(set.data.Validators, validator)
	return nil
}

func (set *ValidatorSet) ForceLeave(validator *validator.Validator) error {
	if false == set.Contains(validator.Address()) {
		return fmt.Errorf("This validator currently is not in the set: %v", validator.Address())
	}

	for i, val := range set.data.Validators {
		if val.Address().EqualsTo(validator.Address()) {
			set.data.Validators = append(set.data.Validators[:i], set.data.Validators[i+1:]...)
			break
		}
	}

	return nil
}

func (set *ValidatorSet) Contains(address crypto.Address) bool {
	for _, v := range set.data.Validators {
		if v.Address().EqualsTo(address) {
			return true
		}
	}

	return false
}

func (set ValidatorSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(set.data)
}

func (set *ValidatorSet) UnmarshalJSON(bs []byte) error {
	err := json.Unmarshal(bs, &set.data)
	if err != nil {
		return err
	}
	set.proxy = _validatorListProxy{}
	return nil
}
