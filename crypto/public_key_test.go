package crypto

import (
	"encoding/hex"
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
	assert.NoError(t, err) /// No error
	assert.Equal(t, pb1, pb2)

	bs, err := pb1.MarshalAmino()
	assert.NoError(t, err)
	assert.Equal(t, bs, []byte(nil))
	var pb3 PublicKey
	err = pb3.UnmarshalAmino(bs)
	assert.NoError(t, err) /// No error
	assert.Equal(t, pb1, pb3)
}

func TestMarshalingPublicKey(t *testing.T) {
	pb1, _ := GenerateKey(nil)
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
	bs1, _ := hex.DecodeString("85BB7D2E1856C281190FA174E7478F596BAFF265733C7AE6BE87E0DE10E57F3356D2CE5823E4BF1D9621812DE9AFD65DE5786C6096D8C08B4B30C219D8AFC3EF")
	bs2, _ := hex.DecodeString("56D2CE5823E4BF1D9621812DE9AFD65DE5786C6096D8C08B4B30C219D8AFC3EF")
	txt1 := "skZfztcE4vkJLYNQ3TcvAkgH24TV1hQfuojiwReVto9JknsoWNZPJVmd6agFiCyGx1px45HJjgRQvRNRrc4oeqZgaPXhQHM"
	txt2 := "pjjHzwbbW5gVGNsc8u3vyX9AxBB7jqXcyV5XavPFesUJiWpaai8"
	pv1, _ := PrivateKeyFromRawBytes(bs1)
	pv2, _ := PrivateKeyFromString(txt1)

	pb1 := pv1.PublicKey()
	pb2, _ := PublicKeyFromRawBytes(bs2)
	pb3, _ := PublicKeyFromString(txt2)
	assert.Equal(t, pv1, pv2)
	assert.Equal(t, pb1, pb2)
	assert.Equal(t, pb1, pb3)
	assert.Equal(t, pv1.String(), txt1)
	assert.Equal(t, pb1.String(), txt2)
	assert.Equal(t, pv1.RawBytes(), bs1)
	assert.Equal(t, pb1.RawBytes(), bs2)
	ac := pb1.AccountAddress()
	va := pb1.ValidatorAddress()
	assert.Equal(t, ac.String(), "ac8KfZqAKYayEWsc6vuwfLu5GDBaCUvoH8B")
	assert.Equal(t, va.String(), "vaB3dLM1UwnarCJsRNLYtwkRRay4zZovj2M")
}


