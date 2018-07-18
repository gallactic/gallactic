package state

import (
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
)

func (st *State) GetValidator(addr crypto.Address) *validator.Validator {
	st.Lock()
	defer st.Unlock()

	_, bytes := st.tree.Get(validatorKey(addr))
	if bytes == nil {
		return nil
	}
	val, err := validator.ValidatorFromBytes(bytes)
	if err != nil {
		panic("Unable to decode encoded validator")
	}

	return val
}

func (st *State) UpdateValidator(val *validator.Validator) error {
	st.Lock()
	defer st.Unlock()

	bs, err := val.Encode()
	if err != nil {
		return err
	}

	st.tree.Set(accountKey(val.Address()), bs)

	return nil
}

func (st *State) ValidatorCount() int {
	count := 0
	st.IterateValidators(func(val *validator.Validator) (stop bool) {
		count++
		return false
	})
	return count
}

func (st *State) IterateValidators(consumer func(*validator.Validator) (stop bool)) (stopped bool, err error) {
	return st.tree.IterateRange(validatorStart, validatorEnd, true, func(key []byte, bs []byte) (stop bool) {
		validator, err := validator.ValidatorFromBytes(bs)
		if err != nil {
			return true
		}
		return consumer(validator)
	}), nil
}
