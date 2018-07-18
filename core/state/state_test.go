package state

import (
	"fmt"
	"testing"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func loadState(t *testing.T, db dbm.DB, hash []byte) *State {
	s, err := LoadState(db, hash)
	require.NoError(t, err)
	require.NotNil(t, s)

	return s
}

func saveState(t *testing.T, state *State) []byte {
	hash, err := state.SaveState()
	require.NoError(t, err)

	fmt.Printf("hash:%v\n", hash)

	return hash
}

func updateAccount(t *testing.T, state *State, account *account.Account) {
	err := state.AccountPool.UpdateAccount(account)
	require.NoError(t, err)
}

func getAccount(t *testing.T, state *State, addr crypto.Address) *account.Account {
	account := state.AccountPool.GetAccount(address)
	return account
}

func TestState_LoadingWrongHash(t *testing.T) {
	db := dbm.NewMemDB()
	s0, err := LoadState(db, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0})
	require.Error(t, err)
	require.Nil(t, s0)
}

func TestState_Loading(t *testing.T) {
	db := dbm.NewMemDB()
	state := NewState(db)

	foo := account.NewAccountFromSecret("Foo")
	bar := account.NewAccountFromSecret("Bar")

	foo.AddToBalance(1)
	bar.AddToBalance(1)

	updateAccount(t, state, foo)
	updateAccount(t, state, bar)

	hash1 := saveState(t, state)
	hash2 := saveState(t, state)
	require.Equal(t, hash1, hash2)

	foo.AddToBalance(1)
	updateAccount(t, state, foo)
	hash3 := saveState(t, state)

	require.NotEqual(t, hash1, hash3)

	/// --- Immutable saved state
	state1 := loadState(t, db, hash1)
	foo.AddToBalance(1)
	_, err := state1.SaveState()
	require.Error(t, err)
	/// ---

	state2 := loadState(t, db, hash2)
	state3 := loadState(t, db, hash3)

	foo2 := getAccount(t, state2, foo.Address())
	foo3 := getAccount(t, state3, foo.Address())

	require.Equal(t, uint64(1), foo2.Balance())
	require.Equal(t, uint64(2), foo3.Balance())
}

func TestState_Loading2(t *testing.T) {
	db := dbm.NewMemDB()
	state := NewState(db)

	foo := account.NewAccountFromSecret("Foo")
	bar := account.NewAccountFromSecret("Bar")

	foo.AddToBalance(1)
	bar.AddToBalance(1)

	updateAccount(t, state, foo)
	hash1 := saveState(t, state)

	updateAccount(t, state, bar)
	hash2 := saveState(t, state)

	require.NotEqual(t, hash1, hash2)

	foo2 := getAccount(t, state, foo.Address())
	bar2 := getAccount(t, state, bar.Address())

	require.Equal(t, uint64(1), foo2.Balance())
	require.Equal(t, uint64(1), bar2.Balance())

	state2 := loadState(t, db, hash2)

	foo3 := getAccount(t, state2, foo.Address())
	bar3 := getAccount(t, state2, bar.Address())

	require.Equal(t, uint64(1), foo3.Balance())
	require.Equal(t, uint64(1), bar3.Balance())
}

func TestState_UpdateAccount(t *testing.T) {
	state := NewState(dbm.NewMemDB())
	foo1 := account.NewAccountFromSecret("Foo")

	foo1.AddToBalance(1)
	foo1.SetCode([]byte{0x60})
	updateAccount(t, state, foo1)

	foo2 := getAccount(t, state, foo1.Address())
	assert.Equal(t, foo1, foo2)
}

/*
func TestState_Publish(t *testing.T) {
	s := NewState(db.NewMemDB())
	ctx := context.Background()
	evs := []*events.Event{
		mkEvent(100, 0),
		mkEvent(100, 1),
	}
	_, err := s.Update(func(ws Updatable) error {
		for _, ev := range evs {
			require.NoError(t, ws.Publish(ctx, ev, nil))
		}
		return nil
	})
	require.NoError(t, err)
	i := 0
	_, err = s.GetEvents(events.NewKey(100, 0), events.NewKey(100, 0),
		func(ev *events.Event) (stop bool) {
			assert.Equal(t, evs[i], ev)
			i++
			return false
		})
	require.NoError(t, err)
	// non-increasing events
	_, err = s.Update(func(ws Updatable) error {
		require.Error(t, ws.Publish(ctx, mkEvent(100, 0), nil))
		require.Error(t, ws.Publish(ctx, mkEvent(100, 1), nil))
		require.Error(t, ws.Publish(ctx, mkEvent(99, 1324234), nil))
		require.NoError(t, ws.Publish(ctx, mkEvent(100, 2), nil))
		require.NoError(t, ws.Publish(ctx, mkEvent(101, 0), nil))
		return nil
	})
	require.NoError(t, err)
}

func TestProtobufEventSerialisation(t *testing.T) {
	ev := mkEvent(112, 23)
	pbEvent := pbevents.GetExecutionEvent(ev)
	bs, err := proto.Marshal(pbEvent)
	require.NoError(t, err)
	pbEventOut := new(pbevents.ExecutionEvent)
	require.NoError(t, proto.Unmarshal(bs, pbEventOut))
	fmt.Println(pbEventOut)
	assert.Equal(t, asJSON(t, pbEvent), asJSON(t, pbEventOut))
}

func mkEvent(height, index uint64) *events.Event {
	return &events.Event{
		Header: &events.Header{
			Height:  height,
			Index:   index,
			TxHash:  sha3.Sha3([]byte(fmt.Sprintf("txhash%v%v", height, index))),
			EventID: fmt.Sprintf("eventID: %v%v", height, index),
		},
		Tx: &events.EventDataTx{
			Tx: txs.Enclose("foo", &payload.CallTx{}).Tx,
		},
		Log: &events.EventDataLog{
			Address: crypto.Address{byte(height), byte(index)},
			Topics:  []binary.Word256{{1, 2, 3}},
		},
	}
}

func asJSON(t *testing.T, v interface{}) string {
	bs, err := json.Marshal(v)
	require.NoError(t, err)
	return string(bs)
}
*/
