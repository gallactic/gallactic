package crypto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalingPublicKey(t *testing.T) {
	pv1, _ := GeneratePrivateKey(nil)
	pb1 := pv1.PublicKey()

	js, err := json.Marshal(&pb1)
	assert.NoError(t, err)

	var pb2 PublicKey
	assert.NoError(t, json.Unmarshal(js, &pb2))
	require.Equal(t, pb1, pb2)

	bs, err := pb1.MarshalAmino()
	assert.NoError(t, err)

	var pb3 PublicKey
	assert.NoError(t, pb3.UnmarshalAmino(bs))

	require.Equal(t, pb2, pb3)

}
