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
	bs, err := hex.DecodeString(text)
	if err != nil {
		return Signature{}, e.Errorf(e.ErrInvalidSignature, "%v", err.Error())
	}

	return SignatureFromRawBytes(bs)
}

func SignatureFromRawBytes(bs []byte) (Signature, error) {
	sig := Signature{
		data: signatureData{
			Signature: bs,
		},
	}

	if err := sig.EnsureValid(); err != nil {
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
	return sig.RawBytes(), nil
}

func (sig *Signature) UnmarshalAmino(bs []byte) error {
	s, err := SignatureFromRawBytes(bs)
	if err != nil {
		return err
	}

	*sig = s
	return nil
}

func (sig Signature) MarshalText() ([]byte, error) {
	return []byte(sig.String()), nil
}

func (sig *Signature) UnmarshalText(text []byte) error {
	s, err := SignatureFromString(string(text))
	if err != nil {
		return err
	}

	*sig = s
	return nil
}

/// ----------
/// ATTRIBUTES

func (sig *Signature) EnsureValid() error {
	bs := sig.RawBytes()
	if len(bs) != ed25519.SignatureSize {
		return e.Errorf(e.ErrInvalidSignature, "Signature should be %v bytes but it is %v bytes", ed25519.SignatureSize, len(bs))
	}

	return nil
}
