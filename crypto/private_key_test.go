package crypto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalingPrivateKey(t *testing.T) {
	pv1, err := GeneratePrivateKey(nil)
	assert.NoError(t, err)

	js, err := json.Marshal(&pv1)
	assert.NoError(t, err)

	var pv2 PrivateKey
	assert.NoError(t, json.Unmarshal(js, &pv2))
	require.Equal(t, pv1, pv2)

	bs, err := pv1.MarshalAmino()
	assert.NoError(t, err)

	var pv3 PrivateKey
	assert.NoError(t, pv3.UnmarshalAmino(bs))

	require.Equal(t, pv2, pv3)

}
