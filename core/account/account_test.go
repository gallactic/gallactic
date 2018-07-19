package account

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshaling(t *testing.T) {
	acc1 := NewAccountFromSecret("Secret")
	acc1.AddToBalance(999999999999999999)
	acc1.SetPermissions(0x77)
	acc1.IncSequence()
	acc1.SetStorageRoot([]byte{1, 2, 3, 4, 5})
	acc1.SetCode([]byte{60, 23, 45})

	/// test amino encoding/decoding
	bs, err := acc1.Encode()
	require.NoError(t, err)
	acc2 := new(Account)
	err = acc2.Decode(bs)
	require.NoError(t, err)
	assert.Equal(t, acc1, acc2)

	acc3, err := AccountFromBytes(bs)
	require.NoError(t, err)
	assert.Equal(t, acc2, acc3)

	/// test json marshaing/unmarshaling
	js, err := json.Marshal(acc1)
	require.NoError(t, err)
	fmt.Println(string(js))
	acc4 := new(Account)
	require.NoError(t, json.Unmarshal(js, acc4))

	assert.Equal(t, acc3, acc4)

	/// should fail
	acc5, err := AccountFromBytes([]byte("asdfghjkl"))
	require.Error(t, err)
	assert.Nil(t, acc5)

}
