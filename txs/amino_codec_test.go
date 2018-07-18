package txs

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/txs/tx"
	"github.com/gallactic/gallactic/core/validator"
	"encoding/json"
	"fmt"
)

func TestAminoEncodeTxDecodeTx(t *testing.T) {
	codec := NewAminoCodec()
	inputAddress,err := crypto.AddressFromString("acHx3dYGX9pB7xPFZA58ZMcN4kYEooJMVds");
	assert.NoError(t, err, "Invalid input address")

	outputAddress, err := crypto.AddressFromString("acTqSGVw94xP1myXrnCm3rBWgzcJ5uEbB1f");
	assert.NoError(t, err, "Invalid output address")



	amount := uint64(2)
	sequence := uint64(3)
	tx, err := tx.NewSendTx(inputAddress, outputAddress, sequence, amount, 0)
	require.NoError(t, err)

	txEnv := NewEnvelop("test-chain-id", tx)
	txBytes, err := codec.EncodeTx(txEnv)
	assert.NoError(t, err)

	txEnvOut, err := codec.DecodeTx(txBytes)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, txEnv, txEnvOut)

	signer, err := validator.NewPrivateValidator("","")
	assert.NoError(t, err, "Can not create Signer")
	assert.NoError(t,txEnv.Sign(signer))

	txBytes, err = codec.EncodeTx(txEnv)
	assert.NoError(t, err)

	txEnvOut, err = codec.DecodeTx(txBytes)
	assert.NoError(t, err, "DecodeTx error")

	assert.Equal(t, txEnv, txEnvOut)

}

func TestJsonEncodeTxDecodeTx(t *testing.T) {

	inputAddress,err := crypto.AddressFromString("acS1KXQMBNHNzKRioXFdnVFcPrSRdSPU8AA");
	assert.NoError(t, err, "Invalid input address")

	outputAddress, err := crypto.AddressFromString("acTqSGVw94xP1myXrnCm3rBWgzcJ5uEbB1f");
	assert.NoError(t, err, "Invalid output address")

	amount := uint64(2)
	sequence := uint64(3)
	tx, err := tx.NewSendTx(inputAddress, outputAddress, sequence, amount, 0)
	require.NoError(t, err)

	txEnv := NewEnvelop("test-chain-id", tx)
	txBytes, err := json.Marshal(txEnv)
	assert.NoError(t, err)
	txOut := new(Envelope)
	err = json.Unmarshal(txBytes,txOut)
	assert.NoError(t,err)
	assert.Equal(t,txEnv, txOut)


	signer, err := validator.NewPrivateValidator("","")
	assert.NoError(t, err, "Can not create Signer")
	assert.NoError(t,txEnv.Sign(signer))

	txBytes, err = json.Marshal(txEnv)
	assert.NoError(t, err)

	txEnvSigned := new(Envelope)
	err = json.Unmarshal(txBytes , txEnvSigned)
	assert.NoError(t, err, "DecodeTx error")

	assert.Equal(t, txEnv, txEnvSigned)

	fmt.Println(string(txBytes))

}
/*
func TestAminoEncodeTxDecodeTx_CallTx(t *testing.T) {
	codec := NewAminoCodec()
	inputAccount := acm.GeneratePrivateAccountFromSecret("fooo")
	amount := uint64(2)
	sequence := uint64(3)
	tx, err := payload.NewCallTx(inputAccount.Address(), crypto.Address{}, sequence, []byte("code"), 21000, amount, 10)
	require.NoError(t, err)
	txEnv := Enclose(chainID, tx)
	require.NoError(t, txEnv.Sign(inputAccount))
	txBytes, err := codec.EncodeTx(txEnv)
	if err != nil {
		t.Fatal(err)
	}
	txEnvOut, err := codec.DecodeTx(txBytes)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, txEnv, txEnvOut)
}

func TestAminoTxEnvelope(t *testing.T) {
	codec := NewAminoCodec()
	privAccFrom := acm.GeneratePrivateAccountFromSecret("foo")
	privAccTo := acm.GeneratePrivateAccountFromSecret("bar")
	toAddress := privAccTo.Address()
	tx, _ := payload.NewCallTx(privAccFrom.Address(), toAddress, 343, []byte{3, 4, 5, 5}, 21000, 12, 3)
	txEnv := Enclose("testChain", tx)
	err := txEnv.Sign(privAccFrom)
	require.NoError(t, err)

	bs, err := codec.EncodeTx(txEnv)
	require.NoError(t, err)
	txEnvOut, err := codec.DecodeTx(bs)
	require.NoError(t, err)
	assert.Equal(t, txEnv, txEnvOut)
}
*/