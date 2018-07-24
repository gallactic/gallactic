package tests

import (
	"testing"

	"github.com/gallactic/gallactic/core/blockchain"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func setupBlockchain(m *testing.M) {
	tDB = dbm.NewMemDB()
	tBC, _ = blockchain.LoadOrNewBlockchain(tDB, tGenesis, tLogger)
	tState = tBC.State()
}
