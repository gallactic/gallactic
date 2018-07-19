package tx

import (
	"encoding/json"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/crypto"
)

type PermissionsTx struct {
	data permissionsData
}
type permissionsData struct {
	Modifier    TxInput             `json:"modifier"`
	Modified    TxOutput            `json:"modified"`
	Permissions account.Permissions `json:"permissions"`
	Set         bool                `json:"set"`
}

func NewPermissionsTx(modifier, modified crypto.Address, perm account.Permissions, set bool, seq, fee uint64) (*PermissionsTx, error) {
	return &PermissionsTx{
		data: permissionsData{
			Modifier: TxInput{
				Address:  modifier,
				Sequence: seq,
				Amount:   fee,
			},
			Modified: TxOutput{
				Address: modified,
				Amount:  0,
			},

			Permissions: perm,
			Set:         set,
		},
	}, nil
}

func (tx *PermissionsTx) Type() Type                       { return TypePermissions }
func (tx *PermissionsTx) Modifier() crypto.Address         { return tx.data.Modifier.Address }
func (tx *PermissionsTx) Modified() crypto.Address         { return tx.data.Modified.Address }
func (tx *PermissionsTx) Fee() uint64                      { return tx.data.Modifier.Amount }
func (tx *PermissionsTx) Permissions() account.Permissions { return tx.data.Permissions }
func (tx *PermissionsTx) Set() bool                        { return tx.data.Set }

func (tx *PermissionsTx) Inputs() []TxInput {
	return []TxInput{tx.data.Modifier}
}

func (tx *PermissionsTx) Outputs() []TxOutput {
	return []TxOutput{tx.data.Modified}
}

/// ----------
/// MARSHALING

func (tx *PermissionsTx) MarshalAmino() ([]byte, error) {
	return cdc.MarshalBinary(tx.data)
}

func (tx *PermissionsTx) UnmarshalAmino(bs []byte) error {
	return cdc.UnmarshalBinary(bs, &tx.data)
}

func (tx PermissionsTx) MarshalJSON() ([]byte, error) {
	return json.Marshal(tx.data)
}

func (tx *PermissionsTx) UnmarshalJSON(bs []byte) error {
	err := json.Unmarshal(bs, &tx.data)
	if err != nil {
		return err
	}
	return nil
}
