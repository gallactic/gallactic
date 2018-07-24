package tests

import (
	"testing"

	"github.com/gallactic/gallactic/core/blockchain"
	dbm "github.com/tendermint/tmlibs/db"
)

func setupBlockchain(m *testing.M) {
	tDB = dbm.NewMemDB()
	tBC, _ = blockchain.LoadOrNewBlockchain(tDB, tGenesis, tLogger)
	tState = tBC.State()
}
