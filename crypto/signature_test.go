package crypto

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalingEmptySignature(t *testing.T) {
	sig1 := Signature{}

	js, err := json.Marshal(sig1)
	assert.NoError(t, err)
	assert.Equal(t, js, []byte("\"\""))
	var sig2 Signature
	err = json.Unmarshal(js, &sig2)
	assert.NoError(t, err) /// No error
	assert.Equal(t, sig1, sig2)

	bs, err := sig1.MarshalAmino()
	assert.NoError(t, err)
	assert.Equal(t, bs, []byte(nil))
	var sig3 Signature
	err = sig3.UnmarshalAmino(bs)
	assert.NoError(t, err) /// No error
	assert.Equal(t, sig1, sig3)
}

func TestMarshalingSignature(t *testing.T) {
	privKey := GeneratePrivateKey(nil)
	sig1, err := privKey.Sign([]byte("Test message"))
	require.NoError(t, err)

	bs, err := sig1.MarshalText()
	fmt.Println(string(bs))
	require.NoError(t, err)

	var sig2 Signature
	err = sig2.UnmarshalText(bs)
	require.NoError(t, err)
	require.Equal(t, sig1, sig2)

	bs, err = sig2.MarshalAmino()
	assert.NoError(t, err)

	var sig3 Signature
	assert.NoError(t, sig3.UnmarshalAmino(bs))

	require.Equal(t, sig2, sig3)
}
