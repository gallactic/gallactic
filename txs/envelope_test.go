package txs

/*
func TestJSONEncodeTxDecodeTx(t *testing.T) {
	codec := NewJSONCodec()
	inputAddress := crypto.Address{1, 2, 3, 4, 5}
	outputAddress := crypto.Address{5, 4, 3, 2, 1}
	amount := uint64(2)
	sequence := uint64(3)
	tx, _ := payload.NewSendTx(inputAddress, outputAddress, seq, amt, 0)
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

/*
func makePrivateAccount(str string) acm.PrivateAccount {
	acc := acm.GeneratePrivateAccountFromSecret(str)
	privateAccounts[acc.Address()] = acc
	return acc
}

func TestSendTxSignable(t *testing.T) {
	sendTx := &payload.SendTx{
		Inputs: []*payload.TxInput{
			{
				Address:  makePrivateAccount("input1").Address(),
				Amount:   12345,
				Sequence: 67890,
			},
			{
				Address:  makePrivateAccount("input2").Address(),
				Amount:   111,
				Sequence: 222,
			},
		},
		Outputs: []*payload.TxOutput{
			{
				Address: makePrivateAccount("output1").Address(),
				Amount:  333,
			},
			{
				Address: makePrivateAccount("output2").Address(),
				Amount:  444,
			},
		},
	}
	testTxMarshalJSON(t, sendTx)
	testTxSignVerify(t, sendTx)
}

func TestCallTxSignable(t *testing.T) {
	toAddress := makePrivateAccount("contract1").Address()
	callTx := &payload.CallTx{
		Input: &payload.TxInput{
			Address:  makePrivateAccount("input1").Address(),
			Amount:   12345,
			Sequence: 67890,
		},
		Address:  &toAddress,
		GasLimit: 111,
		Fee:      222,
		Data:     []byte("data1"),
	}
	testTxMarshalJSON(t, callTx)
	testTxSignVerify(t, callTx)
}

func TestNameTxSignable(t *testing.T) {
	nameTx := &payload.NameTx{
		Input: &payload.TxInput{
			Address:  makePrivateAccount("input1").Address(),
			Amount:   12345,
			Sequence: 250,
		},
		Name: "google.com",
		Data: "secretly.not.google.com",
		Fee:  1000,
	}
	testTxMarshalJSON(t, nameTx)
	testTxSignVerify(t, nameTx)
}

func TestBondTxSignable(t *testing.T) {
	bondTx := &payload.BondTx{
		Inputs: []*payload.TxInput{
			{
				Address:  makePrivateAccount("input1").Address(),
				Amount:   12345,
				Sequence: 67890,
			},
			{
				Address:  makePrivateAccount("input2").Address(),
				Amount:   111,
				Sequence: 222,
			},
		},
		UnbondTo: []*payload.TxOutput{
			{
				Address: makePrivateAccount("output1").Address(),
				Amount:  333,
			},
			{
				Address: makePrivateAccount("output2").Address(),
				Amount:  444,
			},
		},
	}
	testTxMarshalJSON(t, bondTx)
	testTxSignVerify(t, bondTx)
}

func TestUnbondTxSignable(t *testing.T) {
	unbondTx := &payload.UnbondTx{
		Input: &payload.TxInput{
			Address: makePrivateAccount("fooo1").Address(),
		},
		Address: makePrivateAccount("address1").Address(),
		Height:  111,
	}
	testTxMarshalJSON(t, unbondTx)
	testTxSignVerify(t, unbondTx)
}

func TestPermissionsTxSignable(t *testing.T) {
	permsTx := &payload.PermissionsTx{
		Modifier: payload.TxInput{
			Address:  makePrivateAccount("input1").Address(),
			Amount:   12345,
			Sequence: 250,
		},
		Modified:    makePrivateAccount("address1").Address(),
		Permissions: 1,
		Set:         true,
	}

	testTxMarshalJSON(t, permsTx)
	testTxSignVerify(t, permsTx)
}
func TestTxWrapper_MarshalJSON(t *testing.T) {
	toAddress := makePrivateAccount("contract1").Address()
	callTx := &payload.CallTx{
		Input: &payload.TxInput{
			Address:  makePrivateAccount("input1").Address(),
			Amount:   12345,
			Sequence: 67890,
		},
		Address:  &toAddress,
		GasLimit: 111,
		Fee:      222,
		Data:     []byte("data1"),
	}
	testTxMarshalJSON(t, callTx)
	testTxSignVerify(t, callTx)
}

func TestNewPermissionsTxWithSequence(t *testing.T) {
	privateKey1 := makePrivateAccount("Shhh...")
	privateKey2 := makePrivateAccount("Chhh...")

	permTx, _ := payload.NewPermissionsTx(privateKey1.PublicKey().Address(), privateKey2.PublicKey().Address(), permission.Call, true, 1, 100)
	testTxMarshalJSON(t, permTx)
}

func testTxMarshalJSON(t *testing.T, tx payload.Payload) {
	txw := &Tx{Payload: tx}
	bs, err := json.Marshal(txw)
	require.NoError(t, err)
	txwOut := new(Tx)
	err = json.Unmarshal(bs, txwOut)
	require.NoError(t, err)
	bsOut, err := json.Marshal(txwOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}

func testTxSignVerify(t *testing.T, tx payload.Payload) {
	inputs := tx.GetInputs()
	var signers []acm.AddressableSigner
	for _, in := range inputs {
		signers = append(signers, privateAccounts[in.Address])
	}
	txEnv := Enclose(chainID, tx)
	require.NoError(t, txEnv.Sign(signers...), "Error signing tx: %s", debug.Stack())
	require.NoError(t, txEnv.Verify(nil), "Error verifying tx: %s", debug.Stack())
}

*/
/*

func TestSignature(t *testing.T) {
	privKey1 := types.PrivateKeyFromSecret("secret1")
	privKey2 := types.PrivateKeyFromSecret("secret2")
	privKey3 := types.PrivateKeyFromSecret("secret3")

	pubKey1 := privKey1.PublicKey()
	pubKey2 := privKey2.PublicKey()
	pubKey3 := privKey3.PublicKey()

	tx, _ := payload.EmptySendTx()
	tx.AddReceiver(crypto.Address{1}, 1)
	tx.AddSender(pubKey1.AccountAddress(), 1, 1)
	tx.AddSender(pubKey2.AccountAddress(), 1, 1)
	tx.AddSender(pubKey3.AccountAddress(), 1, 1)

	txEnv := Enclose(chainID, tx)

	privAcc1 := types.ConcretePrivateAccount{PrivateKey: privKey1}.PrivateAccount()
	privAcc2 := types.ConcretePrivateAccount{PrivateKey: privKey2}.PrivateAccount()
	err := txEnv.Sign(privAcc1, privAcc2)
	assert.Error(t, err)

	err = txEnv.Verify()
	require.Error(t, err)

	privAcc3 := acm.ConcretePrivateAccount{PrivateKey: privKey3}.PrivateAccount()
	err = txEnv.Sign(privAcc1, privAcc2, privAcc3)
	require.NoError(t, err)

	err = txEnv.Verify()
	require.NoError(t, err)

	/// TODO: Add more tests here
}

*/
