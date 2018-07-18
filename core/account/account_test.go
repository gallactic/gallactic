package account

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountDecode(t *testing.T) {
	acc1 := NewAccountFromSecret("Super Semi Secret")
	acc1.AddToBalance(999999999999999999)
	acc1.SetPermissions(0x77)
	acc1.IncSequence()
	acc1.SetStorageRoot([]byte{1, 2, 3, 4, 5})
	acc1.SetCode([]byte{60, 23, 45})
	fmt.Println(acc1.Address().String())
	bytes, err := acc1.Encode()
	require.NoError(t, err)
	var acc2 Account
	err = acc2.Decode(bytes)
	require.NoError(t, err)
	assert.Equal(t, *acc1, acc2)

	acc3, err := AccountFromBytes([]byte("asdfghjkl"))
	require.Error(t, err)
	assert.Nil(t, acc3)
}

func TestAccountMarshal(t *testing.T) {
	acc1 := NewAccountFromSecret("Secret")
	acc1.SetPermissions(0x77)
	acc1.AddToBalance(100)
	acc1.IncSequence()
	acc1.SetStorageRoot([]byte{1, 2, 3, 4, 5})
	acc1.SetCode([]byte{60, 23, 45})

	bs, err1 := json.Marshal(acc1)
	require.NoError(t, err1)
	fmt.Println(string(bs))

	var acc2 Account
	err2 := json.Unmarshal(bs, &acc2)
	require.NoError(t, err2)

	assert.Equal(t, *acc1, acc2)
}
