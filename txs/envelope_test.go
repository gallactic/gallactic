package txs

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/txs/tx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendMarshaling(t *testing.T) {
	_, pv := crypto.GenerateKey(nil)
	signer := crypto.NewAccountSigner(pv)
	sender := signer.Address()
	tx, err := tx.NewSendTx(sender, crypto.GlobalAddress, 1, 100, 200)
	require.NoError(t, err)

	testMarshaling(t, tx, signer)
}

func TestCallMarshaling(t *testing.T) {
	_, pv := crypto.GenerateKey(nil)
	signer := crypto.NewAccountSigner(pv)
	caller := signer.Address()
	tx, err := tx.NewCallTx(caller, crypto.Address{}, 1, []byte{1, 2, 3, 0xFF}, 2100, 100, 200)
	require.NoError(t, err)

	testMarshaling(t, tx, signer)
}

func TestPermissionMarshaling(t *testing.T) {
	_, pv := crypto.GenerateKey(nil)
	pk, _ := crypto.GenerateKey(nil)
	signer := crypto.NewAccountSigner(pv)
	modifier := signer.Address()
	modified := pk.AccountAddress()
	tx, err := tx.NewPermissionsTx(modifier, modified, permission.Call, true, 1, 100)
	require.NoError(t, err)

	testMarshaling(t, tx, signer)
}

func TestBondMarshaling(t *testing.T) {
	_, pv := crypto.GenerateKey(nil)
	pk, _ := crypto.GenerateKey(nil)
	signer := crypto.NewAccountSigner(pv)
	from := signer.Address()
	tx, err := tx.NewBondTx(from, pk, 9999, 1, 100)
	require.NoError(t, err)

	testMarshaling(t, tx, signer)
}

func TestUnbondMarshaling(t *testing.T) {
	_, pv := crypto.GenerateKey(nil)
	pk, _ := crypto.GenerateKey(nil)
	signer := crypto.NewValidatorSigner(pv)
	from := pv.PublicKey().ValidatorAddress()
	to := pk.AccountAddress()
	tx, err := tx.NewUnbondTx(from, to, 9999, 1, 100)
	require.NoError(t, err)

	testMarshaling(t, tx, signer)
}

func TestSortitionMarshaling(t *testing.T) {
	_, pv := crypto.GenerateKey(nil)
	signer := crypto.NewValidatorSigner(pv)
	val := pv.PublicKey().ValidatorAddress()
	tx, err := tx.NewSortitionTx(val, 1, 555, 1, 100, []byte{1, 2, 3})
	require.NoError(t, err)

	testMarshaling(t, tx, signer)
}

func testMarshaling(t *testing.T, tx tx.Tx, signer crypto.Signer) {
	env1 := Enclose("test-chain", tx)
	var bs []byte

	/// test marshaling without signature
	bs, err := env1.Encode()
	require.NoError(t, err)
	env2 := new(Envelope)
	err = env2.Decode(bs)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, env1, env2)

	bs, err = json.Marshal(env1)
	require.NoError(t, err)
	env3 := new(Envelope)
	err = json.Unmarshal(bs, env3)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, env1, env3)

	/// Now sign it and test marshaling with signature
	err = env1.Sign(signer)
	require.NoError(t, err)
	sb, _ := env1.SignBytes()
	js, _ := json.Marshal(env1)
	fmt.Println("Sign bytes: " + string(sb))
	fmt.Println("Tx JSON: " + string(js))

	bs, err = env1.Encode()
	require.NoError(t, err)
	env4 := new(Envelope)
	err = env4.Decode(bs)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, env1, env4)

	bs, err = json.Marshal(env1)
	require.NoError(t, err)
	env5 := new(Envelope)
	err = json.Unmarshal(bs, env5)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, env1, env5)

	require.NoError(t, env1.Verify(), "Error verifying tx: %s", debug.Stack())
}

func TestSignature(t *testing.T) {
	pubKey1, privKey1 := crypto.GenerateKey(nil)
	pubKey2, privKey2 := crypto.GenerateKey(nil)
	pubKey3, privKey3 := crypto.GenerateKey(nil)
	pubKey4, privKey4 := crypto.GenerateKey(nil)

	signer1 := crypto.NewAccountSigner(privKey1)
	signer2 := crypto.NewAccountSigner(privKey2)
	signer3 := crypto.NewAccountSigner(privKey3)
	signer4 := crypto.NewAccountSigner(privKey4)

	tx1, _ := tx.EmptySendTx()
	tx1.AddReceiver(crypto.GlobalAddress, 1)
	tx1.AddSender(pubKey1.AccountAddress(), 1, 1)
	tx1.AddSender(pubKey2.AccountAddress(), 1, 1)
	tx1.AddSender(pubKey3.AccountAddress(), 1, 1)

	env1 := Enclose("test-chain", tx1)

	err := env1.Sign(signer1, signer2)
	assert.Error(t, err)

	// Should fail, one signature is missed
	err = env1.Verify()
	require.Error(t, err)

	err = env1.Sign(signer1, signer2, signer3)
	require.NoError(t, err)
	err = env1.Verify()
	require.NoError(t, err)

	// invalid public key, should fail
	env2 := Enclose("test-chain", tx1)
	env2.Sign(signer1, signer2, signer3)
	env2.Signatories[0].PublicKey = pubKey2
	err = env2.Verify()
	require.Error(t, err)

	// extra signature, should fail
	env3 := Enclose("test-chain", tx1)
	env3.Sign(signer1, signer2, signer3)
	bs, _ := env3.SignBytes()
	sig, _ := signer4.Sign(bs)
	env3.Signatories = append(env3.Signatories, crypto.Signatory{
		PublicKey: pubKey4,
		Signature: sig})
	err = env3.Verify()
	require.Error(t, err)

	// invalid signature, should fail
	env4 := Enclose("test-chain", tx1)
	env4.Sign(signer1, signer2, signer3)
	env4.Signatories[0].Signature = sig
	err = env2.Verify()
	require.Error(t, err)

	// invalid signBytes, should fail
	env5 := Enclose("test-chain", tx1)
	env5.Sign(signer1, signer2, signer3)
	env5.ChainID = "test-chain-bad"
	err = env5.Verify()
	require.Error(t, err)

	// duplicated sender, should pass
	tx2, _ := tx.EmptySendTx()
	tx2.AddReceiver(crypto.GlobalAddress, 1)
	tx2.AddSender(pubKey1.AccountAddress(), 1, 1)
	tx2.AddSender(pubKey1.AccountAddress(), 1, 1) /// duplicated sender
	tx2.AddSender(pubKey3.AccountAddress(), 1, 1)

	env6 := Enclose("test-chain", tx2)
	// sender1 signs tx twice. It's ridicules but correct.
	env6.Sign(signer1, signer3)
	err = env6.Verify()
	require.NoError(t, err)
}
