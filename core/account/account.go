package account

import (
	"encoding/json"
	"fmt"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	amino "github.com/tendermint/go-amino"
)

// Account structure
type Account struct {
	data accountData
}

type accountData struct {
	Address     crypto.Address `json:"address"`
	Sequence    uint64         `json:"sequence"`
	Balance     uint64         `json:"balance"`
	Code        []byte         `json:"code"`
	Permissions Permissions    `json:"permissions"`
}

///---- Constructors
func NewAccount(addr crypto.Address) (*Account, error) {
	if err := addr.EnsureValid(); err != nil {
		return nil, err
	}

	if !addr.IsAccountAddress() {
		return nil, e.Errorf(e.ErrInvalidAddress, "This is not a valid acccount address: %s", addr.String())
	}

	return &Account{
		data: accountData{
			Address: addr,
		},
	}, nil
}

/// For tests
func NewAccountFromSecret(secret string) *Account {
	acc, _ := NewAccount(crypto.PrivateKeyFromSecret(secret).PublicKey().AccountAddress())
	return acc
}

func AccountFromBytes(bs []byte) (*Account, error) {
	var acc Account
	if err := acc.Decode(bs); err != nil {
		return nil, err
	}

	return &acc, nil
}

func (acc Account) Address() crypto.Address  { return acc.data.Address }
func (acc Account) Balance() uint64          { return acc.data.Balance }
func (acc Account) Sequence() uint64         { return acc.data.Sequence }
func (acc Account) Code() []byte             { return acc.data.Code }
func (acc Account) Permissions() Permissions { return acc.data.Permissions }

func (acc Account) HasPermissions(perm Permissions) bool {
	return acc.data.Permissions.IsSet(perm)
}

func (acc *Account) SetBalance(bal uint64) error {
	acc.data.Balance = bal
	return nil
}

func (acc *Account) SubtractFromBalance(amt uint64) error {
	if amt > acc.Balance() {
		return e.Errorf(e.ErrInsufficientFunds, "Attempt to subtract %v from the balance of %s", amt, acc.Address())
	}
	acc.data.Balance -= amt
	return nil
}

func (acc *Account) AddToBalance(amt uint64) error {
	acc.data.Balance += amt
	return nil
}

func (acc *Account) SetCode(code []byte) error {
	acc.data.Code = code
	return nil
}

func (acc *Account) SetSequence(seq uint64) {
	acc.data.Sequence = seq
}

func (acc *Account) IncSequence() {
	acc.data.Sequence++
}

func (acc *Account) SetPermissions(perm Permissions) error {
	acc.data.Permissions.Set(perm)
	return nil
}

func (acc *Account) UnsetPermissions(perm Permissions) error {
	acc.data.Permissions.Unset(perm)
	return nil
}

///---- Serialization methods
var ac = amino.NewCodec()

func (acc Account) Encode() ([]byte, error) {
	return ac.MarshalBinary(&acc.data)
}

func (acc *Account) Decode(bs []byte) error {
	err := ac.UnmarshalBinary(bs, &acc.data)
	if err != nil {
		return err
	}
	return nil

}

func (acc Account) MarshalJSON() ([]byte, error) {
	return json.Marshal(acc.data)
}

func (acc *Account) UnmarshalJSON(bs []byte) error {
	err := json.Unmarshal(bs, &acc.data)
	if err != nil {
		return err
	}
	return nil
}

func AccountFromJSON(bs []byte) (*Account, error) {
	var acc Account
	if err := acc.UnmarshalJSON(bs); err != nil {
		return nil, err
	}
	return &acc, nil
}

func (acc Account) String() string {
	b, _ := acc.MarshalJSON()
	return fmt.Sprintf("Account%s", string(b))
}
