package txs

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/gallactic/gallactic/crypto"
	e "github.com/gallactic/gallactic/errors"
	"github.com/gallactic/gallactic/txs/tx"
	amino "github.com/tendermint/go-amino"
)

// Envelope contains both the signable Tx and the signatures for each input (in signatories)
type Envelope struct {
	ChainID     string             `json:"chainId"`
	Type        tx.Type            `json:"type"`
	Tx          tx.Tx              `json:"tx"`
	Signatories []crypto.Signatory `json:"signatories,omitempty"`
	hash        []byte
}

func Enclose(chainId string, tx tx.Tx) *Envelope {
	return &Envelope{
		ChainID: chainId,
		Type:    tx.Type(),
		Tx:      tx,
	}
}

// signBytes produces the canonical SignBytes for a Tx
func (env Envelope) signBytes() ([]byte, error) {
	env.Signatories = nil
	bs, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("could not generate canonical SignBytes for tx %v: %v", env, err)
	}

	return bs, nil
}

func (env *Envelope) Hash() []byte {
	if env.hash != nil {
		return env.hash
	}
	if env.Signatories == nil {
		return nil
	}

	bs, err := env.Encode()
	if err != nil {
		return nil
	}
	h := sha256.Sum256(bs) // Make sure that the Gallactic.Tx.Hash is same as Tendermint.Tx.Hash
	env.hash = h[:]

	return env.hash
}

func (env *Envelope) String() string {
	//s, _ := json.Marshal(env.Tx)
	//return fmt.Sprintf("Envelop{TxHash: %X; Tx: %s}", env.Hash(), s)
	return fmt.Sprintf("Envelop{TxHash: %X; Tx: %v}", env.Hash(), env.Tx)
}

// Verify verifies the validity of the Signatories' Signatures in the Envelope. The Signatories must
// appear in the same order as the inputs as returned by Tx.GetInputs().
func (env *Envelope) Verify() error {
	if len(env.Signatories) == 0 {
		return e.Errorf(e.ErrInvalidSignature, "Transaction envelope contains no signatories")
	}

	inputs := env.Tx.Signers()
	if len(inputs) != len(env.Signatories) {
		return e.Errorf(e.ErrInvalidSignature, "Number of inputs (= %v) should equal number of signatories (= %v)",
			len(inputs), len(env.Signatories))
	}
	signBytes, err := env.signBytes()
	if err != nil {
		return e.Errorf(e.ErrInvalidSignature, "Could not generate SignBytes: %v", err)
	}
	// Expect order to match (we could build lookup but we want Verify to be quicker than Sign which does order sigs)
	for i, s := range env.Signatories {
		if !inputs[i].Address.Verify(s.PublicKey) {
			return e.Errorf(e.ErrInvalidSignature, "Address %v can not be verified", inputs[i].Address)
		}

		if !s.PublicKey.Verify(signBytes, s.Signature) {
			return e.Errorf(e.ErrInvalidSignature, "Invalid signature in signatory %v", inputs[i].Address)
		}
	}

	return nil
}

// Sign the Tx by adding Signatories containing the signatures for each Input.
// Signder for each input must be provided (in any order).
func (env *Envelope) Sign(signers ...crypto.Signer) error {
	// Clear any existing
	env.Signatories = env.Signatories[:0]
	signBytes, err := env.signBytes()
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
		env.Signatories = append(env.Signatories, crypto.Signatory{
			PublicKey: publicKey,
			Signature: signature,
		})
	}
	return nil
}

// Generate a transaction Receipt containing the Tx hash.
// Returned by ABCI methods.
func (env *Envelope) GenerateReceipt() *Receipt {
	receipt := &Receipt{
		Type: env.Type,
		Hash: env.Hash(),
	}

	return receipt
}

// Marshaling/Unmarshaling methods
func (env *Envelope) UnmarshalJSON(data []byte) error {
	type _envelope struct {
		ChainID     string             `json:"chainId"`
		Type        tx.Type            `json:"type"`
		Tx          json.RawMessage    `json:"tx"`
		Signatories []crypto.Signatory `json:"signatories,omitempty"`
	}

	w := new(_envelope)
	err := json.Unmarshal(data, w)
	if err != nil {
		return err
	}
	env.ChainID = w.ChainID
	env.Type = w.Type
	env.Signatories = w.Signatories
	// Now we know the Type we can de-serialize tx
	env.Tx = tx.New(w.Type)
	return json.Unmarshal(w.Tx, env.Tx)
}

func NewAminoCodec() *amino.Codec {
	cdc := amino.NewCodec()
	cdc.RegisterInterface((*tx.Tx)(nil), nil)
	registerTx(cdc, &tx.SendTx{})
	registerTx(cdc, &tx.CallTx{})
	registerTx(cdc, &tx.BondTx{})
	registerTx(cdc, &tx.UnbondTx{})
	registerTx(cdc, &tx.PermissionsTx{})
	registerTx(cdc, &tx.SortitionTx{})
	return cdc
}

func registerTx(cdc *amino.Codec, tx tx.Tx) {
	cdc.RegisterConcrete(tx, fmt.Sprintf("gallactic/txs/tx/%v", tx.Type()), nil)
}

var cdc = NewAminoCodec()

func (env *Envelope) Encode() ([]byte, error) {
	return cdc.MarshalBinaryLengthPrefixed(env)
}

func (env *Envelope) Decode(bs []byte) error {
	return cdc.UnmarshalBinaryLengthPrefixed(bs, env)
}

func (env *Envelope) Unmarshal(bs []byte) error {
	return env.Decode(bs)
}

func (env *Envelope) Marshal() ([]byte, error) {
	return env.Encode()
}

func (env *Envelope) MarshalTo(data []byte) (int, error) {
	bs, err := env.Encode()
	if err != nil {
		return -1, err
	}
	return copy(data, bs), nil
}

func (env *Envelope) Size() int {
	bs, _ := env.Encode()
	return len(bs)
}

// For Recipt
func (r *Receipt) Encode() ([]byte, error) {
	return cdc.MarshalBinaryLengthPrefixed(&r)
}

func (r *Receipt) Decode(bs []byte) error {
	return cdc.UnmarshalBinaryLengthPrefixed(bs, &r)
}

func (r *Receipt) Unmarshal(bs []byte) error {
	return r.Decode(bs)
}

func (r *Receipt) Marshal() ([]byte, error) {
	return r.Encode()
}

func (r *Receipt) MarshalTo(data []byte) (int, error) {
	bs, err := r.Encode()
	if err != nil {
		return -1, err
	}
	return copy(data, bs), nil
}

func (r *Receipt) Size() int {
	bs, _ := r.Encode()
	return len(bs)
}
