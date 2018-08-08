package proposal

import (
	"testing"
	"time"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
)

func TestVerifyProposal(t *testing.T) {
	/*Initialize the proposal */
	var p Proposal

	/* creating the Private Key and Public Key */
	pubKey1, privKey1 := crypto.GenerateKey(nil)
	pubKey2, privKey2 := crypto.GenerateKey(nil)
	pubKey3, privKey3 := crypto.GenerateKey(nil)
	pubKey4, privKey4 := crypto.GenerateKey(nil)
	pubKey5, privKey5 := crypto.GenerateKey(nil)

	/* create the address from public_key */
	address1 := pubKey1.AccountAddress()
	address2 := pubKey2.AccountAddress()

	/* create  accounts for genesis */
	acc1, err1 := account.NewAccount(address1)
	assert.NoError(t, err1)
	acc2, err2 := account.NewAccount(address2)
	assert.NoError(t, err2)

	/* create validator account for genesis  using publicKey3 */
	validtorAcc, valerror := validator.NewValidator(pubKey3, 100)
	assert.NoError(t, valerror)
	validtorAcc1, valerror1 := validator.NewValidator(pubKey4, 100)
	assert.NoError(t, valerror1)

	/* accounts list */
	accs := []*account.Account{acc1, acc2}

	//global account
	gAcc, _ := account.NewAccount(crypto.GlobalAddress)

	/* create list of validator account*/
	vals := []*validator.Validator{validtorAcc, validtorAcc1}

	/* create genesis  */
	gen := MakeGenesis("Genesis-001", time.Now().Truncate(0), gAcc, accs, nil, vals)

	/* Pass the genesis to proposal */
	p.Genesis = gen
	//Sign Hash for signing the message for signature
	signHash := gen.Hash()

	/* creating siginer and signature  of accounts */
	signer1 := crypto.NewAccountSigner(privKey1)
	signature1, signerror1 := signer1.Sign(signHash)
	assert.NoError(t, signerror1)

	signer2 := crypto.NewAccountSigner(privKey2)
	signature2, signerror2 := signer2.Sign(signHash)
	assert.NoError(t, signerror2)

	signer3 := crypto.NewValidatorSigner(privKey3)
	signature3, signerror3 := signer3.Sign(signHash)
	assert.NoError(t, signerror3)

	signer4 := crypto.NewValidatorSigner(privKey4)
	signature4, signerror4 := signer4.Sign(signHash)
	assert.NoError(t, signerror4)

	signer5 := crypto.NewAccountSigner(privKey5)
	signature5, signerror5 := signer5.Sign(signHash)
	assert.NoError(t, signerror5)

	/* create array element of public key and signature*/
	s1 := signatory{PublicKey: pubKey1, Signature: signature1}
	s2 := signatory{PublicKey: pubKey2, Signature: signature2}
	s3 := signatory{PublicKey: pubKey3, Signature: signature3}
	s4 := signatory{PublicKey: pubKey4, Signature: signature4}
	s5 := signatory{PublicKey: pubKey5, Signature: signature5}

	/* append the array elemet to propsal struct  Signatories*/

	var sigs []signatory
	sigs = append(sigs, s1)
	sigs = append(sigs, s2)
	sigs = append(sigs, s3)
	sigs = append(sigs, s4)

	p.Signatories = sigs

	/* verify the proposal*/
	err := p.Verify()
	assert.NoError(t, err)

	// change genesis hash to check signature, test should fail
	var p2 Proposal
	gen2 := MakeGenesis("Genesis-002", time.Now().Truncate(0), gAcc, accs, nil, vals)
	p2.Genesis = gen2
	p2.Signatories = sigs

	err = p2.Verify()
	assert.Error(t, err)

	// Removing one signature, test should fail
	var p3 Proposal
	p3.Genesis = gen
	p3.Signatories = append(sigs[:0], sigs[1:]...)

	err = p3.Verify()
	assert.Error(t, err)

	// Adding one signature, test should fail
	var p4 Proposal
	p4.Genesis = gen
	p4.Signatories = sigs
	p4.Signatories = append(p.Signatories, s5)

	err = p4.Verify()
	assert.Error(t, err)

	// change publicKey
	var p5 Proposal
	p5.Genesis = gen
	p5.Signatories = make([]signatory, len(sigs))
	copy(p5.Signatories, sigs)
	p5.Signatories[0].PublicKey = sigs[1].PublicKey

	err = p5.Verify()
	assert.Error(t, err)

	// change signature
	var p6 Proposal
	p6.Genesis = gen
	p6.Signatories = make([]signatory, len(sigs))
	copy(p6.Signatories, sigs)
	p6.Signatories[0].Signature = sigs[1].Signature

	err = p6.Verify()
	assert.Error(t, err)
}
