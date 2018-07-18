package txs

/*
func TestJSONEncodeTxDecodeTx(t *testing.T) {
	codec := NewJSONCodec()
	inputAddress := crypto.Address{1, 2, 3, 4, 5}
	outputAddress := crypto.Address{5, 4, 3, 2, 1}
	amount := uint64(2)
	sequence := uint64(3)
	tx, _ := payload.NewSendTx(inputAddress, outputAddress, sequence, amount, 0)
	txEnv := Enclose(chainID, tx)
	txBytes, err := codec.EncodeTx(txEnv)
	if err != nil {
		t.Fatal(err)
	}
	txEnvOut, err := codec.DecodeTx(txBytes)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, txEnv, txEnvOut)
}

func TestJSONEncodeTxDecodeTx_CallTx(t *testing.T) {
	codec := NewJSONCodec()
	inputAccount := account.GeneratePrivateAccountFromSecret("fooo")
	amount := uint64(2)
	sequence := uint64(3)
	tx, _ := payload.NewCallTx(inputAccount.Address(), crypto.Address{}, sequence, []byte("code"), 21000, amount, 2)
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

func TestJSONEncodeTxDecodeTx_CallTxNoData(t *testing.T) {
	codec := NewJSONCodec()
	inputAccount := account.GeneratePrivateAccountFromSecret("fooo")
	amount := uint64(2)
	sequence := uint64(3)
	tx, _ := payload.NewCallTx(inputAccount.Address(), crypto.Address{}, sequence, nil, 21000, amount, 2)
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
*/
