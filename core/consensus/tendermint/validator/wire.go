package validator

import (
	"github.com/tendermint/go-amino"
	tmCrypto "github.com/tendermint/tendermint/crypto"
)

var cdc = amino.NewCodec()

func init() {
	tmCrypto.RegisterAmino(cdc)
}
