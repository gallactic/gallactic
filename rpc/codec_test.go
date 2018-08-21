package rpc

import (
	"fmt"
	"testing"

	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeysEncoding(t *testing.T) {
	codec := NewTCodec()
	pb, pv := crypto.GenerateKeyFromSecret("codec test")
	type keyPair struct {
		PrivateKey crypto.PrivateKey
		PublicKey  crypto.PublicKey
	}

	kp := keyPair{
		PrivateKey: pv,
		PublicKey:  pb,
	}
	fmt.Println("Original Key Pair :\n", kp)

	bs, err := codec.EncodeBytes(kp)
	fmt.Println("\nEncoded Key Pair :\n", string(bs))
	require.NoError(t, err)

	kpOut := keyPair{}
	codec.DecodeBytes(&kpOut, bs)
	fmt.Println("\nDecoded Key Pair :\n", kpOut)
	assert.Equal(t, kp, kpOut)
}
