package crypto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalingEmptyPublicKey(t *testing.T) {
	pb1 := PublicKey{}

	js, err := json.Marshal(pb1)
	assert.NoError(t, err)
	assert.Equal(t, js, []byte("\"\""))
	var pb2 PublicKey
	err = json.Unmarshal(js, &pb2)
	assert.Error(t, err)
	assert.Equal(t, pb1, pb2)

	bs, err := pb1.MarshalAmino()
	assert.NoError(t, err)
	assert.Equal(t, bs, []byte(nil))
	var pb3 PublicKey
	err = json.Unmarshal(bs, &pb3)
	assert.Error(t, err)
	assert.Equal(t, pb1, pb3)
}

func TestMarshalingPublicKey(t *testing.T) {
	pv1 := GeneratePrivateKey(nil)
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

func TestGenerateAddress(t *testing.T) {
	pv, _ := PrivateKeyFromString("85BB7D2E1856C281190FA174E7478F596BAFF265733C7AE6BE87E0DE10E57F3356D2CE5823E4BF1D9621812DE9AFD65DE5786C6096D8C08B4B30C219D8AFC3EF")
	pb1 := pv.PublicKey()
	pb2, _ := PublicKeyFromString("56D2CE5823E4BF1D9621812DE9AFD65DE5786C6096D8C08B4B30C219D8AFC3EF")
	assert.Equal(t, pb1, pb2)
	ac := pb1.AccountAddress()
	va := pb1.ValidatorAddress()
	assert.Equal(t, ac.String(), "ac8KfZqAKYayEWsc6vuwfLu5GDBaCUvoH8B")
	assert.Equal(t, va.String(), "vaB3dLM1UwnarCJsRNLYtwkRRay4zZovj2M")
}
