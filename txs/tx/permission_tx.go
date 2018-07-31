package tx

import (
	"encoding/json"
	"fmt"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
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
func (tx *PermissionsTx) Modifier() TxInput                { return tx.data.Modifier }
func (tx *PermissionsTx) Modified() TxOutput               { return tx.data.Modified }
func (tx *PermissionsTx) Permissions() account.Permissions { return tx.data.Permissions }
func (tx *PermissionsTx) Set() bool                        { return tx.data.Set }

func (tx *PermissionsTx) Signers() []TxInput {
	return []TxInput{tx.data.Modifier}
}

func (tx *PermissionsTx) Amount() uint64 {
	return 0
}

func (tx *PermissionsTx) Fee() uint64 {
	return tx.data.Modifier.Amount
}

func (tx *PermissionsTx) EnsureValid() error {
	/// Just modifying permission, not transferring money
	if tx.data.Modified.Amount != 0 {
		return e.Error(e.ErrInvalidAmount)
	}

	if err := tx.data.Modifier.ensureValid(); err != nil {
		return err
	}

	if err := tx.data.Modified.ensureValid(); err != nil {
		return err
	}

	if tx.data.Modified.Address == crypto.GlobalAddress {
		return fmt.Errorf("You can not change global account's permission")
	}

	return nil
}

/// ----------
/// MARSHALING

func (tx PermissionsTx) MarshalAmino() ([]byte, error) {
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
