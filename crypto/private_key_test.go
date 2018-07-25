package crypto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalingEmptyPrivateKey(t *testing.T) {
	pv1 := PrivateKey{}

	js, err := json.Marshal(pv1)
	assert.NoError(t, err)
	assert.Equal(t, js, []byte("\"\""))
	var pv2 PrivateKey
	err = json.Unmarshal(js, &pv2)
	assert.Error(t, err) /// return error
	assert.Equal(t, pv1, pv2)

	bs, err := pv1.MarshalAmino()
	assert.NoError(t, err)
	assert.Equal(t, bs, []byte(nil))
	var pv3 PrivateKey
	err = pv3.UnmarshalAmino(bs)
	assert.Error(t, err) /// return error
	assert.Equal(t, pv1, pv3)
}

func TestMarshalingPrivateKey(t *testing.T) {
	pv1 := GeneratePrivateKey(nil)
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
