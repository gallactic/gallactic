package txs

import (
	"encoding/json"
	"fmt"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs/tx"
	"golang.org/x/crypto/ripemd160"
)

type Codec interface {
	Encoder
	Decoder
}

type Encoder interface {
	EncodeTx(envelope *Envelope) ([]byte, error)
}

type Decoder interface {
	DecodeTx(txBytes []byte) (*Envelope, error)
}

// Envelope contains both the signable Tx and the signatures for each input (in signatories)
type Envelope struct {
	ChainID     string      `json:"chainId"`
	Type        tx.Type     `json:"type"`
	Tx          tx.Tx       `json:"tx"`
	Signatories []Signatory `json:"signatories,omitempty"`
	/// TODO: save tx hash to prevent calling json marshal everytime
}

// Signatory contains signature and PublicKey to identify the signer
type Signatory struct {
	PublicKey crypto.PublicKey `json:"publicKey"`
	Signature crypto.Signature `json:"signature"`
}

func Enclose(chainId string, tx tx.Tx) *Envelope {
	return &Envelope{
		ChainID: chainId,
		Type:    tx.Type(),
		Tx:      tx,
	}
}

func (env *Envelope) UnmarshalJSON(data []byte) error {
	type _envelope struct {
		ChainID     string          `json:"chainId"`
		Type        tx.Type         `json:"type"`
		Tx          json.RawMessage `json:"tx"`
		Signatories []Signatory     `json:"signatories,omitempty"`
	}

	w := new(_envelope)
	err := json.Unmarshal(data, w)
	if err != nil {
		return err
	}
	env.ChainID = w.ChainID
	env.Type = w.Type
	env.Signatories = w.Signatories
	// Now we know the Type we can deserialise tx
	env.Tx = tx.New(w.Type)
	return json.Unmarshal(w.Tx, env.Tx)
}

// SignBytes produces the canonical SignBytes for a Tx
func (env Envelope) SignBytes() ([]byte, error) {
	env.Signatories = nil
	bs, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("could not generate canonical SignBytes for tx %v: %v", env, err)
	}
	return bs, nil
}

func (env *Envelope) Hash() []byte {
	hasher := ripemd160.New()
	bytes, err := env.SignBytes()
	if err != nil {
		return nil
	}
	hasher.Write(bytes)
	hash := hasher.Sum(nil)
	return hash
}

func (env *Envelope) String() string {
	return fmt.Sprintf("Envelop{TxHash: %X; Tx: %v}", env.Hash(), env.Tx)
}

// Verify verifies the validity of the Signatories' Signatures in the Envelope. The Signatories must
// appear in the same order as the inputs as returned by Tx.GetInputs().
func (env *Envelope) Verify() error {
	if len(env.Signatories) == 0 {
		return e.Errorf(e.ErrInvalidSignature, "transaction envelope contains no (successfully unmarshalled) signatories")
	}

	errPrefix := fmt.Sprintf("could not verify transaction %X", env.Hash())
	inputs := env.Tx.Signers()
	if len(inputs) != len(env.Signatories) {
		return e.Errorf(e.ErrInvalidSignature, "%s: number of inputs (= %v) should equal number of signatories (= %v)",
			errPrefix, len(inputs), len(env.Signatories))
	}
	signBytes, err := env.SignBytes()
	if err != nil {
		return e.Errorf(e.ErrInvalidSignature, "%s: could not generate SignBytes: %v", errPrefix, err)
	}
	// Expect order to match (we could build lookup but we want Verify to be quicker than Sign which does order sigs)
	for i, s := range env.Signatories {
		if !inputs[i].Address.Verify(s.PublicKey) {
			return e.Errorf(e.ErrInvalidSignature, "%s: address %v can not be verified",
				errPrefix, inputs[i].Address)
		}

		if !s.PublicKey.Verify(signBytes, s.Signature) {
			return e.Errorf(e.ErrInvalidSignature, "%s: invalid signature in signatory %v ", errPrefix, inputs[i].Address)
		}
	}

	return nil
}

// Sign the Tx by adding Signatories containing the signatures for each Input.
// Signder for each input must be provided (in any order).
func (env *Envelope) Sign(signers ...crypto.Signer) error {
	// Clear any existing
	env.Signatories = env.Signatories[:0]
	signBytes, err := env.SignBytes()
	if err != nil {
		return err
	}

	signerMap := make(map[crypto.Address]crypto.Signer)
	for _, signer := range signers {
		signerMap[signer.Address()] = signer
	}
	// Sign in order of inputs
	for _, input := range env.Tx.Signers() {
		signer, ok := signerMap[input.Address]
		if !ok {
			return e.Errorf(e.ErrInvalidSignature, "Account to sign %v not passed to Sign", input)
		}
		signature, err := signer.Sign(signBytes)
		if err != nil {
			return err
		}
		publicKey := signer.PublicKey()
		env.Signatories = append(env.Signatories, Signatory{
			PublicKey: publicKey,
			Signature: signature,
		})
	}
	return nil
}

// BroadcastTx or Transaction receipt
type Receipt struct {
	TxHash []byte
}

// Generate a transaction Receipt containing the Tx hash.
// Returned by ABCI methods.
func (env *Envelope) GenerateReceipt() *Receipt {
	receipt := &Receipt{
		TxHash: env.Hash(),
	}

	return receipt
}
