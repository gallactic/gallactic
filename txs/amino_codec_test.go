package txs

/*
func TestAminoEncodeTxDecodeTx(t *testing.T) {
	codec := NewAminoCodec()
	inputAddress := crypto.Address{1, 2, 3, 4, 5}
	outputAddress := crypto.Address{5, 4, 3, 2, 1}
	amount := uint64(2)
	sequence := uint64(3)
	tx, err := payload.NewSendTx(inputAddress, outputAddress, seq, amt, 0)
	require.NoError(t, err)
	txEnv := Enclose(chainID, tx)
	txBytes, err := codec.EncodeTx(txEnv)
	if err != nil {
		t.Fatal(err)
	}
	txEnvOut, err := codec.DecodeTx(txBytes)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, txEnv, txEnvOut)
}

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
