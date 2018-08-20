package validator

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshaling(t *testing.T) {
	pb, _ := crypto.GenerateKey(nil)
	val1, err := NewValidator(pb, 1)
	assert.NoError(t, err)
	val1.AddToStake(999999999999999999)
	val1.IncSequence()

	/// test amino encoding/decoding
	bs, err := val1.Encode()
	require.NoError(t, err)
	val2 := new(Validator)
	err = val2.Decode(bs)
	require.NoError(t, err)
	assert.Equal(t, val1, val2)

	val3, err := ValidatorFromBytes(bs)
	require.NoError(t, err)
	assert.Equal(t, val2, val3)

	/// test json marshaing/unmarshaling
	js, err := json.Marshal(val1)
	require.NoError(t, err)
	fmt.Println(string(js))
	val4 := new(Validator)
	require.NoError(t, json.Unmarshal(js, val4))

	assert.Equal(t, val3, val4)

	/// should fail
	val5, err := ValidatorFromBytes([]byte("asdfghjkl"))
	require.Error(t, err)
	assert.Nil(t, val5)
}
