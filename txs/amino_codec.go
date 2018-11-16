package txs

import (
	"fmt"

	"github.com/gallactic/gallactic/txs/tx"
	"github.com/tendermint/go-amino"
)

type aminoCodec struct {
	*amino.Codec
}

func NewAminoCodec() *aminoCodec {
	cdc := amino.NewCodec()
	cdc.RegisterInterface((*tx.Tx)(nil), nil)
	registerTx(cdc, &tx.SendTx{})
	registerTx(cdc, &tx.CallTx{})
	registerTx(cdc, &tx.BondTx{})
	registerTx(cdc, &tx.UnbondTx{})
	registerTx(cdc, &tx.PermissionsTx{})
	registerTx(cdc, &tx.SortitionTx{})
	return &aminoCodec{cdc}
}

func (gwc *aminoCodec) EncodeTx(env *Envelope) ([]byte, error) {
	return gwc.MarshalBinaryLengthPrefixed(env)
}

func (gwc *aminoCodec) DecodeTx(bs []byte) (*Envelope, error) {
	env := new(Envelope)
	err := gwc.UnmarshalBinaryLengthPrefixed(bs, env)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func registerTx(cdc *amino.Codec, tx tx.Tx) {
	cdc.RegisterConcrete(tx, fmt.Sprintf("gallactic/txs/tx/%v", tx.Type()), nil)
}
