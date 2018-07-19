package crypto

import (
	"encoding/hex"

	"github.com/gallactic/gallactic/errors"
	"golang.org/x/crypto/ed25519"
)

type Signature struct {
	data signatureData
}

type signatureData struct {
	Signature []byte `json:"signature"`
}

/// ------------
/// CONSTRUCTORS
func SignatureFromString(text string) (Signature, error) {
	var sig Signature
	if err := sig.UnmarshalText([]byte(text)); err != nil {
		return Signature{}, err
	}

	return sig, nil
}

func SignatureFromRawBytes(bs []byte) (Signature, error) {
	var sig Signature
	if err := sig.UnmarshalAmino(bs); err != nil {
		return Signature{}, err
	}

	return sig, nil
}

/// -------
/// CASTING

func (sig Signature) RawBytes() []byte {
	return sig.data.Signature
}

func (sig Signature) String() string {
	return hex.EncodeToString(sig.RawBytes())
}

/// ----------
/// MARSHALING

func (sig Signature) MarshalAmino() ([]byte, error) {
	return sig.data.Signature, nil
}

func (sig *Signature) UnmarshalAmino(bs []byte) error {
	sig.data.Signature = bs
	if err := sig.EnsureValid(); err != nil {
		return err
	}

	return nil
}

func (sig Signature) MarshalText() ([]byte, error) {
	return []byte(sig.String()), nil
}

func (sig *Signature) UnmarshalText(text []byte) error {
	bs, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}

	return sig.UnmarshalAmino(bs)
}

/// ----------
/// ATTRIBUTES

func (sig *Signature) EnsureValid() error {
	bs := sig.RawBytes()
	if len(sig.data.Signature) != ed25519.SignatureSize {
		return e.Errorf(e.ErrInvalidSignature, "Signature should be %v bytes but it is %v bytes", ed25519.SignatureSize, len(bs))
	}

	return nil
}
