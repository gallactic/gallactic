package validator

import (
	"encoding/json"
	"fmt"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	amino "github.com/tendermint/go-amino"
)

type Validator struct {
	data validatorData
}

type validatorData struct {
	PublicKey     crypto.PublicKey `json:"publicKey"`
	Stake         uint64           `json:"stake"`
	BondingHeight uint64           `json:"bondingHeight"`
	Sequence      uint64           `json:"sequence"`
}

func NewValidator(publicKey crypto.PublicKey, bondingHeight uint64) (*Validator, error) {
	val := &Validator{
		data: validatorData{
			PublicKey:     publicKey,
			BondingHeight: bondingHeight,
			Stake:         0,
			Sequence:      0,
		},
	}
	return val, nil
}

func (val *Validator) Address() crypto.Address     { return val.data.PublicKey.ValidatorAddress() }
func (val *Validator) Stake() uint64               { return val.data.Stake }
func (val *Validator) Sequence() uint64            { return val.data.Sequence }
func (val *Validator) PublicKey() crypto.PublicKey { return val.data.PublicKey }
func (val *Validator) BondingHeight() uint64       { return val.data.BondingHeight }

func (val Validator) Power() int64 {
	// Viva democracy, every person will be treated equally in our blockchain
	return 1
}

func (val Validator) MinimumStakeToUnbond() uint64 {
	//TODO:Mostafa
	return 0
}
func (val *Validator) SubtractFromStake(amt uint64) error {
	if amt > val.Stake() {
		return e.Errorf(e.ErrInsufficientFunds, "Attempt to subtract %v from the balance of %s", amt, val.Address())
	}
	val.data.Stake -= amt
	return nil
}

func (val *Validator) AddToStake(amt uint64) error {
	val.data.Stake += amt
	return nil
}

func (val *Validator) IncSequence() {
	val.data.Sequence++
}

///---- Serialization methods
var cdc = amino.NewCodec()

func (val Validator) Encode() ([]byte, error) {
	return cdc.MarshalBinary(val.data)
}

func (val *Validator) Decode(bs []byte) error {
	err := cdc.UnmarshalBinary(bs, &val.data)
	if err != nil {
		return err
	}
	return nil
}

func ValidatorFromBytes(bs []byte) (*Validator, error) {
	var val Validator
	if err := val.Decode(bs); err != nil {
		return nil, err
	}
	return &val, nil
}

func (val Validator) MarshalJSON() ([]byte, error) {
	return json.Marshal(val.data)
}

func (val *Validator) UnmarshalJSON(bs []byte) error {
	err := json.Unmarshal(bs, &val.data)
	if err != nil {
		return err
	}
	return nil
}

func ValidatorFromJSON(bs []byte) (*Validator, error) {
	var val Validator
	if err := val.UnmarshalJSON(bs); err != nil {
		return nil, err
	}
	return &val, nil
}

func (val Validator) String() string {
	b, _ := val.MarshalJSON()
	return fmt.Sprintf("Validator%s", string(b))
}

func (val *Validator) Unmarshal(bs []byte) error {
	return val.Decode(bs)
}

func (val *Validator) Marshal() ([]byte, error) {
	return val.Encode()
}

func (val *Validator) MarshalTo(data []byte) (int, error) {
	bs, err := val.Encode()
	if err != nil {
		return -1, err
	}
	return copy(data, bs), nil
}

func (val *Validator) Size() int {
	bs, _ := val.Encode()
	return len(bs)
}
