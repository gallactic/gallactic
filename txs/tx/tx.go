package tx

import (
	"encoding/json"
	"fmt"

	amino "github.com/tendermint/go-amino"
)

/*
Payload (Transaction) is an atomic operation on the ledger state.

Account Txs:
 - SendTx         Send coins to address
 - CallTx         Send a msg to a contract that runs in the vm

Validation Txs:
 - BondTx         New validator posts a bond
 - UnbondTx       Validator leaves

Admin Txs:
 - PermissionsTx
*/

type Type int8

// Types of Payload implementations
const (
	TypeUnknown = Type(0x00)
	// Account transactions
	TypeSend = Type(0x01)
	TypeCall = Type(0x02)

	// Validation transactions
	TypeBond      = Type(0x11)
	TypeUnbond    = Type(0x12)
	TypeSortition = Type(0x13)

	// Admin transactions
	TypePermissions = Type(0x21)
)

var nameFromType = map[Type]string{
	TypeUnknown:     "UnknownTx",
	TypeSend:        "SendTx",
	TypeCall:        "CallTx",
	TypeBond:        "BondTx",
	TypeUnbond:      "UnbondTx",
	TypeSortition:   "SortitionTx",
	TypePermissions: "PermissionsTx",
}

var typeFromName = make(map[string]Type)
var vc = amino.NewCodec()

func init() {
	for t, n := range nameFromType {
		typeFromName[n] = t
	}
}

func TxTypeFromString(name string) Type {
	return typeFromName[name]
}

func (typ Type) String() string {
	name, ok := nameFromType[typ]
	if ok {
		return name
	}
	return "UnknownTx"
}

func (typ Type) MarshalText() ([]byte, error) {
	return []byte(typ.String()), nil
}

func (typ *Type) UnmarshalText(data []byte) error {
	*typ = TxTypeFromString(string(data))
	return nil
}

type Tx interface {
	Inputs() []TxInput
	Outputs() []TxOutput
	Type() Type
}

func InAmount(tx Tx) uint64 {
	inAmount := uint64(0)
	for _, input := range tx.Inputs() {
		inAmount += input.Amount
	}
	return inAmount
}

func OutAmount(tx Tx) uint64 {
	outAmount := uint64(0)
	for _, output := range tx.Outputs() {
		outAmount += output.Amount
	}
	return outAmount
}

func Fee(tx Tx) uint64 {
	return InAmount(tx) - OutAmount(tx)
}

func String(tx Tx) string {
	bytes, err := json.Marshal(tx)
	if err != nil {
		return "Marshaling error"
	}
	return fmt.Sprintf("%s%s", tx.Type(), string(bytes))
}

func New(txType Type) Tx {
	switch txType {
	case TypeSend:
		return &SendTx{}
	case TypeCall:
		return &CallTx{}
	case TypeBond:
		return &BondTx{}
	case TypeUnbond:
		return &UnbondTx{}
	case TypeSortition:
		return &SortitionTx{}
	case TypePermissions:
		return &PermissionsTx{}
	}
	return nil
}
