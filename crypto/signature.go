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

// Currently this is a stub that reads the raw bytes returned by key_client and returns
// an ed25519 signature.
func SignatureFromRawBytes(bs []byte) (Signature, error) {
	sig := Signature{
		data: signatureData{
			Signature: bs,
		},
	}

	if !sig.IsValid() {
		return Signature{}, e.Error(e.ErrInvalidSignature)
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

func (sig Signature) MarshalText() ([]byte, error) {
	str := sig.String()
	return []byte(str), nil
}

func (sig *Signature) UnmarshalText(bs []byte) error {
	str := string(bs)
	bs, err := hex.DecodeString(str)
	if err != nil {
		return err
	}

	s, err := SignatureFromRawBytes(bs)
	if err != nil {
		return err
	}

	*sig = s
	return nil
}

/// ----------
/// ATTRIBUTES

func (sig Signature) IsValid() bool {
	if len(sig.data.Signature) != ed25519.SignatureSize {
		return false
	}

	return true
}
