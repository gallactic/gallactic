package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalingSignature(t *testing.T) {
	privKey, _ := GeneratePrivateKey(nil)
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
