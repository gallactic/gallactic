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
	tx, err := tx.NewCallTx(caller, crypto.Address{}, 1, []byte{1, 2, 3}, 2100, 100, 200)
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
	ac := NewAminoCodec()
	jc := NewJSONCodec()

	env1 := Enclose("test-chain", tx)
	var bs []byte

	/// test marshaling without signature
	bs, err := ac.EncodeTx(env1)
	require.NoError(t, err)
	env2, err := ac.DecodeTx(bs)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, env1, env2)

	bs, err = jc.EncodeTx(env1)
	require.NoError(t, err)
	env3, err := jc.DecodeTx(bs)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, env1, env3)

	/// Now sign it and test marshaling with signature
	err = env1.Sign(signer)
	require.NoError(t, err)
	sb, _ := env1.SignBytes()
	js, _ := json.Marshal(env1)
	fmt.Println("Sign bytes: " + string(sb))
	fmt.Println("Tx JSON: " + string(js))

	bs, err = ac.EncodeTx(env1)
	require.NoError(t, err)
	env4, err := ac.DecodeTx(bs)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, env1, env4)

	bs, err = jc.EncodeTx(env1)
	require.NoError(t, err)
	env5, err := jc.DecodeTx(bs)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, env1, env5)

	require.NoError(t, env1.Verify(), "Error verifying tx: %s", debug.Stack())
}

func TestSignature(t *testing.T) {
	_, privKey1 := crypto.GenerateKey(nil)
	_, privKey2 := crypto.GenerateKey(nil)
	_, privKey3 := crypto.GenerateKey(nil)

	pubKey1 := privKey1.PublicKey()
	pubKey2 := privKey2.PublicKey()
	pubKey3 := privKey3.PublicKey()

	signer1 := crypto.NewAccountSigner(privKey1)
	signer2 := crypto.NewAccountSigner(privKey2)
	signer3 := crypto.NewAccountSigner(privKey3)

	tx, _ := tx.EmptySendTx()
	tx.AddReceiver(crypto.GlobalAddress, 1)
	tx.AddSender(pubKey1.AccountAddress(), 1, 1)
	tx.AddSender(pubKey2.AccountAddress(), 1, 1)
	tx.AddSender(pubKey3.AccountAddress(), 1, 1)

	txEnv := Enclose("test-chain", tx)

	err := txEnv.Sign(signer1, signer2)
	assert.Error(t, err)

	// Should fail, one signature is missed
	err = txEnv.Verify()
	require.Error(t, err)

	err = txEnv.Sign(signer1, signer2, signer3)
	require.NoError(t, err)

	err = txEnv.Verify()
	require.NoError(t, err)

	/// TODO: Add more tests here
}
