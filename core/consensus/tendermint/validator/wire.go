package validator

import (
	"github.com/tendermint/go-amino"
	tmCrypto "github.com/tendermint/go-crypto"
)

var cdc = amino.NewCodec()

func init() {
	tmCrypto.RegisterAmino(cdc)
}
